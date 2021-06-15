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

type RaftDeviceConfigStore struct {
	logger       *logging.Logger
	raft         RaftCluster
	cachedConfig config.DeviceConfig
	mutex        *sync.RWMutex
	store.DeviceConfigStorer
}

func NewRaftDeviceConfigStore(logger *logging.Logger, raftCluster RaftCluster) store.DeviceConfigStorer {
	return &RaftDeviceConfigStore{
		logger: logger,
		raft:   raftCluster,
		mutex:  &sync.RWMutex{}}
}

func (s *RaftDeviceConfigStore) Len() int {
	return 1
}

func (s *RaftDeviceConfigStore) Put(clusterID uint64, v config.DeviceConfig) error {
	s.logger.Debugf("Setting configuration for cluster %d. config=%+v", clusterID, v)
	data, err := json.Marshal(*v.(*config.Device))
	if err != nil {
		s.logger.Errorf("Error: %s", err)
		return err
	}
	if err := s.raft.SyncPropose(clusterID, data); err != nil {
		s.logger.Errorf("Error: %s", err)
		return err
	}
	s.mutex.RLock()
	s.cachedConfig = v
	s.mutex.RUnlock()
	return nil
}

func (s *RaftDeviceConfigStore) Get(clusterID uint64, CONSISTENCY_LEVEL int) (config.DeviceConfig, error) {

	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_CACHED {
		result = []config.DeviceConfig{s.cachedConfig}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = s.raft.ReadLocal(clusterID, nil)
		if err != nil {
			s.logger.Errorf("Error (clusterID=%d): %s", clusterID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = s.raft.SyncRead(clusterID, nil)
		if err != nil {
			s.logger.Errorf("Error (clusterID=%d): %s", clusterID, err)
			return nil, err
		}
	}

	// Lookup method always returns an array
	// Lookup(nil) = get current state
	// Lookup("*") = get state history
	// Lookup("start:end") = get slice ranging from "start" to "end" (0 based index)

	if result != nil {
		records := result.([]config.DeviceConfig)
		if len(records) > 0 {
			/*for _, r := range records {
				if r == nil {
					continue
				}
				store.logger.Errorf("farm.id=%d, record: %+v", clusterID, r)
				for _, c := range r.GetDevices() {
					store.logger.Errorf("farm.id=%d, device: %+v", clusterID, c)
				}
			}*/
			return records[0], nil
		}
		return nil, nil
	}
	return nil, nil
}

func (store *RaftDeviceConfigStore) GetAll(farmID uint64) []config.DeviceConfig {
	store.logger.Errorf("RaftDeviceConfigStore.GetAll() Not implemented")
	return nil
}
