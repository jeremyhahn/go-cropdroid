// +build cluster

package statemachine

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/lni/dragonboat/v3/statemachine"
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

func TestOpen(t *testing.T) {

	logger := createLogger()
	configClusterID := uint64(1234567890)
	farmChangeConfigChan := make(chan config.FarmConfig, 5)
	farmConfigMachine := NewFarmConfigMachine(logger,
		configClusterID, farmChangeConfigChan, 0)

	lastAppliedIndex, err := farmConfigMachine.Open(nil)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, uint64(0), lastAppliedIndex)
}

func TestFarmStateMachineUpdateLookupEmptyStore(t *testing.T) {

	logger := createLogger()
	configClusterID := uint64(1234567890)
	farmChangeConfigChan := make(chan config.FarmConfig, 5)
	farmConfigMachine := NewFarmConfigMachine(logger,
		configClusterID, farmChangeConfigChan, 0)

	farm1 := config.NewFarm()
	farm1.SetID(uint64(123))
	farm1.SetMode("Test Farm 1")
	farm1.SetDevices([]config.Device{
		{
			Type: "server",
			Configs: []config.DeviceConfigItem{
				{
					Key:   "name",
					Value: "Test Farm"},
				{
					Key:   "mode",
					Value: "test"},
				{
					Key:   "interval",
					Value: "59"}}}})

	farm2 := config.NewFarm()
	farm2.SetMode("Test Farm 2")
	farm2.SetDevices([]config.Device{
		{
			Type: "server",
			Configs: []config.DeviceConfigItem{
				{
					Key:   "name",
					Value: "Test Farm 2"},
				{
					Key:   "mode",
					Value: "test2"},
				{
					Key:   "interval",
					Value: "60"}}}})

	data, err := json.Marshal(farm1)
	assert.Nil(t, err)

	entries := []statemachine.Entry{{
		Index: 1,
		Cmd:   data}}

	lastAppliedIndex, err := farmConfigMachine.Open(nil)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, uint64(0), lastAppliedIndex)

	ents, err := farmConfigMachine.Update(entries)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ents))
	for i, ent := range ents {
		assert.Equal(t, ent.Result.Value, uint64(len(entries[i].Cmd)))
	}

	configClusterIdBytes := util.ClusterHashAsBytes(farm1.GetOrganizationID(), farm1.GetID())
	persisted, err := farmConfigMachine.Lookup(configClusterIdBytes)
	assert.Nil(t, err)
	assert.NotNil(t, persisted)

	farmConfig := persisted.(config.FarmConfig)
	assert.Equal(t, farmConfig.GetID(), farm1.GetID())
}
