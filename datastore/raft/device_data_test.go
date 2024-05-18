//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"encoding/json"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/stretchr/testify/assert"
)

func TestDeviceDataCRUD(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	deviceID := DeviceDataClusterID

	deviceDataDAO := NewRaftDeviceDataDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		deviceID)
	assert.NotNil(t, deviceDataDAO)
	deviceDataDAO.StartLocalCluster(IntegrationTestCluster, true)

	metrics := map[string]float64{
		"metric1": 12.34,
		"metric2": 56.78}
	channels := []int{0, 1, 0, 1, 0}

	deviceStateMap := state.NewDeviceStateMap()
	// /deviceStateMap.SetID(deviceID)
	deviceStateMap.SetFarmID(1)
	deviceStateMap.SetMetrics(metrics)
	deviceStateMap.SetChannels(channels)
	//deviceStateMap.SetTimestamp(timestamp)

	data, err := json.Marshal(deviceStateMap)
	assert.Nil(t, err)

	var unmarshalledData state.DeviceState
	err = json.Unmarshal(data, &unmarshalledData)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), unmarshalledData.Identifier())

	err = deviceDataDAO.Save(deviceID, deviceStateMap)
	assert.Nil(t, err)

	persistedDeviceData, err := deviceDataDAO.GetLast30Days(deviceID, "metric1")
	assert.Nil(t, err)
	assert.NotNil(t, persistedDeviceData)
	assert.Equal(t, persistedDeviceData[0], metrics["metric1"])
}
