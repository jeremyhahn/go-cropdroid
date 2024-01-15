//go:build cluster && pebble
// +build cluster,pebble

package statemachine

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"sync"

// 	"github.com/jeremyhahn/go-cropdroid/common"
// 	fs "github.com/jeremyhahn/go-cropdroid/state"
// 	sm "github.com/lni/dragonboat/v3/statemachine"
// 	logging "github.com/op/go-logging"
// )

// type ConcurrentDeviceStateMachine interface {
// 	CreateConcurrentDeviceStateMachine(clusterID, nodeID uint64) sm.IConcurrentStateMachine
// 	sm.IConcurrentStateMachine
// }

// type ConcurrentDeviceSM struct {
// 	logger                *logging.Logger
// 	clusterID             uint64
// 	nodeID                uint64
// 	deviceID              uint64
// 	deviceType            string
// 	current               fs.DeviceStateMap
// 	deviceStateChangeChan chan common.DeviceStateChange
// 	mutex                 *sync.RWMutex
// 	snapshotIdentifier    uint64
// 	snapshot              fs.FarmStateMap
// 	snapshotMutex         *sync.RWMutex
// 	sm.IConcurrentStateMachine
// 	fs.DeviceStore
// }

// func NewConcurrentDeviceStateMachine(logger *logging.Logger,
// 	deviceID uint64, deviceType string,
// 	deviceStateChangeChan chan common.DeviceStateChange) ConcurrentDeviceStateMachine {

// 	return &ConcurrentDeviceSM{
// 		logger:                logger,
// 		deviceID:              deviceID,
// 		deviceType:            deviceType,
// 		deviceStateChangeChan: deviceStateChangeChan,
// 		mutex:                 &sync.RWMutex{},
// 		snapshotMutex:         &sync.RWMutex{}}
// }

// func (s *ConcurrentDeviceSM) CreateConcurrentDeviceStateMachine(
// 	clusterID, nodeID uint64) sm.IConcurrentStateMachine {

// 	s.clusterID = clusterID
// 	s.nodeID = nodeID
// 	s.snapshotMutex = &sync.RWMutex{}
// 	s.mutex = &sync.RWMutex{}
// 	return s
// }

// func (s *ConcurrentDeviceSM) Lookup(query interface{}) (interface{}, error) {

// 	s.mutex.RLock()
// 	defer s.mutex.RUnlock()

// 	s.logger.Warningf("[DeviceStateMachine.Lookup] query: %+v", query)
// 	s.logger.Warningf("[DeviceStateMachine.Lookup] Current: %+v", s.current)
// 	//s.logger.Warningf("[DeviceStateMachine.Lookup] History: %+v", s.history)

// 	return s.current, nil

// 	// if query == nil {
// 	// 	return []fs.DeviceStateMap{s.current}, nil
// 	// }
// 	// if query.(string) == "*" {
// 	// 	return s.history, nil
// 	// }
// 	// pieces := strings.Split(query.(string), ":")
// 	// start, err := strconv.Atoi(pieces[0])
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// end, err := strconv.Atoi(pieces[1])
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// return s.history[start:end], nil
// }

// func (s *ConcurrentDeviceSM) Update(entries []sm.Entry) ([]sm.Entry, error) {

// 	var deviceState fs.DeviceState
// 	err := json.Unmarshal(data, &deviceState)
// 	if err != nil {
// 		s.logger.Errorf("[DeviceStateMachine.Update] Error: %s\n", err)
// 		return sm.Result{}, err
// 	}

// 	s.mutex.Lock()
// 	s.current = &deviceState
// 	s.mutex.Unlock()

// 	s.logger.Debugf("[DeviceStateMachine.Update] device.id: %d, device: %+v\n",
// 		s.deviceID, string(data))

// 	s.deviceStateChangeChan <- common.DeviceStateChange{
// 		DeviceID:   s.deviceID,
// 		DeviceType: s.deviceType,
// 		StateMap:   &deviceState}

// 	return sm.Result{Value: s.deviceID, Data: data}, nil
// }

// // SaveSnapshot saves the current IStateMachine state into a snapshot using the
// // specified io.Writer object.
// func (s *ConcurrentDeviceSM) SaveSnapshot(stateIdentifier interface{}, w io.Writer,
// 	fc sm.ISnapshotFileCollection, done <-chan struct{}) error {

// 	s.snapshotMutex.Lock()
// 	defer s.snapshotMutex.Unlock()

// 	if stateIdentifier != s.snapshotIdentifier {
// 		err := fmt.Errorf("Farm state machine snapshot identifier mismatch! expected %d got %d",
// 			s.snapshotIdentifier, stateIdentifier)
// 		s.logger.Errorf("[FarmStateMachine.SaveSnapshot] %s", err)
// 		return err
// 	}
// 	bytes, err := json.Marshal(s.snapshot)
// 	if err != nil {
// 		s.logger.Errorf("[FarmStateMachine.SaveSnapshot] Error: %s", err)
// 		return err
// 	}
// 	s.logger.Infof("[FarmStateMachine.SaveSnapshot] Created new snaphot")
// 	_, err = w.Write(bytes)
// 	s.snapshot = nil
// 	return err
// }

// // RecoverFromSnapshot recovers the state using the provided snapshot.
// func (s *ConcurrentDeviceSM) RecoverFromSnapshot(r io.Reader, files []sm.SnapshotFile, done <-chan struct{}) error {

// 	data, err := ioutil.ReadAll(r)
// 	if err != nil {
// 		return err
// 	}

// 	var deviceState fs.DeviceStateMap
// 	err = json.Unmarshal(data, &deviceState)
// 	if err != nil {
// 		s.logger.Errorf("[DeviceStateMachine.RecoverFromSnapshot] Error: %s, deviceState: %+v\n", err, deviceState)
// 		return err
// 	}

// 	s.mutex.Lock()
// 	s.current = deviceState
// 	s.mutex.Unlock()

// 	s.logger.Debugf("[DeviceStateMachine.SaveSnapshot] Recovered from snapshot: %+v", deviceState)
// 	return nil
// }

// // Close closes the IStateMachine instance. There is nothing for us to cleanup
// // or release as this is a pure in memory data store. Note that the Close
// // method is not guaranteed to be called as node can crash at any time.
// func (s *ConcurrentDeviceSM) Close() error { return nil }

// // GetHash returns a uint64 representing the current object state.
// func (s *ConcurrentDeviceSM) GetHash() (uint64, error) {
// 	return uint64(s.current.GetTimestamp().Unix()), nil
// }
