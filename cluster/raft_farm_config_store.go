// +build cluster

package cluster

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"

	logging "github.com/op/go-logging"
)

type RaftFarmConfigStore struct {
	logger *logging.Logger
	raft   RaftCluster
	state.ConfigStorer
}

func NewRaftFarmConfigStore(logger *logging.Logger, raftCluster RaftCluster) state.ConfigStorer {
	return &RaftFarmConfigStore{
		logger: logger,
		raft:   raftCluster}
}

func (store *RaftFarmConfigStore) Len() int {
	return 1
}

func (store *RaftFarmConfigStore) Put(clusterID uint64, v config.FarmConfig) error {
	store.logger.Debugf("[RaftFarmConfigStore.Put] Setting configuration for cluster %d. config=%+v", clusterID, v)
	data, err := json.Marshal(*v.(*config.Farm))
	if err != nil {
		store.logger.Errorf("[RaftFarmConfigStore.Put] Error: %s", err)
		return err
	}
	if err := store.raft.SyncPropose(clusterID, data); err != nil {
		store.logger.Errorf("[RaftFarmConfigStore.Put] Error: %s", err)
		return err
	}
	return nil
}

func (store *RaftFarmConfigStore) Get(clusterID uint64) (config.FarmConfig, error) {
	result, err := store.raft.SyncRead(clusterID, nil)
	if err != nil {
		store.logger.Errorf("[RaftFarmConfigStore.Get] Error (clusterID=%d): %s", clusterID, err)
		return nil, err
	}

	// Lookup method always returns an array
	// Lookup(nil) = get current state
	// Lookup("*") = get state history
	// Lookup("start:end") = get slice ranging from "start" to "end" (0 based index)

	if result != nil {
		records := result.([]config.FarmConfig)
		if len(records) > 0 {
			/*for _, r := range records {
				if r == nil {
					continue
				}
				store.logger.Errorf("[RaftFarmConfigStore.Get] farm.id=%d, record: %+v", clusterID, r)
				for _, c := range r.GetControllers() {
					store.logger.Errorf("[RaftFarmConfigStore.Get] farm.id=%d, controller: %+v", clusterID, c)
				}
			}*/
			return records[0], nil
		}
		return nil, nil
	}
	return nil, nil
}

func (store *RaftFarmConfigStore) GetAll() []config.FarmConfig {
	store.logger.Errorf("RaftFarmConfigStore.GetAll() Not implemented")
	return nil
}
