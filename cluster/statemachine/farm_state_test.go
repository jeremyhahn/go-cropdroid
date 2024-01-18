//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/state"
	fs "github.com/jeremyhahn/go-cropdroid/state"
	"github.com/lni/dragonboat/v3/statemachine"
	"github.com/stretchr/testify/assert"
)

func TestFarmStateMachineUpdateAndLookup(t *testing.T) {

	farmStateChangeChan := make(chan fs.FarmStateMap, 1)

	logger := createLogger()
	sm := NewFarmStateConcurrentStateMachine(logger, 1, farmStateChangeChan)

	farmStateMap, deviceStateMap := createTestFarmStateMap()

	state, err := farmStateMap.GetDevice("testdevice")
	assert.Nil(t, err)
	assert.Equal(t, deviceStateMap, state)
	assert.Equal(t, 12.34, state.GetMetrics()["test"])
	assert.Equal(t, 56.7, state.GetMetrics()["test2"])
	assert.Equal(t, 1, state.GetChannels()[0])
	assert.Equal(t, 0, state.GetChannels()[1])

	bytes, err := json.Marshal(farmStateMap)
	assert.Nil(t, err)

	entries := []statemachine.Entry{{
		Index: 1,
		Cmd:   bytes}}

	results, err := sm.Update(entries)
	assert.Nil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, bytes, results[0].Cmd)

	result2, err := sm.Lookup(nil)
	assert.Nil(t, err)

	queryResult := result2.(fs.FarmStateMap)
	assert.Equal(t, uint64(1), queryResult.GetFarmID())

	queryBytes, err := json.Marshal(queryResult)
	assert.Nil(t, err)
	assert.Equal(t, bytes, queryBytes)

	// Data is published to connected clients via FarmService
	// after the data gets committed to the cluster by Update(entries).
	assert.Equal(t, 1, len(farmStateChangeChan))
	dataToPublish := <-farmStateChangeChan
	assert.NotNil(t, dataToPublish)
	assert.Equal(t, dataToPublish, farmStateMap)
	assert.Equal(t, 0, len(farmStateChangeChan))
}

func TestPrepareSnapshot(t *testing.T) {

	farmStateChangeChan := make(chan fs.FarmStateMap, 1)

	logger := createLogger()
	sm := NewFarmStateConcurrentStateMachine(logger, 1, farmStateChangeChan)

	farmStateMap, _ := createTestFarmStateMap()

	bytes, err := json.Marshal(farmStateMap)
	assert.Nil(t, err)

	entries := []statemachine.Entry{{
		Index: 1,
		Cmd:   bytes}}

	results, err := sm.Update(entries)
	assert.Nil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, bytes, results[0].Cmd)

	snapshotID, err := sm.PrepareSnapshot()
	assert.Equal(t, uint64(1), snapshotID)

	snapshotID2, err := sm.PrepareSnapshot()
	assert.Equal(t, uint64(2), snapshotID2)
}

func TestSaveAndRecoverSnapshot(t *testing.T) {

	farmStateChangeChan := make(chan fs.FarmStateMap, 1)

	logger := createLogger()
	sm := NewFarmStateConcurrentStateMachine(logger, 1, farmStateChangeChan)

	farmStateMap, _ := createTestFarmStateMap()

	stateBytes, err := json.Marshal(farmStateMap)
	assert.Nil(t, err)

	entries := []statemachine.Entry{{
		Index: 1,
		Cmd:   stateBytes}}

	results, err := sm.Update(entries)
	assert.Nil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, stateBytes, results[0].Cmd)

	snapshotID, err := sm.PrepareSnapshot()
	assert.Equal(t, uint64(1), snapshotID)

	done := make(chan struct{}, 1)

	var b bytes.Buffer
	_, err = b.Write(stateBytes)
	assert.Nil(t, err)

	err = sm.SaveSnapshot(snapshotID, &b, nil, done)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(done))

	sm2 := NewFarmStateConcurrentStateMachine(logger, 1, farmStateChangeChan)
	reader := strings.NewReader(string(stateBytes))
	err = sm2.RecoverFromSnapshot(reader, nil, done)
	assert.Nil(t, err)

	result, err := sm2.Lookup(nil)
	assert.Nil(t, err)

	persistedState := result.(state.FarmStateMap)
	assert.Equal(t, persistedState, farmStateMap)
	assert.Equal(t, persistedState.GetDevices(), farmStateMap.GetDevices())
}

// Create mock farm state for test assertions
func createTestFarmStateMap() (state.FarmStateMap, state.DeviceStateMap) {
	deviceStateMap := fs.NewDeviceStateMap()
	deviceStateMap.SetMetrics(map[string]float64{
		"test":  12.34,
		"test2": 56.7})
	deviceStateMap.SetChannels([]int{1, 0, 1, 0, 1, 1})

	farmStateMap := fs.NewFarmStateMap(1)
	farmStateMap.SetDevice("testdevice", deviceStateMap)
	return farmStateMap, deviceStateMap
}
