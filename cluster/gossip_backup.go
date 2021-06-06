// +build never

package cluster

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
	statemachine "github.com/jeremyhahn/go-cropdroid/cluster/state"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/pborman/uuid"
)

const METADATA_SIZE_LIMIT = 512 // bytes

type update struct {
	Action string // add, del
	Data   map[string]string
}

type GossipCluster interface {
	GetHashring() *Consistent
	GetHealthScore() int
	GetMemberCount() int
	GetSystemRaft() RaftCluster
	Join()
	Provision(config config.FarmConfig) error
	Run()
	Shutdown() error
	Leave() error
}

type Gossip struct {
	clock      LamportClock
	mutex      *sync.RWMutex
	params     *ClusterParams
	port       int
	memberlist *memberlist.Memberlist
	broadcasts *memberlist.TransmitLimitedQueue
	hashring   *Consistent                 // Raft clusters on the network
	vnodes     int                         // Number of consistent hash ring virtual nodes
	node       *memberlist.Node            // local node
	nodemeta   *nodeMeta                   // local node metadata (unserialized)
	nodes      map[string]*memberlist.Node // All nodes in the cluster (by address)
	//nodesIndex map[nodeID]*memberlist.Node // All nodes in the cluster (by nodeid)
	raft  RaftCluster         // control-plane / system raft
	rafts map[uint64][]uint64 // data-plane; map[raft_cluster_id][]node_id
	//nodePool   map[uint64]*memberlist.Node // nodes waiting to join a cluster
	//nextRaftID uint64                    // next raft cluster id waiting to be allocated
	items map[string]string
	GossipCluster
}

/*
type memberState struct {
	Member
	statusLTime LamportTime // lamport clock time of last received message
	leaveTime   time.Time   // wall clock time of leave
}*/

type nodeMeta struct {
	Provider string `json:"p"`
	Region   string `json:"r"`
	//ClusterID uint64 `json:"cid"`
	NodeID    uint64 `json:"id"`
	Cpu       int    `json:"cpu"`
	Memory    int    `json:"mem"`
	DiskAvail int    `json:"da"`
	DiskUsed  int    `json:"du"`
	Load      int    `json:"load"`
}

// NewGossipCluster creates a new GossipCluster implementation on the heap
// and returns a pointer. This operation blocks until the minimum number of
// nodes required (3) to create an initial Raft cluster have found each other.
func NewGossipCluster(params *ClusterParams, hashring *Consistent) GossipCluster {

	/* Create the initial memberlist from a safe configuration.
	   Please reference the godoc for other default config types.
	   http://godoc.org/github.com/hashicorp/memberlist#Config
	   https://managementfromscratch.wordpress.com/2016/04/01/introduction-to-gossip/
	*/

	var nodeName string
	hostname, _ := os.Hostname()

	cluster := &Gossip{
		mutex:      &sync.RWMutex{},
		params:     params,
		port:       params.gossipPort,
		rafts:      make(map[uint64][]uint64, 0),
		broadcasts: &memberlist.TransmitLimitedQueue{},
		hashring:   hashring,
		nodes:      make(map[string]*memberlist.Node, 0),
		//nodePool:   make(map[uint64]*memberlist.Node, 0),
		items: make(map[string]string, 0)}

	cluster.clock.Increment() // start with 1

	if params.bootstrap > 0 {
		if params.peers[0] == "" {
			cluster.params.peers = make([]string, 0)
			params.nodeID = 0
			nodeName = fmt.Sprintf("%s-%d-%d", hostname, params.clusterID, params.nodeID)
		} else {
			params.nodeID = uint64(len(cluster.params.peers) + 1)
			nodeName = fmt.Sprintf("%s-%d-%d", hostname, params.clusterID, params.nodeID)
		}
	} else {
		// to generate monotonic node id, need lamport timestamps;
		// using control plane leader for now
		nodeName = fmt.Sprintf("%s-%d-%s", hostname, params.clusterID, uuid.NewUUID().String())
	}

	params.logger.Debugf("Gossip peers: %s", cluster.params.peers)

	conf := memberlist.DefaultLocalConfig()
	conf.Events = cluster
	conf.Delegate = cluster
	conf.Name = nodeName
	//conf.BindAddr = params.listen
	conf.BindPort = cluster.port

	params.logger.Debugf("Local gossip port: %d", conf.BindPort)

	list, err := memberlist.Create(conf)
	if err != nil {
		params.logger.Errorf("Failed to initialize gossip cluster: %s", err)
	}

	cluster.memberlist = list
	return cluster
}

// Join the gossip network. This call blocks until enough nodes are available to build a minimal 3 node raft quorum.
func (cluster *Gossip) Join() {

	var contacted int
	var err error
	//nodeID := uint64(len(cluster.params.peers) + 1)
	var nodeID uint64

	if cluster.params.bootstrap > 0 {
		nodeID = uint64(len(cluster.params.peers) + 1)
	}

	joinCluster := func() (int, error) {
		return cluster.memberlist.Join(cluster.params.peers)
	}

	if len(cluster.params.peers) > 0 {
		contacted, err = joinCluster()
		if err != nil {
			cluster.params.logger.Errorf("Failed to join cluster: %s", err)
		}
	}

	members := cluster.memberlist.Members()
	for len(members) < cluster.params.bootstrap {

		cluster.params.logger.Debugf("Waiting for enough nodes to build initial Raft quorum. %d of %d members have joined. Contacted: %d",
			cluster.memberlist.NumMembers(), cluster.params.bootstrap, contacted)

		time.Sleep(1 * time.Second)
		if len(cluster.params.peers) > 0 {
			contacted, err = joinCluster()
			if err != nil {
				cluster.params.logger.Errorf("Failed to join cluster: %s", err)
			}
		}
		members = cluster.memberlist.Members()
	}

	cluster.node = cluster.memberlist.LocalNode()
	//localNodeAddress := cluster.node.Address()
	numMembers := len(members)

	if cluster.params.bootstrap > 0 {

		raftring := NewHashring(1)

		/*
			for i, member := range members {
				memberAddress := member.Address()
				cluster.params.logger.Debugf("Cluster member: %s", memberAddress)
				nodeIDs[i] = uint64(i + 1)
				cluster.nodes[nodeIDs[i]] = member
			}
			cluster.rafts[raftID] = nodeIDs
		*/

		params := cluster.params
		//params.peers = []string{members[0].Address(), members[1].Address(), members[2].Address()}
		params.nodeID = nodeID
		cluster.raft = NewRaftCluster(params, raftring)

		for _, member := range members {
			cluster.hashring.Add(member.Address())
		}

		nodeIDs := make([]uint64, 0)
		clusterInfo := cluster.raft.GetClusterInfo(params.clusterID)

		for nid, raftAddress := range clusterInfo.Nodes {
			cluster.params.logger.Debugf("comparing nid=%d ro raftAddress=%s", nid, raftAddress)
			//cluster.nodes[nid] = member[nid-1]
			nodeIDs = append(nodeIDs, nid)
			break
		}

		/*
			nodeIDs = make([]uint64, 0)
			// Organize gossip nodes according to raft consensus
			clusterInfo := cluster.raft.GetClusterInfo(params.clusterID)
			for _, member := range members {
				for nid, raftAddress := range clusterInfo.Nodes {
					cluster.params.logger.Debugf("comparing nid=%d ro raftAddress=%s", nid, raftAddress)
					if member.Address() == raftAddress() {
						cluster.nodes[nid] = member
						nodeIDs = append(nodeIDs, nid)
					}
				}
			}
			//cluster.hashring.Add(cluster.raft.GetCluster().RaftAddress())
			cluster.rafts[params.clusterID] = clusterInfo.Nodes
			cluster.hashring.Inc(cluster.nodes[leaderID].Address())

			for raftID, nid := range cluster.rafts {
				cluster.params.logger.Debugf("raftID=%d, nodeID=%d, leaderID=%d", raftID, nid, leaderID)
			}
		*/
	}

	cluster.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return numMembers
		},
		RetransmitMult: 3,
	}

	cluster.nodemeta = &nodeMeta{
		Provider: cluster.params.provider,
		Region:   cluster.params.region,
		//ClusterID: uint64(cluster.params.clusterID),
		NodeID:    nodeID,
		Cpu:       0,
		Memory:    0,
		DiskAvail: 0,
		DiskUsed:  0}

	cluster.params.logger.Debugf("Joined the network as %s:%d", cluster.node.Addr, cluster.node.Port)

	cluster.node.Meta = cluster.NodeMeta(METADATA_SIZE_LIMIT)

	//cluster.nodePool[nodeID] = cluster.node
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (cluster *Gossip) NotifyJoin(node *memberlist.Node) {

	fmt.Println("A node has joined: " + node.String())

	// Join events processed ONLY by bootstrapped control-plane leader
	if cluster.raft != nil && cluster.raft.IsLeader(CONTROL_PLANE_CLUSTER_ID) {
		if cluster.raft.GetNodeCount() < QUORUM_MAX_NODES {

			cluster.params.logger.Error("[Gossip.NotifyJoin] cluster.raft.nodeCount < QUORUM_MAX_NODES")

			if err := cluster.raft.AddNode(cluster.params.clusterID, node.Address(), 0); err != nil {
				cluster.params.logger.Errorf("[Gossip.NotifyJoin] Error: %s", err)
			}
			cluster.hashring.Add(node.Address())
		} else {

			cluster.params.logger.Error("[Gossip.NotifyJoin] cluster.raft.nodeCount >= QUORUM_MAX_NODES")

			if len(cluster.nodes) >= QUORUM_MIN_NODES {
				cluster.params.logger.Errorf("[Gossip.NotifyJoin] TODO: Create new raft cluster using IaaS plugin!")
				cluster.hashring.Add(node.Address())
			}
		}
	} else if cluster.raft == nil && cluster.params.clusterID == CONTROL_PLANE_CLUSTER_ID {
		// New control plane node
		raftring := NewHashring(1)
		params := cluster.params
		params.raft = append(params.raft, node.Address())
		params.nodeID = uint64(len(params.raft))
		cluster.raft = NewRaftCluster(params, raftring)
		leaderID, _, _ := cluster.raft.GetLeaderID(params.clusterID)
		for nodeID, raftAddress := range cluster.raft.GetClusterInfo(params.clusterID).Nodes {
			if nodeID == leaderID {
				cluster.hashring.Add(raftAddress)
				break
			}
		}
		cluster.broadcasts = &memberlist.TransmitLimitedQueue{
			NumNodes: func() int {
				return cluster.memberlist.NumMembers()
			},
			RetransmitMult: 3,
		}
		cluster.nodemeta = &nodeMeta{
			Provider: cluster.params.provider,
			Region:   cluster.params.region,
			//ClusterID: uint64(cluster.params.clusterID),
			NodeID:    params.nodeID,
			Cpu:       0,
			Memory:    0,
			DiskAvail: 0,
			DiskUsed:  0}

		cluster.params.logger.Debugf("Joined the network as %s:%d", cluster.node.Addr, cluster.node.Port)

		cluster.node.Meta = cluster.NodeMeta(METADATA_SIZE_LIMIT)
	}
	if cluster.raft == nil {
		cluster.params.logger.Warning("[Gossip.NotifyJoin] Aborting, cluster.raft is nil...")
	} else {
		cluster.params.logger.Warning("[Gossip.NotifyJoin] Aborting, not the control plane leader...")
	}
}

// memberlist.EventDelegate: https://github.com/hashicorp/memberlist/blob/master/event_delegate.go#L7

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (cluster *Gossip) NotifyLeave(node *memberlist.Node) {
	fmt.Println("A node has left: " + node.String())

	cluster.hashring.Remove(node.Address())

	if cluster.memberlist != nil {
		for _, member := range cluster.memberlist.Members() {
			cluster.params.logger.Debugf("Member: %s %s\n", member.Name, member.Addr)
		}
	}
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (cluster *Gossip) NotifyUpdate(node *memberlist.Node) {
	fmt.Println("A node was updated: " + node.String())
}

// End memberlist.EventDelegate

// memberlist.AliveDelegate: https://godoc.org/github.com/hashicorp/memberlist#AliveDelegate

// NotifyAlive is invoked when a message about a live
// node is received from the network.  Returning a non-nil
// error prevents the node from being considered a peer.
func (cluster *Gossip) NotifyAlive(peer *memberlist.Node) error {
	return nil
}

// memberlist.Delegate: https://github.com/hashicorp/memberlist/blob/master/delegate.go#L6

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
// Returns the node metadata as a serialized byte array
func (cluster *Gossip) NodeMeta(limit int) []byte {
	bytes, err := json.Marshal(cluster.nodemeta)
	if err != nil {
		cluster.params.logger.Error(err)
	}
	return bytes
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed
func (cluster *Gossip) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	switch b[0] {
	case 'd': // data
		var updates []*update
		if err := json.Unmarshal(b[1:], &updates); err != nil {
			return
		}
		cluster.mutex.Lock()
		for _, u := range updates {
			for k, v := range u.Data {
				switch u.Action {
				case "add":
					cluster.items[k] = v
				case "del":
					delete(cluster.items, k)
				}
			}
		}
		cluster.mutex.Unlock()
	}
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit. Care should be taken that this method does not block,
// since doing so would block the entire UDP packet receive loop.
func (cluster *Gossip) GetBroadcasts(overhead, limit int) [][]byte {
	return cluster.broadcasts.GetBroadcasts(overhead, limit)
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (cluster *Gossip) LocalState(join bool) []byte {
	cluster.mutex.RLock()
	m := cluster.items
	cluster.mutex.RUnlock()
	b, _ := json.Marshal(m)
	return b
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (cluster *Gossip) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	if !join {
		return
	}
	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}
	cluster.mutex.Lock()
	for k, v := range m {
		cluster.items[k] = v
	}
	cluster.mutex.Unlock()
}

// End memberlist.Delegate

// Start local API

func (cluster *Gossip) GetSystemRaft() RaftCluster {
	return cluster.raft
}

func (cluster *Gossip) GetHashring() *Consistent {
	return cluster.hashring
}

func (cluster *Gossip) Shutdown() error {
	return cluster.memberlist.Shutdown()
}

// Returns the unserialized node metadata
func (cluster *Gossip) GetNodeMeta() *nodeMeta {
	return cluster.nodemeta
}

func (cluster *Gossip) GetMemberCount() int {
	return cluster.memberlist.NumMembers()
}

func (cluster *Gossip) GetHealthScore() int {
	return cluster.memberlist.GetHealthScore()
}

func (cluster *Gossip) GetMembers() []*memberlist.Node {
	return cluster.memberlist.Members()
}

func (cluster *Gossip) Leave() error {
	return cluster.memberlist.Leave(5 * time.Second)
}

func (cluster *Gossip) Provision(farmConfig config.FarmConfig) error {

	farmProvisionerChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)

	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(fmt.Sprintf("%d-%d", farmConfig.GetOrganizationID(), farmConfig.GetID())))
	clusterID := clusterHash.Sum64()

	params := cluster.params
	params.clusterID = clusterID
	params.dataDir = fmt.Sprintf("%s/%d", cluster.params.dataDir, clusterID)
	sm := statemachine.NewFarmConfigMachine(params.logger, clusterID, farmProvisionerChan, common.DEFAULT_FARM_CONFIG_HISTORY_LENGTH)
	if err := cluster.raft.CreateConfigCluster(clusterID, sm); err != nil {
		return err
	}

	hashring := cluster.raft.GetHashring()
	host, err := hashring.GetLeast(strconv.Itoa(farmConfig.GetID()))
	if err != nil {
		return err
	}
	hashring.Inc(host)

	//cluster.raft.GetNodeHost().RequestLeaderTransfer(clusterID, cluster.nodeTable.nodes[host])

	cluster.params.logger.Errorf("Provisioned new farm: clusterID=%d, leader=%s", clusterID, host)

	/*
		clusters[i] = &ClusterInfo{
			ClusterID:         member.ClusterID,
			NodeID:            member.NodeID,
			Nodes:             member.Nodes,
			ConfigChangeIndex: member.ConfigChangeIndex,
			StateMachineType:  member.StateMachineType,
			IsLeader:          member.IsLeader,
			IsObserver:        member.IsObserver,
			IsWitness:         member.IsWitness,
			Pending:           member.Pending}
	*/
	return nil
}
