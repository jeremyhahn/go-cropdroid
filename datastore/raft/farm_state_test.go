//go:build cluster && pebble
// +build cluster,pebble

package raft

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
	raftNode1 := IntegrationTestCluster.GetRaftNode1()

	farmStateChangeChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)

	farmStateStore := NewRaftFarmStateStore(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		FarmConfigClusterID,
		farmStateChangeChan)
	assert.NotNil(t, farmStateStore)
	farmStateStore.StartLocalCluster(IntegrationTestCluster, true)

	deviceName := "device1"
	now := time.Now()
	timestamp := now.Unix()
	deviceID := idGenerator.NewStringID(deviceName)

	metrics := map[string]float64{
		"metric1": 12.34,
		"metric2": 56.78}

	channels := []int{0, 1, 0, 1, 0}

	deviceStateMap := state.NewDeviceStateMap()
	deviceStateMap.SetID(deviceID)
	deviceStateMap.SetFarmID(FarmConfigClusterID)
	deviceStateMap.SetMetrics(metrics)
	deviceStateMap.SetChannels(channels)
	deviceStateMap.SetTimestamp(now)

	farmStateMap := &state.FarmState{
		ID: farmStateStore.ClusterID(),
		Devices: map[string]state.DeviceStateMap{
			deviceName: deviceStateMap},
		Timestamp: timestamp}

	err := farmStateStore.Put(farmStateStore.ClusterID(), farmStateMap)
	assert.Nil(t, err)

	persistedFarmState, err := farmStateStore.Get(farmStateStore.ClusterID())
	assert.Nil(t, err)

	assert.Equal(t, farmStateStore.ClusterID(), persistedFarmState.GetFarmID())

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
