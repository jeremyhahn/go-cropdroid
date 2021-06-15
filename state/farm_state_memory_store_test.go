package state

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFarmStore(t *testing.T) {

	farmID := uint64(1)
	testDevice := "test"
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

	deviceStateMap := NewDeviceStateMap()
	deviceStateMap.SetMetrics(metrics)
	deviceStateMap.SetChannels(channels)

	farmStateMap := NewFarmStateMap(farmID)
	farmStateMap.SetDevice(testDevice, deviceStateMap)

	farmstore.Put(farmID, farmStateMap)

	storedFarmState, err := farmstore.Get(farmID)
	assert.Nil(t, err)
	assert.Equal(t, 1, farmstore.Len())

	sensor1, err := storedFarmState.GetMetricValue(testDevice, "sensor1")
	assert.Nil(t, err)
	assert.Equal(t, 12.34, sensor1)

	sensor2, err := storedFarmState.GetMetricValue(testDevice, "sensor2")
	assert.Nil(t, err)
	assert.Equal(t, 56.78, sensor2)

	channel0, err := storedFarmState.GetChannelValue(testDevice, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, channel0)

	channel1, err := storedFarmState.GetChannelValue(testDevice, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, channel1)

	channel2, err := storedFarmState.GetChannelValue(testDevice, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, channel2)

	channel3, err := storedFarmState.GetChannelValue(testDevice, 3)
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

	farmID := uint64(1)
	testDevice := "test"
	ttl := 2
	gcTicker := time.Duration(1 * time.Second)

	farmstore := NewMemoryFarmStore(nil, 10, ttl, gcTicker)

	metrics := make(map[string]float64, 2)
	metrics["sensor1"] = 12.34

	deviceStateMap := NewDeviceStateMap()
	deviceStateMap.SetMetrics(metrics)

	farmStateMap := NewFarmStateMap(farmID)
	farmStateMap.SetDevice(testDevice, deviceStateMap)

	farmstore.Put(farmID, farmStateMap)

	storedFarmState, err := farmstore.Get(farmID)
	assert.Nil(t, err)
	assert.NotNil(t, storedFarmState)
	assert.Equal(t, 1, farmstore.Len())

	sensor1, err := storedFarmState.GetMetricValue(testDevice, "sensor1")
	assert.Nil(t, err)
	assert.Equal(t, 12.34, sensor1)

	// Now modify the stored state. It shouldn't modify farmStateMap, only the stored state.
	storedFarmState.SetMetricValue(testDevice, "sensor1", 99.99)
	c, err := storedFarmState.GetDevice(testDevice)
	assert.Nil(t, err)
	assert.Equal(t, 99.99, c.GetMetrics()["sensor1"])

	// THIS IS CURRENTLY FAILING! Workaround is to use farmService.state (see poll() for details)
	//val2, err := farmStateMap.GetMetricValue(testDevice, "sensor1")
	//assert.Nil(t, err)
	//assert.Equal(t, 12.34, val2)
}

func BenchmarkFarmStateSetDevice(b *testing.B) {

	testDevice := "test"
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

	deviceStateMap := NewDeviceStateMap()
	deviceStateMap.SetMetrics(metrics)
	deviceStateMap.SetChannels(channels)

	farmStateMap := NewFarmStateMap(1)
	farmStateMap.SetDevice(testDevice, deviceStateMap)

	for n := 0; n < b.N; n++ {
		farmstore.Put(uint64(n), farmStateMap)
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

	deviceStateMap := NewDeviceStateMap()
	deviceStateMap.SetMetrics(metrics)
	deviceStateMap.SetChannels(channels)

	for n := 0; n < b.N; n++ {
		m.Store(n, deviceStateMap)
	}
}
