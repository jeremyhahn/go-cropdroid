//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/statemachine"
	"github.com/jeremyhahn/go-cropdroid/state"

	logging "github.com/op/go-logging"
)

type RaftFarmStateStorer interface {
	ClusterID() uint64
	RaftCluster
	state.FarmStorer
}

type RaftFarmStateStore struct {
	logger              *logging.Logger
	raft                cluster.RaftNode
	clusterID           uint64
	farmID              uint64
	farmStateChangeChan chan state.FarmStateMap
	RaftFarmStateStorer
}

// Creates a new FarmStateStore Raft DAO using a ConcurrentStateMachine to store
// a single FarmStateMap in memory.
func NewRaftFarmStateStore(logger *logging.Logger, raftNode cluster.RaftNode,
	farmID uint64, farmStateChangeChan chan state.FarmStateMap) RaftFarmStateStorer {

	farmStateID := raftNode.GetParams().IdGenerator.NewFarmStateID(farmID)
	logger.Debugf("Creating new *state.FarmState for farmID: %d, farmStateID: %d", farmID, farmStateID)
	return &RaftFarmStateStore{
		logger:              logger,
		raft:                raftNode,
		farmID:              farmID,
		clusterID:           farmStateID,
		farmStateChangeChan: farmStateChangeChan}
}

func (store *RaftFarmStateStore) ClusterID() uint64 {
	return store.clusterID
}

/* Raft cluster operational methods */

// Starts a farm state Raft cluster on the current node using and in-memory ConcurrentStateMachine
func (store *RaftFarmStateStore) StartClusterNode(waitForClusterReady bool) error {
	store.logger.Debugf("Creating *state.FarmState raft cluster id %d on node %d", store.clusterID, store.raft.GetConfig().NodeID)
	params := store.raft.GetParams()
	sm := statemachine.NewFarmStateConcurrentStateMachine(store.logger, store.clusterID, store.farmStateChangeChan)
	err := store.raft.CreateConcurrentCluster(store.clusterID, params.Join, sm.CreateFarmStateConcurrentStateMachine)
	if err != nil {
		store.logger.Fatal(err)
		return err
	}
	if waitForClusterReady {
		store.raft.WaitForClusterReady(store.clusterID)
	}
	return nil
}

// Starts a local multi-node farm state Raft cluster using and in-memory ConcurrentStateMachine
func (store *RaftFarmStateStore) StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error {
	localCluster.app.Logger.Debugf("Creating local %d node raft cluster *state.FarmState: %d",
		localCluster.nodeCount, store.clusterID)
	clusterID := uint64(0)
	for i := 0; i < localCluster.nodeCount; i++ {
		raftNode := localCluster.GetRaftNode(i)
		farmStateStore := NewRaftFarmStateStore(store.logger, raftNode, store.farmID, store.farmStateChangeChan)
		err := farmStateStore.StartClusterNode(false)
		if err != nil {
			store.logger.Fatal(err)
		}
		clusterID = farmStateStore.ClusterID()
	}
	if waitForClusterReady {
		store.raft.WaitForClusterReady(clusterID)
	}
	return nil
}

func (store *RaftFarmStateStore) WaitForClusterReady() {
	store.raft.WaitForClusterReady(store.clusterID)
}

/* StateStore methods */

func (store *RaftFarmStateStore) Len() int {
	return 1
}

func (store *RaftFarmStateStore) Put(farmStateID uint64, v state.FarmStateMap) error {
	store.logger.Debugf("Put *state.FarmState for farmStateID: %d", farmStateID)
	data, err := json.Marshal(*v.(*state.FarmState))
	if err != nil {
		store.logger.Errorf("[RaftFarmStateStore.Put] Error: %s", err)
		return err
	}
	if err := store.raft.SyncPropose(farmStateID, data); err != nil {
		store.logger.Errorf("[RaftFarmStateStore.Put] Error: %s", err)
		return err
	}
	return nil
}

func (store *RaftFarmStateStore) Get(farmStateID uint64) (state.FarmStateMap, error) {
	store.logger.Debugf("Get *state.FarmState for farmStateID: %d", farmStateID)
	//result, err := store.raft.SyncRead(farmStateID, nil)
	result, err := store.raft.ReadLocal(farmStateID, nil)
	if err != nil {
		store.logger.Errorf("[RaftFarmStateStore.Get] Error (farmStateID=%d): %s",
			farmStateID, err)
		return nil, err
	}
	if result != nil {
		return result.(state.FarmStateMap), nil
	}
	return nil, nil
}
