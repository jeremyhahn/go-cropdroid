//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"
)

func TestDeviceDataCRUD(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	//consistencyLevel := common.CONSISTENCY_LOCAL
	//testFarmStateName := "root@localhost"

	deviceName := "device1"
	deviceID := idGenerator.NewStringID(deviceName)

	Cluster.CreateDeviceDataCluster(deviceID)

	deviceDataDAO := NewRaftDeviceDataDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1())

	assert.NotNil(t, deviceDataDAO)

	metrics := map[string]float64{
		"metric1": 12.34,
		"metric2": 56.78}
	channels := []int{0, 1, 0, 1, 0}

	deviceStateMap := state.NewDeviceStateMap()
	deviceStateMap.SetID(deviceID)
	deviceStateMap.SetFarmID(1)
	deviceStateMap.SetMetrics(metrics)
	deviceStateMap.SetChannels(channels)
	//deviceStateMap.SetTimestamp(timestamp)

	data, err := json.Marshal(deviceStateMap)
	assert.Nil(t, err)

	var unmarshalledData state.DeviceState
	err = json.Unmarshal(data, &unmarshalledData)
	assert.Nil(t, err)
	assert.Equal(t, unmarshalledData.Identifier(), deviceStateMap.Identifier())

	err = deviceDataDAO.Save(deviceID, deviceStateMap)
	assert.Nil(t, err)

	persistedDeviceData, err := deviceDataDAO.GetLast30Days(deviceID, "metric1")
	assert.Nil(t, err)
	assert.NotNil(t, persistedDeviceData)
	assert.Equal(t, persistedDeviceData[0], metrics["metric1"])
}
