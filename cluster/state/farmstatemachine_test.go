// +build cluster

package state

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	fs "github.com/jeremyhahn/go-cropdroid/state"
	logging "github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
)

var TestSuiteName = "cropdroid_cluster_state_test"

func createLogger() *logging.Logger {

	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logger := logging.MustGetLogger(TestSuiteName)

	return logger
}

func TestFarmStateMachineUpdateLookupEmptyStore(t *testing.T) {

	logger := createLogger()
	sm := NewFarmStateMachine(logger, 1)

	controllerStateMap := fs.NewControllerStateMap()
	controllerStateMap.SetMetrics(map[string]float64{
		"test":  12.34,
		"test2": 56.7})
	controllerStateMap.SetChannels([]int{1, 0, 1, 0, 1, 1})

	farmStateMap := fs.NewFarmStateMap(1)
	farmStateMap.SetController("testcontroller", controllerStateMap)

	state, err := farmStateMap.GetController("testcontroller")
	assert.Nil(t, err)
	assert.Equal(t, controllerStateMap, state)
	assert.Equal(t, 12.34, state.GetMetrics()["test"])
	assert.Equal(t, 56.7, state.GetMetrics()["test2"])
	assert.Equal(t, 1, state.GetChannels()[0])
	assert.Equal(t, 0, state.GetChannels()[1])

	bytes, err := json.Marshal(farmStateMap)
	assert.Nil(t, err)

	result, err := sm.Update(bytes)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, uint64(1), result.Value)
	assert.Equal(t, bytes, result.Data)

	fmt.Printf("%+v\n", result)

	result2, err := sm.Lookup(nil)
	assert.Nil(t, err)

	queryResult := result2.([]fs.FarmStateMap)
	assert.Equal(t, 1, len(queryResult))
	assert.Equal(t, 1, queryResult[0].GetFarmID())

	queryBytes, err := json.Marshal(queryResult[0])
	assert.Nil(t, err)
	assert.Equal(t, bytes, queryBytes)

	fmt.Printf("%+v\n", queryBytes[0])

}

func TestFarmStateMachineUpdateLookupNonEmptyStore(t *testing.T) {

	logger := createLogger()
	sm := NewFarmStateMachine(logger, 1)

	controllerStateMap := fs.NewControllerStateMap()
	controllerStateMap.SetMetrics(map[string]float64{
		"test":  12.34,
		"test2": 56.7})
	controllerStateMap.SetChannels([]int{1, 0, 1, 0, 1, 1})

	farmStateMap := fs.NewFarmStateMap(1)
	farmStateMap.SetController("testcontroller", controllerStateMap)

	state, err := farmStateMap.GetController("testcontroller")
	assert.Nil(t, err)
	assert.Equal(t, controllerStateMap, state)
	assert.Equal(t, 12.34, state.GetMetrics()["test"])
	assert.Equal(t, 56.7, state.GetMetrics()["test2"])
	assert.Equal(t, 1, state.GetChannels()[0])
	assert.Equal(t, 0, state.GetChannels()[1])

	bytes, err := json.Marshal(farmStateMap)
	assert.Nil(t, err)

	result, err := sm.Update(bytes)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	fmt.Printf("%+v\n", result)

	assert.Equal(t, uint64(1), result.Value)
	assert.Equal(t, bytes, result.Data)

	result2, err := sm.Update(bytes)
	assert.Nil(t, err)
	assert.NotNil(t, result2)

	fmt.Printf("%+v\n", result2)

	result3, err := sm.Lookup(nil)
	assert.Nil(t, err)

	queryResult := result3.([]fs.FarmStateMap)
	assert.Equal(t, 1, len(queryResult))
	assert.Equal(t, 1, queryResult[0].GetFarmID())

	queryBytes, err := json.Marshal(queryResult[0])
	assert.Nil(t, err)
	assert.Equal(t, bytes, queryBytes)

	fmt.Printf("%+v\n", queryBytes[0])

}
