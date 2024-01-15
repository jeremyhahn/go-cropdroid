//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jeremyhahn/go-cropdroid/cluster/gossip"
	"github.com/jeremyhahn/go-cropdroid/cluster/util"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config/dao"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
)

const (
	// serfEventBacklogWarning is the threshold at which point log
	// warnings will be emitted indicating a problem when processing serf
	// events.
	serfEventBacklogWarning = 200
)

type update struct {
	Action string // add, del
	Data   map[string]string
}

type GossipNode interface {
	GetHashring() *util.Consistent
	GetHealthScore() int
	GetMemberCount() int
	GetSerfStats() map[string]string
	GetSystemRaft() RaftNode
	Join()
	Provision(*common.ProvisionerParams) error
	Run()
	SetInitializer(dao.Initializer)
	Shutdown() error
	GetParams() *util.ClusterParams
}

type Gossip struct {
	mutex       *sync.RWMutex
	stateMutex  *sync.Mutex
	raftMutex   *sync.Mutex
	params      *util.ClusterParams
	port        int
	serf        *serf.Serf
	state       ClusterState
	eventCh     chan serf.Event
	hashring    *util.Consistent    // Raft clusters on the network
	vnodes      int                 // Number of consistent hash ring virtual nodes
	member      serf.Member         // local node
	nodemeta    *nodeMeta           // local node metadata (unserialized)
	raft        RaftNode            // control-plane / system raft
	rafts       map[uint64][]uint64 // data-plane; map[raft_cluster_id][]raft_node_id
	shutdownCh  chan struct{}
	hosts       map[string]uint64 // map raft address to its node id
	nodeID      uint64            // unique id of this node in the Gossip cluster
	initializer dao.Initializer
	GossipNode
}

type nodeMeta struct {
	Provider string `json:"p"`
	Region   string `json:"r"`
	//ClusterID uint64 `json:"cid"`
	NodeID    uint64 `json:"id"`
	RaftPort  int    `json:"raftPort"`
	Cpu       int    `json:"cpu"`
	Memory    int    `json:"mem"`
	DiskAvail int    `json:"da"`
	DiskUsed  int    `json:"du"`
	Load      int    `json:"load"`
}

// ClusterState is the state of this node in the Gossip cluster.
type ClusterState int

const (
	ClusterBootstrapping ClusterState = iota
	ClusterAssigned
)

func (s ClusterState) String() string {
	switch s {
	case ClusterBootstrapping:
		return "bootstrapping"
	case ClusterAssigned:
		return "assigned"
	default:
		return "unknown"
	}
}

// NewGossipNode creates a new GossipNode implementation on the heap
// and returns a pointer. This operation blocks until the minimum number of
// nodes required (3) to create an initial Raft cluster is formed.
func NewGossipNode(params *util.ClusterParams, hashring *util.Consistent) GossipNode {

	cluster := &Gossip{
		mutex:      &sync.RWMutex{},
		stateMutex: &sync.Mutex{},
		raftMutex:  &sync.Mutex{},
		params:     params,
		state:      ClusterBootstrapping,
		eventCh:    make(chan serf.Event, 1024),
		port:       params.GossipPort,
		rafts:      make(map[uint64][]uint64, 0),
		hashring:   hashring,
		//nodes:      make(map[string]*memberlist.Node, 0),
		//nodePool:   make(map[uint64]*memberlist.Node, 0),
		hosts:      make(map[string]uint64, 0),
		shutdownCh: make(chan struct{})}

	params.Logger.Debugf("Gossip peers: %s", cluster.params.GossipPeers)

	serfConfig := serf.DefaultConfig()
	serfConfig.Init()
	serfConfig.NodeName = params.NodeName
	serfConfig.EventCh = cluster.eventCh
	serfConfig.MemberlistConfig = memberlist.DefaultLocalConfig()
	serfConfig.MemberlistConfig.BindAddr = params.Listen
	serfConfig.MemberlistConfig.BindPort = cluster.port
	serfConfig.MemberlistConfig.AdvertiseAddr = params.Listen
	serfConfig.MemberlistConfig.AdvertisePort = cluster.port
	serfConfig.Tags["raftPort"] = strconv.Itoa(cluster.params.RaftOptions.Port)
	//serfConfig.MemberlistConfig.SecretKey = encryptKey
	//serfConfig.MemberlistConfig.Events = cluster
	//serfConfig.MemberlistConfig.Delegate = cluster

	s, err := serf.Create(serfConfig)
	if err != nil {
		params.Logger.Fatal(err)
	}
	cluster.serf = s
	params.Logger.Debugf("Local gossip port: %d", cluster.port)

	return cluster
}

func (cluster *Gossip) SetInitializer(initializer dao.Initializer) {
	cluster.initializer = initializer
}

// Join the gossip network. This call blocks until enough nodes are available to
// build a minimal 3 node raft quorum.
func (cluster *Gossip) Join() {
	/*
		cluster.node = cluster.memberlist.LocalNode()
		cluster.nodemeta = &nodeMeta{
			Provider: cluster.params.provider,
			Region:   cluster.params.region,
			//ClusterID: uint64(cluster.params.RaftOptions.SystemClusterID),
			NodeID:    nodeID,
			RaftPort:  cluster.params.RaftPort,
			Cpu:       0,
			Memory:    0,
			DiskAvail: 0,
			DiskUsed:  0}

		bytes, err := json.Marshal(cluster.nodemeta)
		if err != nil {
			cluster.params.Logger.Error(err)
		}
		cluster.node.Meta = bytes
	*/

	var contacted int
	var err error

	if len(cluster.params.GossipPeers) > 0 {
		contacted, err = cluster.serf.Join(cluster.params.GossipPeers, true) // true = dont replay events prior to join
		if err != nil {
			cluster.params.Logger.Errorf("Failed to join cluster: %s", err)
			time.Sleep(1 * time.Second)
			cluster.Join()
			return
		}
	}

	members := cluster.serf.Members()
	for len(members) < cluster.params.Bootstrap {

		cluster.params.Logger.Debugf("Waiting for enough nodes to build initial Raft quorum. %d of %d members have joined, contacted %d.",
			cluster.serf.NumNodes(), cluster.params.Bootstrap, contacted)

		time.Sleep(1 * time.Second)
		members = cluster.serf.Members()
	}
	cluster.member = cluster.serf.LocalMember()

	if cluster.params.Bootstrap > 0 {

		raftring := util.NewHashring(1)
		cluster.raft = NewRaftNode(cluster.params, raftring)

		for _, member := range members {
			cluster.hashring.Add(cluster.parseMemberAddress(member))
		}

		nodeIDs := make([]uint64, 0)
		clusterInfo := cluster.raft.GetClusterInfo(cluster.params.RaftOptions.SystemClusterID)
		for nid, raftAddress := range clusterInfo.Nodes {
			cluster.hosts[raftAddress] = nid
			nodeIDs = append(nodeIDs, nid)
		}
		cluster.raftMutex.Lock()
		cluster.rafts[cluster.params.RaftOptions.SystemClusterID] = nodeIDs
		cluster.raftMutex.Unlock()
	} else {
		//node := cluster.serf.Memberlist().LocalNode()

		raftAddress := fmt.Sprintf("%s:%d", cluster.member.Addr.String(), cluster.params.RaftOptions.Port)

		cluster.params.Logger.Debugf("Sending EventWorkerAvailable message with Raft address: %s", raftAddress)

		cluster.serf.UserEvent(gossip.EventWorkerAvailable.String(), []byte(raftAddress), false)
	}

	//cluster.state = ClusterAlive
}

func (cluster *Gossip) Run() {

	var numQueuedEvents int
	for {
		numQueuedEvents = len(cluster.eventCh)
		if numQueuedEvents > serfEventBacklogWarning {
			cluster.params.Logger.Warningf("number of queued serf events above warning threshold.  queued_events=%d, warning_threshold=%d",
				numQueuedEvents, serfEventBacklogWarning)
		}

		select {
		case e := <-cluster.eventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				memberEvent := e.(serf.MemberEvent)
				cluster.params.Logger.Debugf("EventMemberJoin: %+v", memberEvent)
			case serf.EventMemberLeave, serf.EventMemberFailed, serf.EventMemberReap:
				//c.nodeFail(e.(serf.MemberEvent))
			case serf.EventUser:
				userEvent := e.(serf.UserEvent)
				cluster.params.Logger.Debugf("EventUser: %+v", userEvent)
				cluster.handleEvent(userEvent)
			case serf.EventMemberUpdate: // Ignore
				//c.nodeUpdate(e.(serf.MemberEvent))
			case serf.EventQuery: // Ignore
			default:
				cluster.params.Logger.Warningf("unhandled Serf Event: %+v", e)
			}
		case <-cluster.shutdownCh:
			return
		}
	}
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (cluster *Gossip) join(member serf.Member) {

	if cluster.params.NodeID <= 3 {
		return // not concerned with raft bootstrap nodes
	}

	memberAddress := cluster.parseMemberAddress(member)
	cluster.params.Logger.Debugf("A node has joined: %s", memberAddress)

	//meta := cluster.parseNodeMeta(member.Tags)
	meta := member.Tags

	cluster.params.Logger.Errorf("[Gossip.NotifyJoin] meta: %+v", meta)

	raftPort, err := strconv.Atoi(meta["raftPort"])
	if err != nil {
		panic(err)
	}
	raftAddress := cluster.parseRaftAddress(member.Addr.String(), raftPort)

	if cluster.raft == nil {
		cluster.params.Logger.Warningf("[Gossip.NotifyJoin] Aborting, cluster.raft is nil... (clusterID=%d, nodeID=%d, raftAddress=%s, cluster.state=%d)",
			cluster.params.RaftOptions.SystemClusterID, cluster.params.NodeID, raftAddress, cluster.state)
	} else {
		cluster.params.Logger.Warningf("[Gossip.NotifyJoin] Aborting, not the control plane leader... (clusterID=%d, nodeID=%d, raftAddress=%s, cluster.state=%d)",
			cluster.params.RaftOptions.SystemClusterID, cluster.params.NodeID, raftAddress, cluster.state)
	}
}

func (cluster *Gossip) handleEvent(e serf.UserEvent) {

	cluster.params.Logger.Debugf("Received user event: %+v", e.String())
	cluster.params.Logger.Debugf("Received user event type: %+v", e.EventType)
	cluster.params.Logger.Debugf("Received user event payload: %+v", string(e.Payload))

	switch e.Name {
	case gossip.EventWorkerAvailable.String():
		workerAddress := string(e.Payload)
		if cluster.raft != nil && cluster.raft.IsLeader(cluster.params.RaftOptions.SystemClusterID) {
			nodeID, raftAddress, err := cluster.raft.AddNode(cluster.params.RaftOptions.SystemClusterID, workerAddress, 0)
			if err != nil {
				cluster.params.Logger.Errorf("[Gossip.handleEvent] (nodeID=%d) (member=%s) Error: %s",
					cluster.params.NodeID, workerAddress, err)
				return
			}
			workerAssignedMessage, err := json.Marshal(gossip.WorkerAssignedMessage{
				NodeID: nodeID,
				// MemberAddress: This is the node sending the event, not needed for this op
				RaftAddress: raftAddress})
			if err != nil {
				cluster.params.Logger.Errorf("[Gossip.handleEvent] Error encoding worker-added message", err)
				return
			}
			cluster.serf.UserEvent(gossip.EventWorkerAssigned.String(), workerAssignedMessage, false)
			cluster.nodeID = nodeID // "global" node id on the gossip network
		}

	case gossip.EventWorkerAssigned.String():

		var workerAssignedMessage gossip.WorkerAssignedMessage
		err := json.Unmarshal(e.Payload, &workerAssignedMessage)
		if err != nil {
			cluster.params.Logger.Error("[Gossip.handleMessage] worker-assigned error: %s", err)
			return
		}

		if workerAssignedMessage.RaftAddress != cluster.RaftAddress() {
			cluster.hashring.Add(cluster.GossipAddress())
			return
		}

		raftring := util.NewHashring(1)
		params := cluster.params
		params.Raft = []string{workerAssignedMessage.RaftAddress}
		params.NodeID = workerAssignedMessage.NodeID
		cluster.raft = NewRaftNode(params, raftring)
		cluster.member = cluster.serf.LocalMember()
		cluster.hosts[workerAssignedMessage.RaftAddress] = workerAssignedMessage.NodeID

		members := cluster.serf.Members()
		for _, member := range members {
			cluster.hashring.Add(cluster.parseMemberAddress(member))
		}

		nodeIDs := make([]uint64, 0)
		clusterInfo := cluster.raft.GetClusterInfo(params.RaftOptions.SystemClusterID)
		for nid := range clusterInfo.Nodes {
			//for nid, raftAddress := range clusterInfo.Nodes {
			//cluster.nodes[raftAddress] = member[nid-1]
			nodeIDs = append(nodeIDs, nid)
		}
		// Add the cluster system raft
		cluster.raftMutex.Lock()
		cluster.rafts[params.RaftOptions.SystemClusterID] = nodeIDs
		cluster.raftMutex.Unlock()

	case gossip.EventProvisionRequest.String():

		var provisionRequest gossip.ProvisionRequest
		err := json.Unmarshal(e.Payload, &provisionRequest)
		if err != nil {
			cluster.params.Logger.Errorf("[Gossip.handleMessage] provision-request error: %s", err)
			return
		}

		cluster.params.Logger.Debugf("[Gossip.EventProvisionRequest] %+v", provisionRequest)

		// If the node just joined the network, it will start receving provisioning requests immediately.
		// Wait until the Raft cluster has bootstrapped before proceeding.
		cluster.raftMutex.Lock()
		_, ok := cluster.rafts[cluster.params.RaftOptions.SystemClusterID]
		cluster.raftMutex.Unlock()
		for !ok {
			cluster.params.Logger.Debug("[Gossip.handleMessage] provision-request Waiting for the system Raft cluster become available...")
			time.Sleep(1 * time.Second)
		}

		provParams := &common.ProvisionerParams{
			UserID:           provisionRequest.UserID,
			RoleID:           provisionRequest.RoleID,
			OrganizationID:   provisionRequest.OrganizationID,
			FarmName:         provisionRequest.FarmName,
			ConfigStoreType:  provisionRequest.ConfigStoreType,
			StateStoreType:   provisionRequest.StateStoreType,
			DataStoreType:    provisionRequest.DataStoreType,
			ConsistencyLevel: provisionRequest.ConsistencyLevel}
		farmConfig, err := cluster.initializer.BuildConfig(provParams)
		if err != nil {
			cluster.params.Logger.Errorf("[Gossip.handleMessage] Error: %s", err)
			return
		}

		cluster.params.FarmProvisionerChan <- *farmConfig

		// cluster.raft.WaitForClusterReady(provisionRequest.ConfigClusterID)
		// cluster.raft.WaitForClusterReady(provisionRequest.StateClusterID)

		nodeIDs := make([]uint64, 0)
		// use the node ids from the physical nodehost insted of the new raft group
		// clusterID so there is no need to wait for provisioning to complete on
		// every node in the cluster before proceeding.
		clusterInfo := cluster.raft.GetClusterInfo(cluster.params.RaftOptions.SystemClusterID)
		for nid, raftAddress := range clusterInfo.Nodes {
			cluster.hosts[raftAddress] = nid
			nodeIDs = append(nodeIDs, nid)
		}
		cluster.raftMutex.Lock()
		cluster.rafts[provisionRequest.StateClusterID] = nodeIDs
		cluster.rafts[provisionRequest.ConfigClusterID] = nodeIDs
		cluster.raftMutex.Unlock()

		hashring := cluster.raft.GetHashring()

		// Set config leader
		host, err := hashring.GetLeast(strconv.Itoa(int(provisionRequest.ConfigClusterID)))
		if err != nil {
			cluster.params.Logger.Errorf("[Gossip.handleMessage] Error: %s", err)
			return
		}
		hashring.Inc(host)
		cluster.raft.GetNodeHost().RequestLeaderTransfer(provisionRequest.ConfigClusterID, cluster.hosts[host])

		// Set state leader
		host, err = hashring.GetLeast(strconv.Itoa(int(provisionRequest.StateClusterID)))
		if err != nil {
			cluster.params.Logger.Errorf("[Gossip.handleMessage] Error: %s", err)
			return
		}
		hashring.Inc(host)
		//hashring.Inc(host)
		cluster.raft.GetNodeHost().RequestLeaderTransfer(provisionRequest.StateClusterID, cluster.hosts[host])

	default:
		cluster.params.Logger.Errorf("Received unknown user event: %s", e.Name)
	}
}

func (cluster *Gossip) GetSystemRaft() RaftNode {
	return cluster.raft
}

func (cluster *Gossip) GetHashring() *util.Consistent {
	return cluster.hashring
}

func (cluster *Gossip) Shutdown() error {
	if err := cluster.serf.Leave(); err != nil {
		cluster.params.Logger.Errorf("[Gossip.Shutdown] Error: %s", err)
	}
	return cluster.serf.Shutdown()
}

// Returns the unserialized node metadata
func (cluster *Gossip) GetNodeMeta() *nodeMeta {
	return cluster.nodemeta
}

func (cluster *Gossip) GetMemberCount() int {
	return cluster.serf.NumNodes()
}

func (cluster *Gossip) GetSerfStats() map[string]string {
	return cluster.serf.Stats()
}

func (cluster *Gossip) GossipAddress() string {
	return fmt.Sprintf("%s:%d", cluster.member.Addr.String(), cluster.member.Port)
}

func (cluster *Gossip) RaftAddress() string {
	return fmt.Sprintf("%s:%d", cluster.member.Addr.String(), cluster.params.RaftOptions.Port)
}

func (cluster *Gossip) HasRaft(clusterID uint64) bool {
	cluster.raftMutex.Lock()
	defer cluster.raftMutex.Unlock()
	if _, ok := cluster.rafts[clusterID]; !ok {
		return false
	}
	return true
}

/*
func (cluster *Gossip) isMe(clusterID, nodeID uint64) bool {
	if raft, ok := cluster.rafts[clusterID]; !ok {
		return false
	}
	for i, nodeID := range cluster.rafts[clusterID] {
		if nodeID == cluster.nodeID {
			return true
		}
	}
	return false
}*/

// Provision sends a new provisioning request to the gossip network. This method
// does not wait to confirm the cluster is ready before returning, but instead delegates
// that responsibility to the farm provisioner channel consumer.
func (cluster *Gossip) Provision(params *common.ProvisionerParams) error {

	cluster.params.Logger.Errorf("Provision request! gossip.address=%s, raft.peers=%+v",
		cluster.GossipAddress(), cluster.raft.GetPeers())

	farmKey := fmt.Sprintf("%d-%s", params.OrganizationID, params.FarmName)
	farmID := cluster.params.IdGenerator.NewID(farmKey)

	stateClusterKey := fmt.Sprintf("%s-%d", params.FarmName, farmID)
	stateClusterID := cluster.params.IdGenerator.NewID(stateClusterKey)

	bytes, err := json.Marshal(&gossip.ProvisionRequest{
		FarmName:         params.FarmName,
		ConfigClusterID:  farmID,
		StateClusterID:   stateClusterID,
		ConfigStoreType:  params.ConfigStoreType,
		StateStoreType:   params.StateStoreType,
		DataStoreType:    params.DataStoreType,
		ConsistencyLevel: params.ConsistencyLevel,
		// include nodeID so a response can be given back
		// to the user who initiated the request
		NodeID: cluster.nodeID,
		UserID: params.UserID,
		RoleID: params.RoleID})
	if err != nil {
		return err
	}
	cluster.serf.UserEvent(gossip.EventProvisionRequest.String(), bytes, false)

	cluster.raft.WaitForClusterReady(farmID)
	cluster.raft.WaitForClusterReady(stateClusterID)

	cluster.raftMutex.Lock()
	_, ok := cluster.rafts[stateClusterID]
	cluster.raftMutex.Unlock()
	for !ok {
		cluster.params.Logger.Debugf("[Gossip.Provision] Waiting on provisioning completion: stateClusterID=%d, configClusterID=%d",
			stateClusterID, farmID)
		time.Sleep(1 * time.Second)
		cluster.raftMutex.Lock()
		_, ok = cluster.rafts[stateClusterID]
		cluster.raftMutex.
			Unlock()
	}
	return nil
}

func (cluster *Gossip) GetParams() *util.ClusterParams {
	return cluster.params
}

func (cluster *Gossip) parseMemberAddress(member serf.Member) string {
	return fmt.Sprintf("%s:%d", member.Addr.String(), member.Port)
}

func (cluster *Gossip) parseRaftAddress(nodeAddress string, raftPort int) string {
	pieces := strings.Split(nodeAddress, ":")
	if pieces[0][0] == '[' {
		pieces[0] = "localhost"
	}
	return fmt.Sprintf("%s:%d", pieces[0], raftPort)
}

func (cluster *Gossip) parseNodeMeta(data []byte) nodeMeta {
	var meta nodeMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		cluster.params.Logger.Errorf("[Gossip.parseNodeMeta] Error unpacking node metadata: %s", err)
		return nodeMeta{}
	}
	return meta
}
