//go:build cluster && rocksdb
// +build cluster,rocksdb

package cluster

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"time"

	"github.com/lni/dragonboat/v3"
	"github.com/lni/dragonboat/v3/client"
	"github.com/lni/dragonboat/v3/config"
	"github.com/lni/dragonboat/v3/logger"
	"github.com/lni/dragonboat/v3/plugin/rocksdb"
	"github.com/lni/dragonboat/v3/raftio"

	"github.com/lni/go-cropdroid/common"

	sm "github.com/lni/dragonboat/v3/statemachine"
)

const (
	CONTROL_PLANE_CLUSTER_ID = 0
	QUORUM_MIN_NODES         = 3
	QUORUM_MAX_NODES         = 7
)

var (
	ErrOffline = errors.New("Cluster offline")
)

type RaftCluster interface {
	AddNode(clusterID uint64, address string, retryCount int) (uint64, string, error)
	CreateStateCluster(id uint64, sm state.FarmStateMachine) error
	CreateConfigCluster(clusterID uint64, sm state.FarmConfigMachine) error
	GetConfig() config.Config
	GetClusterCount() int
	GetClusterInfo(clusterID uint64) *ClusterInfo
	GetClusterInfos() []*ClusterInfo
	GetClusterStatus() []*ClusterStatus
	GetHashring() *Consistent
	GetLeaderID(clusterID uint64) (uint64, bool, error)
	GetLeaderInfo(clusterID uint64) *ClusterInfo
	GetNodeCount() int
	GetNodeHost() *dragonboat.NodeHost
	GetParams() *ClusterParams
	GetPeers() map[uint64]string
	Hash(key string) uint64
	IsLeader(clusterID uint64) bool
	Shutdown() error
	ReadLocal(clusterID uint64, query interface{}) (interface{}, error)
	SyncPropose(clusterID uint64, cmd []byte) error
	SyncRead(clusterID uint64, query interface{}) (interface{}, error)
	WaitForClusterReady(clusterID uint64) bool
}

type ClusterStatus struct {
	ClusterID         string            `json:"clusterId"`
	NodeID            string            `json:"nodeId"`
	Nodes             map[uint64]string `json:"nodes"`
	ConfigChangeIndex string            `json:"configChangeIndex"`
	StateMachineType  sm.Type           `json:"stateMachineType"`
	IsLeader          bool              `json:"isLeader"`
	IsObserver        bool              `json:"isObserver"`
	IsWitness         bool              `json:"isWitness"`
	Pending           bool              `json:"pending"`
}

type ClusterInfo struct {
	ClusterID         uint64            `json:"clusterId"`
	NodeID            uint64            `json:"nodeId"`
	Nodes             map[uint64]string `json:"nodes"`
	ConfigChangeIndex uint64            `json:"configChangeIndex"`
	StateMachineType  sm.Type           `json:"stateMachineType"`
	IsLeader          bool              `json:"isLeader"`
	IsObserver        bool              `json:"isObserver"`
	IsWitness         bool              `json:"isWitness"`
	Pending           bool              `json:"pending"`
}

type Raft struct {
	nodeHost     *dragonboat.NodeHost
	params       *ClusterParams
	session      map[uint64]*client.Session
	config       config.Config
	peers        map[uint64]string
	index        map[string]uint64 // map node address to raft node id
	proposalChan map[uint64]chan []byte
	timeout      time.Duration
	hashring     *Consistent // raft group leaders on this raft
	RaftCluster
}

type RocksDBFactory struct {
	factory rocksdb.Factory
	config.LogDBFactory
}

func NewRaftCluster(params *ClusterParams, hashring *Consistent) RaftCluster {

	// As a summary, when -
	//  - starting a brand new Raft cluster, set join to false and specify all initial
	//    member node details in the initialMembers map.
	//  - joining a new node to an existing Raft cluster, set join to true and leave
	//    the initialMembers map empty. This requires the joining node to have already
	//    been added as a member node of the Raft cluster.
	//  - restarting an crashed or stopped node, set join to false and leave the
	//    initialMembers map to be empty. This applies to both initial member nodes
	//    and those joined later.

	params.logger.Debugf("Starting raft server with rocksdb backend: %+v", params)

	var nodeAddr string
	raftNodeID := params.nodeID + uint64(1)
	initialMembers := make(map[uint64]string)
	if !params.join {
		// Bootsrap the cluster
		for i, peer := range params.raft {
			initialMembers[uint64(i+1)] = peer
			hashring.Add(peer)
		}
	} else {
		params.logger.Fatal("Adding additional raft members needs to be debugged... raft_pebble.go#131")
		nodeAddr = params.raft[0]
		hashring.Add(nodeAddr)
	}
	if nodeAddr == "" {
		nodeAddr = initialMembers[raftNodeID]
	}

	fmt.Fprintf(os.Stdout, "Raft node address: %s\n", nodeAddr)

	// change the log verbosity
	logger.GetLogger("raft").SetLevel(logger.ERROR)
	logger.GetLogger("rsm").SetLevel(logger.WARNING)
	logger.GetLogger("transport").SetLevel(logger.WARNING)
	logger.GetLogger("grpc").SetLevel(logger.WARNING)

	rc := config.Config{
		NodeID:             raftNodeID,
		ElectionRTT:        5,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    10,
		CompactionOverhead: 5,
		ClusterID:          params.clusterID,
	}

	r := &Raft{
		params:       params,
		proposalChan: make(map[uint64]chan []byte, 2),
		config:       rc,
		peers:        initialMembers,
		timeout:      3 * time.Second,
		session:      make(map[uint64]*client.Session, 1),
		hashring:     hashring}

	//factory := &rocksdb.Factory{}
	nhc := config.NodeHostConfig{
		WALDir:            fmt.Sprintf("%s/node%d", params.dataDir, r.config.NodeID),
		NodeHostDir:       fmt.Sprintf("%s/node%d", params.dataDir, r.config.NodeID),
		RTTMillisecond:    200,
		RaftAddress:       nodeAddr,
		RaftEventListener: r,
		Expert:            config.ExpertConfig{LogDBFactory: &rocksdb.Factory{}},
	}

	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		panic(err)
	}
	r.nodeHost = nh
	r.session[params.clusterID] = nh.GetNoOPSession(params.clusterID)
	r.proposalChan[params.clusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)

	diskkv := state.NewDiskKV(params.dataDir)
	if err := nh.StartOnDiskCluster(initialMembers, params.join, diskkv.CreateStateMachine, rc); err != nil {
		params.logger.Fatalf("Failed to create system raft cluster: %s", err)
	}

	if isLeader := r.WaitForClusterReady(params.clusterID); isLeader == true {
		r.hashring.Inc(nodeAddr)
	} else {
		leaderID, _, _ := r.nodeHost.GetLeaderID(params.clusterID)
		for nodeID, raftAddress := range r.GetClusterInfo(params.clusterID).Nodes {
			r.hashring.Add(raftAddress)
			if nodeID == leaderID {
				r.hashring.Inc(raftAddress)
			}
		}
	}
	return r
	//defer r.Stop()
}

func (r *Raft) WaitForClusterReady(clusterID uint64) bool {
	var getLeaderFunc = func() (uint64, bool, error) {
		leaderID, ready, err := r.nodeHost.GetLeaderID(clusterID)
		r.params.logger.Debugf("[Raft.WaitForClusterReady] clusterID=%d, leaderID: %d, nodeID=%d, ready=%t",
			clusterID, leaderID, r.config.NodeID, ready)
		return leaderID, ready, err
	}
	leaderID, ready, err := getLeaderFunc()
	if err != nil {
		r.params.logger.Error(err)
	}
	for !ready {
		r.params.logger.Debugf("[Raft.WaitForClusterReady] Waiting for cluster %d to become ready...", clusterID)
		time.Sleep(1 * time.Second)
		_, ready, _ = getLeaderFunc()
	}
	for leaderID == 0 {
		r.params.logger.Debugf("[Raft.WaitForClusterReady] Waiting on cluster %d leader election...", clusterID)
		time.Sleep(1 * time.Second)
		leaderID, _, _ = getLeaderFunc()
	}
	if leaderID == r.config.NodeID {
		return true
	}
	return false
}

func (r *Raft) GetHashring() *Consistent {
	return r.hashring
}

func (r *Raft) AddNode(clusterID uint64, address string, retryCount int) (uint64, string, error) {
	nodeID := uint64(r.GetNodeCount() + 1)
	r.params.logger.Errorf("[Raft.AddNode] Adding new node, id=%d, address=%s to cluster %d", nodeID, address, clusterID)
	if r.IsLeader(r.params.clusterID) {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		membership, err := r.nodeHost.SyncGetClusterMembership(ctx, clusterID)
		cancel()
		if err != nil {
			return 0, "", err
		}
		ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
		if err := r.nodeHost.SyncRequestAddNode(ctx, clusterID, nodeID, address, membership.ConfigChangeID); err != nil {
			r.params.logger.Errorf("Failed to add new Raft node (clusterID=%d, nodeID=%d, retryCount=%d) Error: %s",
				clusterID, nodeID, retryCount, err)
			time.Sleep(1 * time.Second)
			r.AddNode(clusterID, address, retryCount+1)
		}
		cancel()
		/*
			newConfig := r.config
			newConfig.ClusterID = clusterID
			if clusterID == CONTROL_PLANE_CLUSTER_ID {
				if err := r.nodeHost.StartCluster(r.peers, r.params.join, state.NewSystemStateMachine, newConfig); err != nil {
					//if err == dragonboat.ErrClusterAlreadyExist && retryCount < 5 {
					//	r.AddNode(clusterID, address, retryCount+1)
					//  time.Sleep(1 * time.Second)
					//}
					r.params.logger.Errorf("Failed to add new node (clusterID=%d, nodeID=%d, address=%s): %s", clusterID, nodeID, address, err)
					return err
				}
			} else {
				diskkv := state.NewDiskKV(r.params.dataDir)
				if err := r.nodeHost.StartOnDiskCluster(r.peers, r.params.join, diskkv.CreateStateMachine, newConfig); err != nil {
					//if err == dragonboat.ErrClusterAlreadyExist && retryCount < 5 {
					//  r.AddNode(clusterID, address, retryCount+1)
					//  time.Sleep(1 * time.Second)
					//}
					r.params.logger.Errorf("Failed to add new node: %s", err)
					return err
				}
			}*/
		//r.hashring.Add(address)
	}
	r.hashring.Add(address)
	return nodeID, address, nil
}

func (r *Raft) DeleteNode(clusterID, nodeID uint64, address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	err := r.nodeHost.SyncRequestDeleteNode(ctx, clusterID, nodeID, 0)
	cancel()
	r.hashring.Remove(address)
	return err
}

func (r *Raft) CreateStateCluster(clusterID uint64, sm state.FarmStateMachine) error {
	newConfig := r.config
	newConfig.ClusterID = clusterID
	r.session[clusterID] = r.nodeHost.GetNoOPSession(clusterID)
	r.proposalChan[clusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)
	if err := r.nodeHost.StartCluster(r.peers, false, sm.CreateStateMachine, newConfig); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		return err
	}
	return nil
}

func (r *Raft) CreateConfigCluster(clusterID uint64, sm state.FarmConfigMachine) error {
	newConfig := r.config
	newConfig.ClusterID = clusterID
	r.session[clusterID] = r.nodeHost.GetNoOPSession(clusterID)
	r.proposalChan[clusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)
	if err := r.nodeHost.StartCluster(r.peers, false, sm.CreateConfigMachine, newConfig); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		return err
	}
	return nil
}

func (r *Raft) Shutdown() error {
	if err := r.nodeHost.StopNode(r.params.clusterID, r.config.NodeID); err != nil {
		r.params.logger.Error("Error shutting down Raft node. clusterID: %d, nodeID=%d, error=%s",
			r.params.clusterID, r.config.NodeID, err)
	}
	//for clusterID, _ := range r.session {
	//for clusterID, session := range r.session {
	/*
		panic: not a regular session

		goroutine 1 [running]:
		github.com/lni/dragonboat/v3/client.(*Session).assertRegularSession(...)
				go/pkg/mod/github.com/lni/dragonboat/v3@v3.2.2/client/session.go:125
	*/
	//ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	//r.nodeHost.SyncCloseSession(ctx, session)
	//cancel()
	//if err := r.nodeHost.StopCluster(clusterID); err != nil {
	//	r.params.logger.Error("Error shutting down Raft cluster %d. Error: ", clusterID, err)
	//}
	//delete(r.session, clusterID)
	//}
	r.nodeHost.Stop()
	return nil
}

func (r *Raft) GetParams() *ClusterParams {
	return r.params
}

func (r *Raft) GetPeers() map[uint64]string {
	return r.peers
}

func (r *Raft) GetConfig() config.Config {
	return r.config
}

func (r *Raft) IsLeader(clusterID uint64) bool {
	opt := dragonboat.NodeHostInfoOption{SkipLogInfo: true}
	membership := r.nodeHost.GetNodeHostInfo(opt)
	for _, member := range membership.ClusterInfoList {
		if member.IsLeader && member.NodeID == r.config.NodeID && member.ClusterID == clusterID {
			return true
		}
	}
	return false
}

func (r *Raft) GetNodeCount() int {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	membership, err := r.nodeHost.GetClusterMembership(ctx, r.params.clusterID)
	cancel()
	if err != nil {
		r.params.logger.Error(err)
		return 0
	}
	return len(membership.Nodes)
}

func (r *Raft) GetClusterCount() int {
	opt := dragonboat.NodeHostInfoOption{SkipLogInfo: true}
	return len(r.nodeHost.GetNodeHostInfo(opt).ClusterInfoList)
}

// Returns the cluster info for the specified cluster
func (r *Raft) GetClusterInfo(clusterID uint64) *ClusterInfo {
	opt := dragonboat.NodeHostInfoOption{SkipLogInfo: true}
	membership := r.nodeHost.GetNodeHostInfo(opt)
	for _, member := range membership.ClusterInfoList {
		if member.ClusterID == clusterID {
			return &ClusterInfo{
				ClusterID:         member.ClusterID,
				NodeID:            member.NodeID,
				Nodes:             member.Nodes,
				ConfigChangeIndex: member.ConfigChangeIndex,
				StateMachineType:  member.StateMachineType,
				IsLeader:          member.IsLeader,
				IsObserver:        member.IsObserver,
				IsWitness:         member.IsWitness,
				Pending:           member.Pending}
		}
	}
	return nil
}

// GetClusterInfos returns an info object for every raft group in the cluster
func (r *Raft) GetClusterInfos() []*ClusterInfo {
	opt := dragonboat.NodeHostInfoOption{SkipLogInfo: true}
	membership := r.nodeHost.GetNodeHostInfo(opt)
	clusters := make([]*ClusterInfo, len(membership.ClusterInfoList))
	for i, member := range membership.ClusterInfoList {
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
	}
	return clusters
}

// GetClusterStatus returns an info object for every raft group in the cluster
// using string data types in place of uint64s (to prevent mis-rounding to int
// when browsing / UI).
func (r *Raft) GetClusterStatus() []*ClusterStatus {
	opt := dragonboat.NodeHostInfoOption{SkipLogInfo: true}
	membership := r.nodeHost.GetNodeHostInfo(opt)
	clusters := make([]*ClusterStatus, len(membership.ClusterInfoList))
	for i, member := range membership.ClusterInfoList {
		clusters[i] = &ClusterStatus{
			ClusterID:         fmt.Sprintf("%d", member.ClusterID),
			NodeID:            fmt.Sprintf("%d", member.NodeID),
			Nodes:             member.Nodes,
			ConfigChangeIndex: fmt.Sprintf("%d", member.ConfigChangeIndex),
			StateMachineType:  member.StateMachineType,
			IsLeader:          member.IsLeader,
			IsObserver:        member.IsObserver,
			IsWitness:         member.IsWitness,
			Pending:           member.Pending}
	}
	return clusters
}

// Returns the cluster info if this node is the leader
func (r *Raft) GetLeaderInfo(clusterID uint64) *ClusterInfo {
	opt := dragonboat.NodeHostInfoOption{SkipLogInfo: true}
	membership := r.nodeHost.GetNodeHostInfo(opt)
	for _, member := range membership.ClusterInfoList {
		if member.IsLeader && member.NodeID == r.config.NodeID && clusterID == r.params.clusterID {
			return &ClusterInfo{
				ClusterID:         member.ClusterID,
				NodeID:            member.NodeID,
				Nodes:             member.Nodes,
				ConfigChangeIndex: member.ConfigChangeIndex,
				StateMachineType:  member.StateMachineType,
				IsLeader:          member.IsLeader,
				IsObserver:        member.IsObserver,
				IsWitness:         member.IsWitness,
				Pending:           member.Pending}
		}
	}
	return nil
}

func (r *Raft) LeaderUpdated(info raftio.LeaderInfo) {
	clusterInfo := r.GetClusterInfo(info.ClusterID)
	if clusterInfo != nil {
		r.params.logger.Warningf("Raft cluster membership updated! %+v", info)
	}
}

func (r *Raft) GetNodeHost() *dragonboat.NodeHost {
	return r.nodeHost
}

func (r *Raft) SyncPropose(clusterID uint64, cmd []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	result, err := r.nodeHost.SyncPropose(ctx, r.session[clusterID], cmd)
	cancel()
	if err != nil {
		r.params.logger.Errorf("[Raft.SyncPropose] Error: %s", err)
		return err
	}
	r.params.logger.Debugf("[Raft.SyncPropose] Raft confirmation: clusterID=%d, nodeID=%d, message=%s, result=%+v",
		clusterID, r.config.NodeID, string(cmd), result)

	return nil
}

func (r *Raft) SyncRead(clusterID uint64, query interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	result, err := r.nodeHost.SyncRead(ctx, clusterID, query)
	cancel()
	return result, err
}

func (r *Raft) ReadLocal(clusterID uint64, query interface{}) (interface{}, error) {
	return r.nodeHost.StaleRead(clusterID, query)
}

func (r *Raft) GetLeaderID(clusterID uint64) (uint64, bool, error) {
	return r.nodeHost.GetLeaderID(clusterID)
}

func (r *Raft) Hash(key string) uint64 {
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(key))
	return clusterHash.Sum64()
}
