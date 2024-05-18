//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"
)

func TestFarmStateCRUD(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)

	err := Cluster.CreateFarmStateCluster(FarmStateClusterID)
	assert.Nil(t, err)

	farmStateStore := NewRaftFarmStateStore(Cluster.app.Logger,
		Cluster.GetRaftNode1())
	assert.NotNil(t, farmStateStore)

	deviceName := "device1"
	timestamp := time.Now().Unix()
	deviceID := idGenerator.NewStringID(deviceName)

	metrics := map[string]float64{
		"metric1": 12.34,
		"metric2": 56.78}

	channels := []int{0, 1, 0, 1, 0}

	deviceStateMap := state.NewDeviceStateMap()
	deviceStateMap.SetID(deviceID)
	deviceStateMap.SetFarmID(FarmStateClusterID)
	deviceStateMap.SetMetrics(metrics)
	deviceStateMap.SetChannels(channels)
	//deviceStateMap.SetTimestamp(timestamp)

	farmStateMap := &state.FarmState{
		ID: FarmStateClusterID,
		Devices: map[string]state.DeviceStateMap{
			deviceName: deviceStateMap},
		Timestamp: timestamp}

	err = farmStateStore.Put(FarmStateClusterID, farmStateMap)
	assert.Nil(t, err)

	persistedFarmState, err := farmStateStore.Get(FarmStateClusterID)
	assert.Nil(t, err)

	assert.Equal(t, FarmStateClusterID, persistedFarmState.GetFarmID())

	persistedDevice, err := persistedFarmState.GetDevice(deviceName)
	assert.Nil(t, err)
	assert.Equal(t, deviceID, persistedDevice.Identifier())

	persistedMetrics, err := persistedFarmState.GetMetrics(deviceName)
	assert.Nil(t, err)
	assert.Equal(t, metrics, persistedMetrics)

	persistedChannels, err := persistedFarmState.GetChannels(deviceName)
	assert.Nil(t, err)
	assert.Equal(t, channels, persistedChannels)
}
