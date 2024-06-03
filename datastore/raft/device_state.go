//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/statemachine"
	"github.com/jeremyhahn/go-cropdroid/state"

	logging "github.com/op/go-logging"
)

type RaftDeviceStateStorer interface {
	CreateClusterNode(deviceID uint64, deviceType string,
		deviceStateChangeChan chan common.DeviceStateChange) error
	state.DeviceStateStorer
	//datastore.DeviceDataStore
}

type RaftDeviceStateStore struct {
	logger *logging.Logger
	raft   cluster.RaftNode
	RaftDeviceStateStorer
}

// Creates a new DeviceStateStore Raft DAO using a ConcurrentStateMachine to store
// all devices associated with a farm in memory.
func NewRaftDeviceStateStore(logger *logging.Logger, raftCluster cluster.RaftNode) RaftDeviceStateStorer {
	logger.Debug("Creating new RaftDeviceStateStore")
	return &RaftDeviceStateStore{
		logger: logger,
		raft:   raftCluster}
}

/* Raft cluster operational methods */

// Starts a device state Raft cluster on the current node
func (store *RaftDeviceStateStore) CreateClusterNode(deviceID uint64, deviceType string,
	deviceStateChangeChan chan common.DeviceStateChange) error {

	store.logger.Debugf("CreateClusterNode RaftDeviceStateStore with deviceID: %d, deviceType: %s", deviceID, deviceType)
	params := store.raft.GetParams()
	sm := statemachine.NewDeviceStateConcurrentStateMachine(store.logger, deviceID, deviceType, deviceStateChangeChan)
	if err := store.raft.CreateRegularCluster(deviceID, params.Join, sm.CreateDeviceStateConcurrentStateMachine); err != nil {
		return err
	}
	return nil
}

/* StateStore methods */

func (s *RaftDeviceStateStore) Len() int {
	return 1
}

// Puts a new device state entry into the Raft database
func (s *RaftDeviceStateStore) Put(clusterID uint64, v state.DeviceStateMap) error {
	s.logger.Debugf("Put *state.DeviceState for clusterID: %d, %+v", clusterID, v.(*state.DeviceState))
	data, err := json.Marshal(*v.(*state.DeviceState))
	if err != nil {
		s.logger.Errorf("Put RaftDeviceStateStore json.Marshall (clusterID=%d) error: %s", clusterID, err)
		return err
	}
	if err := s.raft.SyncPropose(clusterID, data); err != nil {
		s.logger.Errorf("Put RaftDeviceStateStore SyncPropose (clusterID=%d) error: %s", clusterID, err)
		return err
	}
	return nil
}

// Gets the current real-time state for the specified device
func (s *RaftDeviceStateStore) Get(clusterID uint64) (state.DeviceStateMap, error) {
	s.logger.Debugf("Get RaftDeviceStateStore for clusterID: %d", clusterID)
	result, err := s.raft.SyncRead(clusterID, nil)
	if err != nil {
		s.logger.Errorf("Get RaftDeviceStateStore (clusterID=%d) SyncRead error %s", clusterID, err)
		return nil, err
	}
	if result != nil {
		return result.(*state.DeviceState), nil
	}
	return nil, nil
}
