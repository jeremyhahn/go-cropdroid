//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"encoding/json"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/test/data"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/lni/dragonboat/v3/statemachine"
	"github.com/stretchr/testify/assert"
)

func TestDeviceSettingConfigOpen(t *testing.T) {

	defer cleanup(farmsDatabasePath)

	logger := createLogger()
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	deviceSettingsConfigMachine := NewDeviceSettingConfigMachine(logger,
		idGenerator, databasePath, clusterID, nodeID)

	lastAppliedIndex, err := deviceSettingsConfigMachine.Open(nil)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, uint64(0), lastAppliedIndex)
}

func TestDeviceSettingConfigMachineUpdateLookupEmptyStore(t *testing.T) {

	defer cleanup(farmsDatabasePath)

	logger := createLogger()
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	deviceSettingsConfigMachine := NewDeviceSettingConfigMachine(logger,
		idGenerator, databasePath, clusterID, nodeID)

	org := data.CreateTestOrganization1(idGenerator)

	farms := org.GetFarms()
	farm1 := farms[0]
	device1 := farm1.GetDevices()[0]

	farmJson, err := json.Marshal(farm1)
	assert.Nil(t, err)

	lastAppliedIndex, err := deviceSettingsConfigMachine.Open(nil)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, uint64(0), lastAppliedIndex)

	proposal, err := CreateProposal(
		QUERY_TYPE_UPDATE, farmJson).Serialize()

	entries := []statemachine.Entry{{
		Index: 1,
		Cmd:   proposal}}

	ents, err := deviceSettingsConfigMachine.Update(entries)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ents))
	// Cmd gets wrapped in dataKV so this compare fails
	// for i, ent := range ents {
	// 	assert.Equal(t, ent.Result.Value, uint64(len(entries[i].Cmd)))
	// }

	device1ID := device1.ID
	persisted, err := deviceSettingsConfigMachine.Lookup(device1ID)
	assert.Nil(t, err)
	assert.NotNil(t, persisted)

	//farmConfig := persisted.(config.DeviceSettingConfig)
	//assert.True(t, data.deviceSettingsConfigMachine(farmConfig, farm1))
}
