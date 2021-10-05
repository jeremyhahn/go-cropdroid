// +build cluster

package statemachine

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"

	fs "github.com/jeremyhahn/go-cropdroid/state"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type FarmStateMachine interface {
	CreateFarmStateMachine(clusterID, nodeID uint64) sm.IConcurrentStateMachine
	sm.IConcurrentStateMachine
}

type FarmSM struct {
	logger              *logging.Logger
	clusterID           uint64
	nodeID              uint64
	farmID              uint64
	data                []fs.FarmStateMap
	dataMutex           *sync.RWMutex
	snapshotIdentifier  uint64
	snapshot            []fs.FarmStateMap
	snapshotMutex       *sync.RWMutex
	farmStateChangeChan chan fs.FarmStateMap
	FarmStateMachine
	fs.FarmStore
}

// This is the farm state machine the Dragonboat library uses to
// maintain farm state for a Raft node.
// https://github.com/lni/dragonboat/blob/master/statemachine/concurrent.go
func NewFarmStateMachine(logger *logging.Logger, farmID uint64,
	farmStateChangeChan chan fs.FarmStateMap) FarmStateMachine {

	return &FarmSM{
		logger:              logger,
		farmID:              farmID,
		data:                make([]fs.FarmStateMap, 0),
		farmStateChangeChan: farmStateChangeChan,
		dataMutex:           &sync.RWMutex{},
		snapshotMutex:       &sync.RWMutex{}}
}

func (s *FarmSM) CreateFarmStateMachine(clusterID, nodeID uint64) sm.IConcurrentStateMachine {
	s.clusterID = clusterID
	s.nodeID = nodeID
	s.dataMutex = &sync.RWMutex{}
	s.snapshotMutex = &sync.RWMutex{}
	return s
}

func (s *FarmSM) Lookup(query interface{}) (interface{}, error) {

	s.logger.Debugf("[FarmStateMachine.Lookup] query: %+v", query)

	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	dataLen := len(s.data)
	if dataLen == 0 {
		return nil, nil
	}

	dataIdx := dataLen - 1

	s.logger.Debugf("[FarmStateMachine.Lookup] dataLen: %d", dataLen)
	s.logger.Debugf("[FarmStateMachine.Lookup] dataIdx: %d", dataIdx)

	current := s.data[dataIdx]

	s.logger.Debugf("[FarmStateMachine.Lookup] current state: %+v", current)

	if query == nil {
		return []fs.FarmStateMap{current}, nil
	}
	if query.(string) == "*" {
		return s.data, nil
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
	return s.data[start:end], nil
}

func (s *FarmSM) Update(entries []sm.Entry) ([]sm.Entry, error) {

	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	dataLen := len(entries)
	s.data = make([]fs.FarmStateMap, dataLen, dataLen)

	for i, entry := range entries {
		var farmState fs.FarmState
		err := json.Unmarshal(entry.Cmd, &farmState)
		if err != nil {
			s.logger.Errorf("[FarmStateMachine.Update] Error: %s\n", err)
			return entries, err
		}
		s.data[i] = &farmState

		s.logger.Debugf("[FarmStateMachine.Update] farm.id: %d, store.data.len: %d, farm: %+v\n",
			s.farmID, dataLen, string(entry.Cmd))

		entry.Result = sm.Result{
			Value: uint64(1),
			Data:  nil}
	}
	s.farmStateChangeChan <- s.data[dataLen-1]
	return entries, nil
}

func (s *FarmSM) PrepareSnapshot() (interface{}, error) {

	s.snapshotMutex.Lock()
	defer s.snapshotMutex.Unlock()

	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	s.snapshotIdentifier++

	dataLen := len(s.data)
	s.snapshot = make([]fs.FarmStateMap, dataLen, dataLen)
	copy(s.snapshot, s.data)

	return s.snapshotIdentifier, nil
}

// SaveSnapshot saves the current IStateMachine state into a snapshot using the
// specified io.Writer object.
func (s *FarmSM) SaveSnapshot(stateIdentifier interface{}, w io.Writer,
	fc sm.ISnapshotFileCollection, done <-chan struct{}) error {

	s.snapshotMutex.Lock()
	defer s.snapshotMutex.Unlock()

	if stateIdentifier != s.snapshotIdentifier {
		err := fmt.Errorf("Farm state machine snapshot identifier mismatch! expected %d got %d",
			s.snapshotIdentifier, stateIdentifier)
		s.logger.Errorf("[FarmStateMachine.SaveSnapshot] %s", err)
		return err
	}
	bytes, err := json.Marshal(s.snapshot)
	if err != nil {
		s.logger.Errorf("[FarmStateMachine.SaveSnapshot] Error: %s", err)
		return err
	}
	s.logger.Infof("[FarmStateMachine.SaveSnapshot] Created new snaphot. History length: %d",
		len(s.snapshot))
	_, err = w.Write(bytes)
	s.snapshot = nil
	return err
}

// RecoverFromSnapshot recovers the state using the provided snapshot.
func (s *FarmSM) RecoverFromSnapshot(r io.Reader, files []sm.SnapshotFile, done <-chan struct{}) error {

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var farmState []fs.FarmState
	err = json.Unmarshal(data, &farmState)
	if err != nil {
		s.logger.Errorf("[FarmStateMachine.RecoverFromSnapshot] Error: %s, farmState: %+v\n", err, farmState)
		return err
	}
	s.logger.Debugf("[FarmStateMachine.SaveSnapshot] Recovered from snapshot. History length: %d", len(farmState))

	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	if len(farmState) > 0 {
		s.data = make([]fs.FarmStateMap, len(farmState))
		for i, h := range farmState {
			s.data[i] = &h
		}
	}
	return nil
}

// Close closes the IStateMachine instance. There is nothing for us to cleanup
// or release as this is a pure in memory data store. Note that the Close
// method is not guaranteed to be called as node can crash at any time.
func (s *FarmSM) Close() error { return nil }

// GetHash returns a uint64 representing the current object state.
func (s *FarmSM) GetHash() (uint64, error) {
	dataLen := len(s.data)
	if dataLen == 0 {
		return 0, nil
	}
	return uint64(s.data[dataLen-1].GetTimestamp()), nil
}
