//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/state"

	logging "github.com/op/go-logging"
)

type RaftFarmStateStore struct {
	logger *logging.Logger
	raft   cluster.RaftNode
	state.FarmStorer
}

func NewRaftFarmStateStore(logger *logging.Logger,
	raftNode cluster.RaftNode) state.FarmStorer {
	return &RaftFarmStateStore{
		logger: logger,
		raft:   raftNode}
}

func (store *RaftFarmStateStore) Len() int {
	return 1
}

func (store *RaftFarmStateStore) Put(farmStateID uint64, v state.FarmStateMap) error {
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
	//result, err := store.raft.SyncRead(farmStateID, nil)
	result, err := store.raft.ReadLocal(farmStateID, nil)
	if err != nil {
		store.logger.Errorf("[RaftFarmStateStore.Get] Error (farmStateID=%d): %s",
			farmStateID, err)
		return nil, err
	}
	// Lookup method always returns an array
	// Lookup(nil) = get current state
	// Lookup("*") = get state history
	// Lookup("start:end") = get slice ranging from "start" to "end" (0 based index)
	if result != nil {
		return result.(state.FarmStateMap), nil
	}
	return nil, nil
}

func (store *RaftFarmStateStore) GetAll() []*state.StoreViewItem {
	store.logger.Errorf("RaftFarmStateStore.GetAll() Not implemented")
	return nil
}
