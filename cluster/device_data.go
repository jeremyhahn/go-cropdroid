//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/state"

	logging "github.com/op/go-logging"
)

type RaftDeviceDataStore struct {
	logger *logging.Logger
	raft   RaftNode
	datastore.DeviceDataStore
}

func NewRaftDeviceDataStore(logger *logging.Logger, raftCluster RaftNode) *RaftDeviceDataStore {
	return &RaftDeviceDataStore{
		logger: logger,
		raft:   raftCluster}
}

// Puts a new device state entry into the Raft database
func (s *RaftDeviceDataStore) Save(clusterID uint64, v state.DeviceStateMap) error {
	data, err := json.Marshal(*v.(*state.DeviceState))
	if err != nil {
		s.logger.Errorf("[RaftDeviceDataStore.Put] Error: %s", err)
		return err
	}
	if err := s.raft.SyncPropose(clusterID, data); err != nil {
		s.logger.Errorf("[RaftDeviceDataStore.Put] Error: %s", err)
		return err
	}
	return nil
}

// Returns all records for the given metric within the last 30 days
func (s *RaftDeviceDataStore) GetLast30Days(deviceID uint64, metric string) ([]float64, error) {
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
