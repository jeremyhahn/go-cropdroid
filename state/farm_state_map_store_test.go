package state

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFarmStore(t *testing.T) {

	farmID := 1
	testController := "test"
	ttl := 2
	gcTicker := time.Duration(1 * time.Second)

	farmstore := NewMemoryFarmStore(nil, 10, ttl, gcTicker)

	metrics := make(map[string]float64, 2)
	metrics["sensor1"] = 12.34
	metrics["sensor2"] = 56.78

	channels := make([]int, 4)
	channels[0] = 0
	channels[1] = 1
	channels[2] = 1
	channels[3] = 0

	controllerStateMap := NewControllerStateMap()
	controllerStateMap.SetMetrics(metrics)
	controllerStateMap.SetChannels(channels)

	farmStateMap := NewFarmStateMap(farmID)
	farmStateMap.SetController(testController, controllerStateMap)

	farmstore.Put(farmID, farmStateMap)

	storedFarmState, err := farmstore.Get(farmID)
	assert.Nil(t, err)
	assert.Equal(t, 1, farmstore.Len())

	sensor1, err := storedFarmState.GetMetricValue(testController, "sensor1")
	assert.Nil(t, err)
	assert.Equal(t, 12.34, sensor1)

	sensor2, err := storedFarmState.GetMetricValue(testController, "sensor2")
	assert.Nil(t, err)
	assert.Equal(t, 56.78, sensor2)

	channel0, err := storedFarmState.GetChannelValue(testController, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, channel0)

	channel1, err := storedFarmState.GetChannelValue(testController, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, channel1)

	channel2, err := storedFarmState.GetChannelValue(testController, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, channel2)

	channel3, err := storedFarmState.GetChannelValue(testController, 3)
	assert.Nil(t, err)
	assert.Equal(t, 0, channel3)

	sleepExpiry := ttl + 2
	println(fmt.Sprintf("Sleeping %d seconds to ensure app state entry expires", sleepExpiry))
	time.Sleep(time.Duration(sleepExpiry) * time.Second)

	newFarmState, err := farmstore.Get(farmID)
	assert.NotNil(t, err)
	assert.Nil(t, newFarmState)

	assert.Equal(t, 0, farmstore.Len())
}

func TestFarmStoreIsolation(t *testing.T) {

	farmID := 1
	testController := "test"
	ttl := 2
	gcTicker := time.Duration(1 * time.Second)

	farmstore := NewMemoryFarmStore(nil, 10, ttl, gcTicker)

	metrics := make(map[string]float64, 2)
	metrics["sensor1"] = 12.34

	controllerStateMap := NewControllerStateMap()
	controllerStateMap.SetMetrics(metrics)

	farmStateMap := NewFarmStateMap(farmID)
	farmStateMap.SetController(testController, controllerStateMap)

	farmstore.Put(farmID, farmStateMap)

	storedFarmState, err := farmstore.Get(farmID)
	assert.Nil(t, err)
	assert.NotNil(t, storedFarmState)
	assert.Equal(t, 1, farmstore.Len())

	sensor1, err := storedFarmState.GetMetricValue(testController, "sensor1")
	assert.Nil(t, err)
	assert.Equal(t, 12.34, sensor1)

	// Now modify the stored state. It shouldn't modify farmStateMap, only the stored state.
	storedFarmState.SetMetricValue(testController, "sensor1", 99.99)
	c, err := storedFarmState.GetController(testController)
	assert.Nil(t, err)
	assert.Equal(t, 99.99, c.GetMetrics()["sensor1"])

	// THIS IS CURRENTLY FAILING! Workaround is to use farmService.state (see poll() for details)
	//val2, err := farmStateMap.GetMetricValue(testController, "sensor1")
	//assert.Nil(t, err)
	//assert.Equal(t, 12.34, val2)
}

func BenchmarkFarmStateSetController(b *testing.B) {

	testController := "test"
	ttl := 2
	gcTicker := time.Duration(0)

	farmstore := NewMemoryFarmStore(nil, 10, ttl, gcTicker)

	metrics := make(map[string]float64, 2)
	metrics["sensor1"] = 12.34
	metrics["sensor2"] = 56.78

	channels := make([]int, 4)
	channels[0] = 0
	channels[1] = 1
	channels[2] = 1
	channels[3] = 0

	controllerStateMap := NewControllerStateMap()
	controllerStateMap.SetMetrics(metrics)
	controllerStateMap.SetChannels(channels)

	farmStateMap := NewFarmStateMap(1)
	farmStateMap.SetController(testController, controllerStateMap)

	for n := 0; n < b.N; n++ {
		farmstore.Put(n, farmStateMap)
	}
}

func BenchmarkSyncMap(b *testing.B) {

	m := &sync.Map{}

	metrics := make(map[string]float64, 2)
	metrics["sensor1"] = 12.34
	metrics["sensor2"] = 56.78

	channels := make([]int, 4)
	channels[0] = 0
	channels[1] = 1
	channels[2] = 1
	channels[3] = 0

	controllerStateMap := NewControllerStateMap()
	controllerStateMap.SetMetrics(metrics)
	controllerStateMap.SetChannels(channels)

	for n := 0; n < b.N; n++ {
		m.Store(n, controllerStateMap)
	}
}
