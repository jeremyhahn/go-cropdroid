// +build cluster

package state

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"hash/fnv"
	"io"
	"io/ioutil"

	"github.com/jeremyhahn/go-cropdroid/config"
	fs "github.com/jeremyhahn/go-cropdroid/state"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type FarmConfigMachine interface {
	CreateConfigMachine(clusterID, nodeID uint64) sm.IStateMachine
	sm.IStateMachine
}

type FarmConfigSM struct {
	logger               *logging.Logger
	clusterID            uint64
	nodeID               uint64
	farmID               uint64
	current              config.FarmConfig
	history              []config.FarmConfig
	historyMaxSize       int
	farmConfigChangeChan chan config.FarmConfig
	sm.IStateMachine
	fs.FarmStore
}

/*cs := colfer.ControllerState{}
  bytes, err := cs.UnmarshalBinary()
  if err != nil {
    panic(err)
  }*/

func NewFarmConfigMachine(logger *logging.Logger, farmID uint64,
	farmConfigChangeChan chan config.FarmConfig, historyMaxSize int) FarmConfigMachine {
	return &FarmConfigSM{
		logger:               logger,
		farmID:               farmID,
		history:              make([]config.FarmConfig, 0, historyMaxSize),
		farmConfigChangeChan: farmConfigChangeChan,
		historyMaxSize:       historyMaxSize}
}

func (s *FarmConfigSM) CreateConfigMachine(clusterID, nodeID uint64) sm.IStateMachine {
	s.clusterID = clusterID
	s.nodeID = nodeID
	return s
}

// TODO config.Controller.Configs is set to json:"-" and yaml:"-" because the API
// displays controller configs as key/value items. Probably time to create a view
// specific model for the API and remove the "-" so this loop is no longer needed.
func (s *FarmConfigSM) hydrateConfigs(farmConfig config.Farm) config.Farm {
	for _, controller := range farmConfig.GetControllers() {
		configs := make([]config.ControllerConfigItem, 0)
		for k, v := range controller.ConfigMap {
			configs = append(configs, *config.CreateControllerConfigItem(0, controller.GetID(), k, v))
		}
		controller.SetConfigs(configs)
	}
	return farmConfig
}

func (s *FarmConfigSM) Lookup(query interface{}) (interface{}, error) {

	s.logger.Warningf("[FarmConfigMachine.Lookup] query: %+v", query)
	s.logger.Warningf("[FarmConfigMachine.Lookup] Current: %+v", s.current)
	//s.logger.Warningf("[FarmConfigMachine.Lookup] History: %+v", s.history)

	return []config.FarmConfig{s.current}, nil
}

func (s *FarmConfigSM) Update(data []byte) (sm.Result, error) {
	var farmConfig config.Farm
	err := json.Unmarshal(data, &farmConfig)
	if err != nil {
		s.logger.Errorf("[FarmConfigMachine.Update] Error: %s\n", err)
		return sm.Result{}, err
	}

	farmConfig = s.hydrateConfigs(farmConfig)
	farmConfig.ParseConfigs()

	//newHistory := make([]config.FarmConfig, 0, s.historyMaxSize)
	newHistory := make([]config.FarmConfig, s.historyMaxSize)
	for i, conf := range s.history {
		newHistory[i+1] = conf
	}
	newHistory[0] = s.current
	s.current = &farmConfig

	s.logger.Debugf("[FarmConfigMachine.Update] farm.id: %d, store.history.len: %d, farm: %+v\n",
		s.farmID, len(s.history), string(data))

	s.farmConfigChangeChan <- &farmConfig

	return sm.Result{Value: s.farmID, Data: data}, nil
}

// SaveSnapshot saves the current IStateMachine state into a snapshot using the
// specified io.Writer object.
func (s *FarmConfigSM) SaveSnapshot(w io.Writer, fc sm.ISnapshotFileCollection, done <-chan struct{}) error {
	snap := s.history
	if s.current != nil {
		snap = append(snap, s.current)
	}
	bytes, err := json.Marshal(snap)
	if err != nil {
		s.logger.Errorf("[FarmConfigMachine.SaveSnapshot] Error: %s", err)
		return err
	}
	s.logger.Infof("[FarmConfigMachine.SaveSnapshot] Created new snaphot. History length: %d", len(snap))
	_, err = w.Write(bytes)
	return err
}

// RecoverFromSnapshot recovers the state using the provided snapshot.
func (s *FarmConfigSM) RecoverFromSnapshot(r io.Reader, files []sm.SnapshotFile, done <-chan struct{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var farmConfig []config.Farm
	err = json.Unmarshal(data, &farmConfig)
	if err != nil {
		s.logger.Errorf("[FarmConfigMachine.RecoverFromSnapshot] Error: %s, farmConfig: %+v\n", err, farmConfig)
		return err
	}
	s.logger.Debugf("[FarmConfigMachine.SaveSnapshot] Recovered from snapshot. History length: %d", len(farmConfig))
	if len(farmConfig) > 0 {
		s.history = make([]config.FarmConfig, len(farmConfig))
		for i, h := range farmConfig {
			s.history[i] = &h
		}
		s.current = s.history[len(farmConfig)-1]
	}
	return nil
}

// Close closes the IStateMachine instance. There is nothing for us to cleanup
// or release as this is a pure in memory data store. Note that the Close
// method is not guaranteed to be called as node can crash at any time.
func (s *FarmConfigSM) Close() error { return nil }

// GetHash returns a uint64 representing the current object state.
func (s *FarmConfigSM) GetHash() (uint64, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(s.current)
	if err != nil {
		return 0, err
	}
	h := fnv.New64a()
	h.Write(b.Bytes())
	return h.Sum64(), nil
}
