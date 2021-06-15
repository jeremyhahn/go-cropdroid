// +build cluster

package state

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	fs "github.com/jeremyhahn/go-cropdroid/state"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type FarmStateMachine interface {
	CreateStateMachine(clusterID, nodeID uint64) sm.IStateMachine
	sm.IStateMachine
}

type FarmSM struct {
	logger              *logging.Logger
	clusterID           uint64
	nodeID              uint64
	farmID              uint64
	current             fs.FarmStateMap
	history             []fs.FarmStateMap
	farmStateChangeChan chan fs.FarmStateMap
	sm.IStateMachine
	fs.FarmStore
}

/*cs := colfer.DeviceState{}
  bytes, err := cs.UnmarshalBinary()
  if err != nil {
    panic(err)
  }*/

func NewFarmStateMachine(logger *logging.Logger, farmID uint64,
	farmStateChangeChan chan fs.FarmStateMap) FarmStateMachine {

	return &FarmSM{
		logger:              logger,
		farmID:              farmID,
		history:             make([]fs.FarmStateMap, 0),
		farmStateChangeChan: farmStateChangeChan}
}

func (s *FarmSM) CreateStateMachine(clusterID, nodeID uint64) sm.IStateMachine {
	s.clusterID = clusterID
	s.nodeID = nodeID
	return s
}

func (s *FarmSM) Lookup(query interface{}) (interface{}, error) {

	s.logger.Warningf("[FarmStateMachine.Lookup] query: %+v", query)
	s.logger.Warningf("[FarmStateMachine.Lookup] Current: %+v", s.current)
	//s.logger.Warningf("[FarmStateMachine.Lookup] History: %+v", s.history)

	if query == nil {
		return []fs.FarmStateMap{s.current}, nil
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

func (s *FarmSM) Update(data []byte) (sm.Result, error) {
	var farmState fs.FarmState
	err := json.Unmarshal(data, &farmState)
	if err != nil {
		s.logger.Errorf("[FarmStateMachine.Update] Error: %s\n", err)
		return sm.Result{}, err
	}
	if s.current != nil {
		s.history = append(s.history, s.current)
	}
	s.current = &farmState

	s.logger.Debugf("[FarmStateMachine.Update] farm.id: %d, store.history.len: %d, farm: %+v\n",
		s.farmID, len(s.history), string(data))

	s.farmStateChangeChan <- &farmState

	return sm.Result{Value: s.farmID, Data: data}, nil
}

// SaveSnapshot saves the current IStateMachine state into a snapshot using the
// specified io.Writer object.
func (s *FarmSM) SaveSnapshot(w io.Writer, fc sm.ISnapshotFileCollection, done <-chan struct{}) error {
	snap := s.history
	if s.current != nil {
		snap = append(snap, s.current)
	}
	bytes, err := json.Marshal(snap)
	if err != nil {
		s.logger.Errorf("[FarmStateMachine.SaveSnapshot] Error: %s", err)
		return err
	}
	s.logger.Infof("[FarmStateMachine.SaveSnapshot] Created new snaphot. History length: %d", len(snap))
	_, err = w.Write(bytes)
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
	if len(farmState) > 0 {
		s.history = make([]fs.FarmStateMap, len(farmState))
		for i, h := range farmState {
			s.history[i] = &h
		}
		s.current = s.history[len(farmState)-1]
	}
	return nil
}

// Close closes the IStateMachine instance. There is nothing for us to cleanup
// or release as this is a pure in memory data store. Note that the Close
// method is not guaranteed to be called as node can crash at any time.
func (s *FarmSM) Close() error { return nil }

// GetHash returns a uint64 representing the current object state.
func (s *FarmSM) GetHash() (uint64, error) {
	return uint64(s.current.GetTimestamp()), nil
}
