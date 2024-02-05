package state

import (
	"os"
	"testing"
	"time"

	logging "github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
)

func TestDeviceStateStore(t *testing.T) {

	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logger := logging.MustGetLogger("device-state")

	deviceID := uint64(1)
	ttl := 10
	gcTicker := time.Duration(1 * time.Second)

	deviceStateStore := NewMemoryDeviceStore(logger, 10, ttl, gcTicker)

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

	deviceStateStore.Put(deviceID, deviceStateMap)

	storedDeviceState, err := deviceStateStore.Get(deviceID)
	assert.Nil(t, err)
	assert.Equal(t, 1, deviceStateStore.Len())

	storedMetrics := storedDeviceState.GetMetrics()
	storedChannels := storedDeviceState.GetChannels()

	assert.Equal(t, 12.34, storedMetrics["sensor1"])
	assert.Equal(t, 56.78, storedMetrics["sensor2"])

	assert.Equal(t, 0, storedChannels[0])
	assert.Equal(t, 1, storedChannels[1])
	assert.Equal(t, 1, storedChannels[2])
	assert.Equal(t, 0, storedChannels[3])

	metrics["sensor2"] = 45.67
	deviceStateMap.SetMetrics(metrics)
	deviceStateStore.Put(deviceID, deviceStateMap)
	storedDeviceState2, err := deviceStateStore.Get(deviceID)
	storedMetrics2 := storedDeviceState2.GetMetrics()
	assert.Equal(t, 45.67, storedMetrics2["sensor2"])

	//    sleepExpiry := ttl + 2
	//    println(fmt.Sprintf("Sleeping %d seconds to ensure app state entry expires", sleepExpiry))
	//    time.Sleep(time.Duration(sleepExpiry) * time.Second)

	//    newFarmState, err := farmstore.Get(farmID)
	//    assert.NotNil(t, err)
	//    assert.Nil(t, newFarmState)
}
