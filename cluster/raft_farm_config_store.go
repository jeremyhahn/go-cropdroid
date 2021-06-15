// +build cluster

package cluster

import (
	"encoding/json"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/store"

	logging "github.com/op/go-logging"
)

type RaftFarmConfigStore struct {
	logger       *logging.Logger
	raft         RaftCluster
	cachedConfig config.FarmConfig
	mutex        *sync.RWMutex
	store.FarmConfigStorer
}

func NewRaftFarmConfigStore(logger *logging.Logger, raftCluster RaftCluster) store.FarmConfigStorer {
	return &RaftFarmConfigStore{
		logger: logger,
		raft:   raftCluster,
		mutex:  &sync.RWMutex{}}
}

func (s *RaftFarmConfigStore) Len() int {
	return 1
}

func (s *RaftFarmConfigStore) Put(clusterID uint64, v config.FarmConfig) error {
	s.logger.Debugf("[RaftFarmConfigStore.Put] Setting configuration for cluster %d. config=%+v", clusterID, v)
	data, err := json.Marshal(*v.(*config.Farm))
	if err != nil {
		s.logger.Errorf("[RaftFarmConfigStore.Put] Error: %s", err)
		return err
	}
	if err := s.raft.SyncPropose(clusterID, data); err != nil {
		s.logger.Errorf("[RaftFarmConfigStore.Put] Error: %s", err)
		return err
	}
	s.mutex.RLock()
	s.cachedConfig = v
	s.mutex.RUnlock()
	return nil
}

func (s *RaftFarmConfigStore) Get(clusterID uint64, CONSISTENCY_LEVEL int) (config.FarmConfig, error) {

	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_CACHED {
		result = []config.FarmConfig{s.cachedConfig}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = s.raft.ReadLocal(clusterID, nil)
		if err != nil {
			s.logger.Errorf("[RaftFarmConfigStore.Get] Error (clusterID=%d): %s", clusterID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = s.raft.SyncRead(clusterID, nil)
		if err != nil {
			s.logger.Errorf("[RaftFarmConfigStore.Get] Error (clusterID=%d): %s", clusterID, err)
			return nil, err
		}
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
				for _, c := range r.GetDevices() {
					store.logger.Errorf("[RaftFarmConfigStore.Get] farm.id=%d, device: %+v", clusterID, c)
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
