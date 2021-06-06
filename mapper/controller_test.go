package mapper

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/stretchr/testify/assert"
)

func TestControllerMapStateToEntity(t *testing.T) {

	mapper := NewControllerMapper(NewMetricMapper(), NewChannelMapper())

	metrics := map[string]float64{
		"mem":     1200,
		"sensor1": 12.34}
	channels := []int{1, 0, 1}

	state := state.NewControllerStateMap()
	state.SetMetrics(metrics)
	state.SetChannels(channels)

	config := config.Controller{
		ID:          1,
		Type:        "test",
		Description: "Fake controller used for unit testing",
		Configs: []config.ControllerConfigItem{
			config.ControllerConfigItem{
				Key:   "enable",
				Value: "true"},
			config.ControllerConfigItem{
				Key:   "notify",
				Value: "true"},
			config.ControllerConfigItem{
				Key:   "uri",
				Value: ""}},
		Metrics: []config.Metric{
			config.Metric{
				ID:           1,
				ControllerID: 2,
				Name:         "Available Memory",
				Key:          "mem",
				Enable:       true,
				Notify:       true,
				Unit:         "bytes",
				AlarmLow:     1000,
				AlarmHigh:    10000},
			config.Metric{
				ID:           2,
				ControllerID: 2,
				Name:         "Fake Temp Sensor",
				Key:          "sensor1",
				Enable:       true,
				Notify:       true,
				Unit:         "°",
				AlarmLow:     92.34,
				AlarmHigh:    65.12}},
		Channels: []config.Channel{
			config.Channel{
				ID:           1,
				ControllerID: 2,
				ChannelID:    0,
				Name:         "Test Channel 1",
				Enable:       true,
				Notify:       true,
				Conditions:   nil,
				Schedule:     nil,
				Duration:     1,
				Debounce:     2,
				Backoff:      3,
				AlgorithmID:  4},
			config.Channel{
				ID:           2,
				ControllerID: 3,
				ChannelID:    1,
				Name:         "Test Channel 2",
				Enable:       false,
				Notify:       false,
				Conditions:   nil,
				Schedule:     nil,
				Duration:     1,
				Debounce:     2,
				Backoff:      3,
				AlgorithmID:  4},
			config.Channel{
				ID:           3,
				ControllerID: 4,
				ChannelID:    2,
				Name:         "Test Channel 3",
				Enable:       false,
				Notify:       false,
				Conditions:   nil,
				Schedule:     nil,
				Duration:     1,
				Debounce:     2,
				Backoff:      3,
				AlgorithmID:  4}}}

	controller, err := mapper.MapStateToController(state, config)
	assert.Nil(t, err)
	assert.Equal(t, config.GetID(), controller.GetID())
	assert.Equal(t, config.GetType(), controller.GetType())
	assert.Equal(t, config.GetDescription(), controller.GetDescription())
	/*assert.Equal(t, config.IsEnabled(), controller.IsEnabled())
	assert.Equal(t, config.IsNotify(), controller.IsNotify())
	assert.Equal(t, config.GetURI(), controller.GetURI())*/
	//assert.Equal(t, config.GetHardwareVersion(), controller.GetHardwareVersion())
	//assert.Equal(t, config.GetFirmwareVersion(), controller.GetFirmwareVersion())

	memMetric, err := controller.GetMetric("mem")
	assert.Nil(t, err)
	assert.Equal(t, metrics["mem"], memMetric.GetValue())

	sensor1Metric, err := controller.GetMetric("sensor1")
	assert.Nil(t, err)
	assert.Equal(t, metrics["sensor1"], sensor1Metric.GetValue())

	channel0, err := controller.GetChannel(0)
	assert.Nil(t, err)
	assert.Equal(t, channels[0], channel0.GetValue())

	channel1, err := controller.GetChannel(1)
	assert.Nil(t, err)
	assert.Equal(t, channels[1], channel1.GetValue())

	channel2, err := controller.GetChannel(2)
	assert.Nil(t, err)
	assert.Equal(t, channels[2], channel2.GetValue())
}

func TestControllerMapConfigToModel(t *testing.T) {

	mapper := NewControllerMapper(NewMetricMapper(), NewChannelMapper())

	controllerEntity := &config.Controller{
		ID: 2,
		//OrganizationID: 2,
		Type:        "test",
		Description: "Fake microcontroller used for testing"}

	configEntities := []config.ControllerConfigItem{
		config.ControllerConfigItem{
			Key:   "test.enable",
			Value: "true"},
		config.ControllerConfigItem{
			Key:   "test.notify",
			Value: "true"},
		config.ControllerConfigItem{
			Key:   "test.uri",
			Value: "true"},
		config.ControllerConfigItem{
			Key:   "key1",
			Value: "key1.value1"},
		config.ControllerConfigItem{
			Key:   "key2",
			Value: "key2.value"},
		config.ControllerConfigItem{
			Key:   "test.key2",
			Value: "test.key2.value"}}

	controllerModel, err := mapper.MapConfigToModel(controllerEntity, configEntities)
	assert.Nil(t, err)
	assert.Equal(t, controllerEntity.GetID(), controllerModel.GetID())
	assert.Equal(t, controllerEntity.GetType(), controllerModel.GetType())
	assert.Equal(t, controllerEntity.GetDescription(), controllerModel.GetDescription())
	//assert.Equal(t, controllerEntity.GetHardwareVersion(), controllerModel.GetHardwareVersion())
	//assert.Equal(t, controllerEntity.GetFirmwareVersion(), controllerModel.GetFirmwareVersion())

	configMap := controllerModel.GetConfigs()
	assert.Equal(t, len(configEntities), len(configMap))
	assert.Equal(t, configMap["key1"], "key1.value1")
	assert.Equal(t, configMap["key2"], "key2.value")
	assert.Equal(t, configMap["test.key2"], "test.key2.value")
}
