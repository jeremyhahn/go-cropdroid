//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/statemachine"
	"github.com/jeremyhahn/go-cropdroid/state"

	logging "github.com/op/go-logging"
)

type RaftDeviceDataDAO struct {
	logger *logging.Logger
	raft   cluster.RaftNode
	datastore.DeviceDataStore
}

func NewRaftDeviceDataDAO(logger *logging.Logger, raftCluster cluster.RaftNode) *RaftDeviceDataDAO {
	return &RaftDeviceDataDAO{
		logger: logger,
		raft:   raftCluster}
}

// Puts a new device state entry into the Raft database
func (s *RaftDeviceDataDAO) Save(deviceID uint64, v state.DeviceStateMap) error {
	data, err := json.Marshal(v)
	if err != nil {
		s.logger.Errorf("[RaftDeviceDataDAO.Put] Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, data).Serialize()
	if err != nil {
		s.logger.Errorf("Error: %s", err)
		return err
	}
	deviceDataClusterID := s.raft.GetParams().IdGenerator.CreateDeviceDataClusterID(deviceID)
	if err := s.raft.SyncPropose(deviceDataClusterID, proposal); err != nil {
		s.logger.Errorf("[RaftDeviceDataStore.Put] Error: %s", err)
		return err
	}
	return nil
}

// Returns all records for the given metric within the last 30 days
func (s *RaftDeviceDataDAO) GetLast30Days(deviceID uint64, metric string) ([]float64, error) {
	// Lookup method always returns an array
	// Lookup(nil) = get current state
	// Lookup("*") = get state history
	// Lookup("start:end") = get slice ranging from "start" to "end" (0 based index)
	deviceDataClusterID := s.raft.GetParams().IdGenerator.CreateDeviceDataClusterID(deviceID)
	result, err := s.raft.SyncRead(deviceDataClusterID, statemachine.QUERY_TYPE_WILDCARD)
	if err != nil {
		s.logger.Errorf("Error (deviceDataClusterID=%d): %s", deviceDataClusterID, err)
		return nil, err
	}
	var resultSet []float64
	if result != nil {
		records := result.([]state.DeviceState)
		resultSet = make([]float64, len(records))
		if len(records) > 0 {
			for i, record := range records {
				val, exists := record.GetMetrics()[metric]
				if !exists {
					return resultSet, datastore.ErrMetricKeyNotFound
				}
				resultSet[i] = val
			}
		}
		return resultSet, nil
	}
	return resultSet, nil
}
