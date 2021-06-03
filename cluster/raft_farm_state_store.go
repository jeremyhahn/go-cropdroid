// +build cluster

package cluster

import (
	"encoding/json"

	"github.com/jeremyhahn/cropdroid/state"

	logging "github.com/op/go-logging"
)

type RaftFarmStateStore struct {
	logger *logging.Logger
	raft   RaftCluster
	state.FarmStorer
}

func NewRaftFarmStateStore(logger *logging.Logger, raftCluster RaftCluster) state.FarmStorer {
	return &RaftFarmStateStore{
		logger: logger,
		raft:   raftCluster}
}

func (store *RaftFarmStateStore) Len() int {
	return 1
}

func (store *RaftFarmStateStore) Put(farmID int, v state.FarmStateMap) error {
	data, err := json.Marshal(*v.(*state.FarmState))
	if err != nil {
		store.logger.Errorf("[RaftFarmStateStore.Put] Error: %s", err)
		return err
	}
	if err := store.raft.SyncPropose(uint64(farmID), data); err != nil {
		store.logger.Errorf("[RaftFarmStateStore.Put] Error: %s", err)
		return err
	}
	return nil
}

func (store *RaftFarmStateStore) Get(farmID int) (state.FarmStateMap, error) {
	result, err := store.raft.SyncRead(uint64(farmID), nil)
	if err != nil {
		store.logger.Errorf("[RaftFarmStateStore.Get] Error (clusterID=%d): %s", farmID, err)
		return nil, err
	}

	// Lookup method always returns an array
	// Lookup(nil) = get current state
	// Lookup("*") = get state history
	// Lookup("start:end") = get slice ranging from "start" to "end" (0 based index)

	if result != nil {
		records := result.([]state.FarmStateMap)
		if len(records) > 0 {
			/*for _, r := range records {
				if r == nil {
					continue
				}
				store.logger.Errorf("[RaftFarmStateStore.Get] farm.id=%d, record: %+v", farmID, r)
				for _, c := range r.GetControllers() {
					store.logger.Errorf("[RaftFarmStateStore.Get] farm.id=%d, controller: %+v", farmID, c)
				}
			}*/
			return records[0], nil
		}
		return nil, nil
	}
	return nil, nil
}

func (store *RaftFarmStateStore) GetAll() []*state.StoreViewItem {
	store.logger.Errorf("RaftFarmStateStore.GetAll() Not implemented")
	return nil
}
