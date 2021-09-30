// +build cluster

package cluster

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/state"

	logging "github.com/op/go-logging"
)

type RaftDeviceStateStore struct {
	logger *logging.Logger
	raft   RaftCluster
	common.DeviceStore
}

func NewRaftDeviceStateStore(logger *logging.Logger, raftCluster RaftCluster) common.DeviceStore {
	return &RaftDeviceStateStore{
		logger: logger,
		raft:   raftCluster}
}

func (s *RaftDeviceStateStore) Len() int {
	return 1
}

// Puts a new device state entry into the Raft database
func (s *RaftDeviceStateStore) Put(clusterID uint64, v state.DeviceStateMap) error {
	data, err := json.Marshal(*v.(*state.DeviceState))
	if err != nil {
		s.logger.Errorf("[RaftDeviceStateStore.Put] Error: %s", err)
		return err
	}
	if err := s.raft.SyncPropose(clusterID, data); err != nil {
		s.logger.Errorf("[RaftDeviceStateStore.Put] Error: %s", err)
		return err
	}
	return nil
}

// Gets the current real-time state for the specified device
func (s *RaftDeviceStateStore) Get(clusterID uint64) (state.DeviceStateMap, error) {

	// Lookup method always returns an array
	// Lookup(nil) = get current state
	// Lookup("*") = get state history
	// Lookup("start:end") = get slice ranging from "start" to "end" (0 based index)
	result, err := s.raft.SyncRead(clusterID, nil)
	if err != nil {
		s.logger.Errorf("[RaftDeviceStateStore.Get] Error (clusterID=%d): %s", clusterID, err)
		return nil, err
	}

	if result != nil {
		records := result.([]state.DeviceStateMap)
		if len(records) > 0 {
			return records[0], nil
		}
		return nil, nil
	}
	return nil, nil
}

func (s *RaftDeviceStateStore) GetAll() []*state.DeviceStoreViewItem {
	s.logger.Errorf("Not implemented")
	return nil
}

/* Implements datastore.DeviceStore */

// Saves a new record for historical record keeping and reporting
func (s *RaftDeviceStateStore) Save(deviceID uint64, deviceState state.DeviceStateMap) error {
	// Do nothing, the record was already saved to the raft database
	// when the state store called Put() above.
	return nil
}

// Returns all records for the given metric within the last 30 days
func (s *RaftDeviceStateStore) GetLast30Days(deviceID uint64, metric string) ([]float64, error) {
	// Lookup method always returns an array
	// Lookup(nil) = get current state
	// Lookup("*") = get state history
	// Lookup("start:end") = get slice ranging from "start" to "end" (0 based index)
	// result, err := s.raft.SyncRead(deviceID, "*")
	// if err != nil {
	// 	s.logger.Errorf("Error (clusterID=%d): %s", deviceID, err)
	// 	return nil, err
	// }

	// if result != nil {
	// 	records := result.([]state.DeviceStateMap)
	// 	if len(records) > 0 {
	// 		for _, r := range records {
	// 			if r == nil {
	// 				continue
	// 			}
	// 			s.logger.Errorf("device.id=%d, record: %+v", deviceID, r)
	// 			for _, d := range r.GetDevices() {
	// 				s.logger.Debugf("Found record for device.id=%d, device: %+v", deviceID, d)
	// 			}
	// 		}
	// 		return records[0], nil
	// 	}
	// 	return nil, nil
	// }
	// return nil, nil

	var floats []float64
	return floats, nil
}
