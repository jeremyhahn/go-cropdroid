//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/common"
	fs "github.com/jeremyhahn/go-cropdroid/state"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type DeviceDataStateMachine interface {
	CreateDeviceDataStateMachine(clusterID, nodeID uint64) sm.IStateMachine
	sm.IStateMachine
}

type DeviceDataSM struct {
	logger                *logging.Logger
	clusterID             uint64
	nodeID                uint64
	deviceID              uint64
	deviceType            string
	current               fs.DeviceStateMap
	history               []fs.DeviceStateMap
	deviceStateChangeChan chan common.DeviceStateChange
	mutex                 *sync.RWMutex
	sm.IConcurrentStateMachine
	fs.DeviceStore
}

func NewDeviceDataStateMachine(logger *logging.Logger,
	deviceID uint64, deviceType string,
	deviceStateChangeChan chan common.DeviceStateChange) DeviceDataStateMachine {

	return &DeviceDataSM{
		logger:                logger,
		deviceID:              deviceID,
		deviceType:            deviceType,
		history:               make([]fs.DeviceStateMap, 0),
		deviceStateChangeChan: deviceStateChangeChan,
		mutex:                 &sync.RWMutex{}}
}

func (s *DeviceDataSM) CreateDeviceDataStateMachine(clusterID, nodeID uint64) sm.IStateMachine {
	s.clusterID = clusterID
	s.nodeID = nodeID
	s.mutex = &sync.RWMutex{}
	return s
}

func (s *DeviceDataSM) Lookup(query interface{}) (interface{}, error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Warningf("[DeviceDataStateMachine.Lookup] query: %+v", query)
	s.logger.Warningf("[DeviceDataStateMachine.Lookup] Current: %+v", s.current)
	//s.logger.Warningf("[DeviceDataStateMachine.Lookup] History: %+v", s.history)

	if query == nil {
		return []fs.DeviceStateMap{s.current}, nil
	}
	if query.(string) == "*" {
		return s.history, nil
	}
	pieces := strings.Split(query.(string), ":")
	start, err := strconv.Atoi(pieces[0])
	if err != nil {
		return nil, err
	}
	end, err := strconv.Atoi(pieces[1])
	if err != nil {
		return nil, err
	}
	return s.history[start:end], nil
}

func (s *DeviceDataSM) Update(data []byte) (sm.Result, error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var deviceState fs.DeviceState
	err := json.Unmarshal(data, &deviceState)
	if err != nil {
		s.logger.Errorf("[DeviceDataStateMachine.Update] Error: %s\n", err)
		return sm.Result{}, err
	}
	if s.current != nil {
		s.history = append(s.history, s.current)
	}
	s.current = &deviceState

	s.logger.Debugf("[DeviceDataStateMachine.Update] device.id: %d, store.history.len: %d, device: %+v\n",
		s.deviceID, len(s.history), string(data))

	// s.deviceStateChangeChan <- &deviceState

	s.deviceStateChangeChan <- common.DeviceStateChange{
		DeviceID:   s.deviceID,
		DeviceType: s.deviceType,
		StateMap:   &deviceState}

	return sm.Result{Value: s.deviceID, Data: data}, nil
}

// SaveSnapshot saves the current IStateMachine state into a snapshot using the
// specified io.Writer object.
func (s *DeviceDataSM) SaveSnapshot(w io.Writer, fc sm.ISnapshotFileCollection, done <-chan struct{}) error {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	snap := s.history
	if s.current != nil {
		snap = append(snap, s.current)
	}
	bytes, err := json.Marshal(snap)
	if err != nil {
		s.logger.Errorf("[DeviceDataStateMachine.SaveSnapshot] Error: %s", err)
		return err
	}
	s.logger.Infof("[DeviceDataStateMachine.SaveSnapshot] Created new snaphot. History length: %d", len(snap))
	_, err = w.Write(bytes)
	return err
}

// RecoverFromSnapshot recovers the state using the provided snapshot.
func (s *DeviceDataSM) RecoverFromSnapshot(r io.Reader, files []sm.SnapshotFile, done <-chan struct{}) error {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var deviceState []fs.DeviceState
	err = json.Unmarshal(data, &deviceState)
	if err != nil {
		s.logger.Errorf("[DeviceDataStateMachine.RecoverFromSnapshot] Error: %s, deviceState: %+v\n", err, deviceState)
		return err
	}
	s.logger.Debugf("[DeviceDataStateMachine.SaveSnapshot] Recovered from snapshot. History length: %d", len(deviceState))
	if len(deviceState) > 0 {
		s.history = make([]fs.DeviceStateMap, len(deviceState))
		for i, h := range deviceState {
			s.history[i] = &h
		}
		s.current = s.history[len(deviceState)-1]
	}
	return nil
}

// Close closes the IStateMachine instance. There is nothing for us to cleanup
// or release as this is a pure in memory data store. Note that the Close
// method is not guaranteed to be called as node can crash at any time.
func (s *DeviceDataSM) Close() error { return nil }

// GetHash returns a uint64 representing the current object state.
func (s *DeviceDataSM) GetHash() (uint64, error) {
	return uint64(s.current.GetTimestamp().Unix()), nil
}
