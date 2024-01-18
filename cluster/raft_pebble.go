//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"sync"
	"time"

	"github.com/lni/dragonboat/v3"
	"github.com/lni/dragonboat/v3/client"
	"github.com/lni/dragonboat/v3/config"
	"github.com/lni/dragonboat/v3/logger"
	"github.com/lni/dragonboat/v3/raftio"

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
	"github.com/jeremyhahn/go-cropdroid/common"

	"github.com/jeremyhahn/go-cropdroid/cluster/util"

	sm "github.com/lni/dragonboat/v3/statemachine"
)

var (
	ErrOffline = errors.New("Cluster offline")
)

type RaftNode interface {
	AddNode(clusterID uint64, address string, retryCount int) (uint64, string, error)
	CreateRegularCluster(clusterID uint64, join bool, stateMachineFunc sm.CreateStateMachineFunc) error
	CreateConcurrentCluster(clusterID uint64, join bool, stateMachineFunc sm.CreateConcurrentStateMachineFunc) error
	CreateOnDiskCluster(clusterID uint64, join bool, stateMachineFunc sm.CreateOnDiskStateMachineFunc) error
	GetConfig() config.Config
	GetClusterCount() int
	GetClusterInfo(clusterID uint64) *ClusterInfo
	GetClusterInfos() []*ClusterInfo
	GetClusterStatus() []*ClusterStatus
	GetHashring() *util.Consistent
	GetLeaderID(clusterID uint64) (uint64, bool, error)
	GetLeaderInfo(clusterID uint64) *ClusterInfo
	GetNodeCount() int
	GetNodeHost() *dragonboat.NodeHost
	GetParams() *util.ClusterParams
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
	nodeHost          *dragonboat.NodeHost
	params            *util.ClusterParams
	session           map[uint64]*client.Session
	sessionMutex      *sync.RWMutex
	config            config.Config
	peers             map[uint64]string
	index             map[string]uint64 // map node address to raft node id
	proposalChan      map[uint64]chan []byte
	proposalChanMutex *sync.RWMutex
	timeout           time.Duration
	hashring          *util.Consistent // raft group leaders on this raft
	RaftNode
}

func NewRaftNode(params *util.ClusterParams, hashring *util.Consistent) RaftNode {

	// As a summary, when -
	//  - starting a brand new Raft cluster, set join to false and specify all initial
	//    member node details in the initialMembers map.
	//  - joining a new node to an existing Raft cluster, set join to true and leave
	//    the initialMembers map empty. This requires the joining node to have already
	//    been added as a member node of the Raft cluster.
	//  - restarting an crashed or stopped node, set join to false and leave the
	//    initialMembers map to be empty. This applies to both initial member nodes
	//    and those joined later.

	params.Logger.Debugf("Starting raft server with pebble backend: %+v", params)

	/*
		if len(params.peers) == 0 {
			panic("cluster peers required")
		}*/

	var nodeAddr string
	raftNodeID := params.NodeID + uint64(1)
	initialMembers := make(map[uint64]string)
	if !params.Join {
		// Bootsrap the cluster
		for i, peer := range params.Raft {
			initialMembers[uint64(i+1)] = peer
			hashring.Add(peer)
		}
	} else {
		params.Logger.Fatal("Adding additional raft members needs to be debugged... raft_pebble.go#NewRaftCluster()")
		nodeAddr = params.Raft[0]
		hashring.Add(nodeAddr)
	}
	if nodeAddr == "" {
		nodeAddr = initialMembers[raftNodeID]
	}

	fmt.Fprintf(os.Stdout, "Raft node=%d, address=%s\n", raftNodeID, nodeAddr)

	// change the log verbosity
	logger.GetLogger("raft").SetLevel(logger.ERROR)
	logger.GetLogger("rsm").SetLevel(logger.WARNING)
	logger.GetLogger("transport").SetLevel(logger.WARNING)
	logger.GetLogger("grpc").SetLevel(logger.WARNING)

	rc := config.Config{
		NodeID: raftNodeID,
		//ElectionRTT:        5,
		//HeartbeatRTT:       1,
		ElectionRTT:        10,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    10,
		CompactionOverhead: 5,
		ClusterID:          params.ClusterID,
	}

	r := &Raft{
		params:            params,
		proposalChan:      make(map[uint64]chan []byte, 2),
		proposalChanMutex: &sync.RWMutex{},
		config:            rc,
		peers:             initialMembers,
		//timeout:           3 * time.Second,
		timeout:      3600 * time.Second,
		session:      make(map[uint64]*client.Session, 1),
		sessionMutex: &sync.RWMutex{},
		hashring:     hashring}

	nhc := config.NodeHostConfig{
		WALDir:            fmt.Sprintf("%s/node%d", params.DataDir, r.config.NodeID),
		NodeHostDir:       fmt.Sprintf("%s/node%d", params.DataDir, r.config.NodeID),
		RTTMillisecond:    100,
		RaftAddress:       nodeAddr,
		RaftEventListener: r,
	}

	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		panic(err)
	}
	r.nodeHost = nh
	r.session[params.ClusterID] = nh.GetNoOPSession(params.ClusterID)
	r.proposalChan[params.ClusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)

	systemSM := statemachine.NewServerConfigMachine(params.Logger, params.IdGenerator, params.DataDir, params.ClusterID, params.NodeID)
	if err := nh.StartOnDiskCluster(initialMembers, params.Join, systemSM.CreateServerConfigMachine, rc); err != nil {
		params.Logger.Fatalf("Failed to create system raft cluster: %s", err)
	}

	if isLeader := r.WaitForClusterReady(params.ClusterID); isLeader == true {
		r.hashring.Inc(nodeAddr)
	} else {
		leaderID, _, _ := r.nodeHost.GetLeaderID(params.ClusterID)
		for nodeID, raftAddress := range r.GetClusterInfo(params.ClusterID).Nodes {
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
		r.params.Logger.Debugf("[Raft.WaitForClusterReady] clusterID=%d, leaderID: %d, nodeID=%d, ready=%t",
			clusterID, leaderID, r.config.NodeID, ready)
		return leaderID, ready, err
	}
	leaderID, ready, err := getLeaderFunc()
	if err != nil {
		r.params.Logger.Error(err)
	}
	for !ready {
		r.params.Logger.Infof("[Raft.WaitForClusterReady] Waiting for cluster %d to become ready...", clusterID)
		time.Sleep(1 * time.Second)
		_, ready, _ = getLeaderFunc()
	}
	if r.params.RaftOptions.RequestedLeaderID > 0 && leaderID != uint64(r.params.RaftOptions.RequestedLeaderID) {
		r.params.Logger.Infof("[Raft.WaitForClusterReady] Requesting node %d be raft leader", r.params.RaftOptions.RequestedLeaderID)
		err = r.nodeHost.RequestLeaderTransfer(r.params.ClusterID, uint64(r.params.RaftOptions.RequestedLeaderID))
		if err != nil {
			r.params.Logger.Error(err)
		}
	}
	for leaderID == 0 {
		r.params.Logger.Infof("[Raft.WaitForClusterReady] Waiting on cluster %d leader election...", clusterID)
		time.Sleep(1 * time.Second)
		leaderID, _, _ = getLeaderFunc()
	}
	if leaderID == r.config.NodeID {
		return true
	}
	return false
}

func (r *Raft) GetHashring() *util.Consistent {
	return r.hashring
}

func (r *Raft) AddNode(clusterID uint64, address string, retryCount int) (uint64, string, error) {
	nodeID := uint64(r.GetNodeCount() + 1)
	r.params.Logger.Errorf("[Raft.AddNode] Adding new node, id=%d, address=%s to cluster %d", nodeID, address, clusterID)
	if r.IsLeader(r.params.ClusterID) {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		membership, err := r.nodeHost.SyncGetClusterMembership(ctx, clusterID)
		cancel()
		if err != nil {
			return 0, "", err
		}
		ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
		if err := r.nodeHost.SyncRequestAddNode(ctx, clusterID, nodeID, address, membership.ConfigChangeID); err != nil {
			r.params.Logger.Errorf("Failed to add new Raft node (clusterID=%d, nodeID=%d, retryCount=%d) Error: %s",
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
					r.params.Logger.Errorf("Failed to add new node (clusterID=%d, nodeID=%d, address=%s): %s", clusterID, nodeID, address, err)
					return err
				}
			} else {
				diskkv := state.NewDiskKV(r.params.DataDir)
				if err := r.nodeHost.StartOnDiskCluster(r.peers, r.params.join, diskkv.CreateStateMachine, newConfig); err != nil {
					//if err == dragonboat.ErrClusterAlreadyExist && retryCount < 5 {
					//  r.AddNode(clusterID, address, retryCount+1)
					//  time.Sleep(1 * time.Second)
					//}
					r.params.Logger.Errorf("Failed to add new node: %s", err)
					return err
				}
			}*/
		//r.hashring.Add(address)
	}
	r.hashring.Add(address)
	return nodeID, address, nil
}

func (r *Raft) CreateRegularCluster(clusterID uint64, join bool, stateMachineFunc sm.CreateStateMachineFunc) error {
	newConfig := r.config
	newConfig.ClusterID = clusterID

	r.sessionMutex.Lock()
	r.session[clusterID] = r.nodeHost.GetNoOPSession(clusterID)
	r.sessionMutex.Unlock()

	r.proposalChanMutex.Lock()
	r.proposalChan[clusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)
	r.proposalChanMutex.Unlock()

	if err := r.nodeHost.StartCluster(r.peers, false, stateMachineFunc, newConfig); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		return err
	}
	return nil
}

func (r *Raft) CreateConcurrentCluster(clusterID uint64, join bool, stateMachineFunc sm.CreateConcurrentStateMachineFunc) error {
	newConfig := r.config
	newConfig.ClusterID = clusterID

	r.sessionMutex.Lock()
	r.session[clusterID] = r.nodeHost.GetNoOPSession(clusterID)
	r.sessionMutex.Unlock()

	r.proposalChanMutex.Lock()
	r.proposalChan[clusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)
	r.proposalChanMutex.Unlock()

	if err := r.nodeHost.StartConcurrentCluster(r.peers, false, stateMachineFunc, newConfig); err != nil {
		r.params.Logger.Error(os.Stderr, "failed to add IConcurrentStateMachine cluster %s, error=%v\n", clusterID, err)
		return err
	}
	return nil
}

func (r *Raft) CreateOnDiskCluster(clusterID uint64, join bool, stateMachineFunc sm.CreateOnDiskStateMachineFunc) error {
	newConfig := r.config
	newConfig.ClusterID = clusterID

	r.sessionMutex.Lock()
	r.session[clusterID] = r.nodeHost.GetNoOPSession(clusterID)
	r.sessionMutex.Unlock()

	r.proposalChanMutex.Lock()
	r.proposalChan[clusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)
	r.proposalChanMutex.Unlock()

	if err := r.nodeHost.StartOnDiskCluster(r.peers, join, stateMachineFunc, newConfig); err != nil {
		r.params.Logger.Errorf("failed to create IOnDiskStateMachine cluster %d, error=%v", clusterID, err)
		return err
	}
	return nil
}

// func (r *Raft) CreateFarmConfigCluster(clusterID uint64, sm statemachine.FarmConfigMachine) error {
// 	newConfig := r.config
// 	newConfig.ClusterID = clusterID
// 	r.session[clusterID] = r.nodeHost.GetNoOPSession(clusterID)
// 	r.proposalChan[clusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)
// 	if err := r.nodeHost.StartOnDiskCluster(r.peers, false, sm.CreateFarmConfigMachine, newConfig); err != nil {
// 		//if err := r.nodeHost.StartOnDiskCluster(r.peers, false, statemachine.NewDiskKV, newConfig); err != nil {
// 		//if err := r.nodeHost.StartCluster(r.peers, false, sm.CreateFarmConfigMachine, newConfig); err != nil {
// 		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
// 		return err
// 	}
// 	return nil
// }

// func (r *Raft) CreateFarmStateCluster(clusterID uint64, sm statemachine.FarmStateMachine) error {
// 	newConfig := r.config
// 	newConfig.ClusterID = clusterID

// 	r.sessionMutex.Lock()
// 	r.session[clusterID] = r.nodeHost.GetNoOPSession(clusterID)
// 	r.sessionMutex.Unlock()

// 	r.proposalChanMutex.Lock()
// 	r.proposalChan[clusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)
// 	r.proposalChanMutex.Unlock()

// 	//if err := r.nodeHost.StartCluster(r.peers, false, sm.CreateStateMachine, newConfig); err != nil {
// 	if err := r.nodeHost.StartConcurrentCluster(r.peers, false, sm.CreateFarmStateMachine, newConfig); err != nil {
// 		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
// 		return err
// 	}
// 	return nil
// }

// func (r *Raft) CreateDeviceStateCluster(clusterID uint64, sm statemachine.DeviceStateMachine) error {
// 	newConfig := r.config
// 	newConfig.ClusterID = clusterID

// 	r.sessionMutex.Lock()
// 	r.session[clusterID] = r.nodeHost.GetNoOPSession(clusterID)
// 	r.sessionMutex.Unlock()

// 	r.proposalChanMutex.Lock()
// 	r.proposalChan[clusterID] = make(chan []byte, common.BUFFERED_CHANNEL_SIZE)
// 	r.proposalChanMutex.Unlock()

// 	if err := r.nodeHost.StartCluster(r.peers, false, sm.CreateDeviceStateMachine, newConfig); err != nil {
// 		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
// 		return err
// 	}
// 	return nil
// }

func (r *Raft) DeleteNode(clusterID, nodeID uint64, address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	err := r.nodeHost.SyncRequestDeleteNode(ctx, clusterID, nodeID, 0)
	cancel()
	r.hashring.Remove(address)
	return err
}

func (r *Raft) Shutdown() error {
	if err := r.nodeHost.StopNode(r.params.ClusterID, r.config.NodeID); err != nil {
		r.params.Logger.Error("Error shutting down Raft node. clusterID: %d, nodeID=%d, error=%s",
			r.params.ClusterID, r.config.NodeID, err)
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
	//	r.params.Logger.Error("Error shutting down Raft cluster %d. Error: ", clusterID, err)
	//}
	//delete(r.session, clusterID)
	//}
	r.nodeHost.Stop()
	return nil
}

func (r *Raft) GetParams() *util.ClusterParams {
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
	membership, err := r.nodeHost.GetClusterMembership(ctx, r.params.ClusterID)
	cancel()
	if err != nil {
		r.params.Logger.Error(err)
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
		if member.IsLeader && member.NodeID == r.config.NodeID && clusterID == r.params.ClusterID {
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
		r.params.Logger.Warningf("Raft cluster membership updated! info=%+v, clusterInfo=%+v",
			info, clusterInfo)
	}
}

func (r *Raft) GetNodeHost() *dragonboat.NodeHost {
	return r.nodeHost
}

func (r *Raft) SyncPropose(clusterID uint64, cmd []byte) error {
	r.sessionMutex.RLock()
	session, ok := r.session[clusterID]
	r.sessionMutex.RUnlock()
	if !ok {
		return common.ErrClusterNotFound
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	result, err := r.nodeHost.SyncPropose(ctx, session, cmd)
	cancel()
	if err != nil {
		r.params.Logger.Errorf("[Raft.SyncPropose] Error: %s", err)
		return err
	}
	r.params.Logger.Debugf("[Raft.SyncPropose] Raft confirmation: clusterID=%d, nodeID=%d, message=%s, result=%+v",
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
	requestState, err := r.nodeHost.ReadIndex(clusterID, r.timeout)
	if err != nil {
		return nil, err
	}
	select {
	case requestResult := <-requestState.ResultC():
		if requestResult.Completed() {
			return r.nodeHost.ReadLocalNode(requestState, query)
		}
		if requestResult.Aborted() {
			return nil, errors.New("read request aborted")
		}
		if requestResult.Dropped() {
			return nil, errors.New("read request dropped")
		}
		if requestResult.Rejected() {
			return nil, errors.New("read request rejected")
		}
		if requestResult.Terminated() {
			return nil, errors.New("read request terminated")
		}
		if requestResult.Timeout() {
			return nil, errors.New("read request timeout")
		}
	}
	return nil, errors.New("read request unknown error")

	//return r.nodeHost.StaleRead(clusterID, query)
}

// GetLeaderID returns the leader node ID of the specified Raft cluster based
// on local node's knowledge. The returned boolean value indicates whether the
// leader information is available.
func (r *Raft) GetLeaderID(clusterID uint64) (uint64, bool, error) {
	return r.nodeHost.GetLeaderID(clusterID)
}

func (r *Raft) Hash(key string) uint64 {
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(key))
	return clusterHash.Sum64()
}
