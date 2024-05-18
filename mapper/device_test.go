package mapper

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/stretchr/testify/assert"
)

func TestDeviceMapStateToEntity(t *testing.T) {

	mapper := NewDeviceMapper(NewMetricMapper(), NewChannelMapper())

	metrics := map[string]float64{
		"mem":     1200,
		"sensor1": 12.34}
	channels := []int{1, 0, 1}

	state := state.NewDeviceStateMap()
	state.SetMetrics(metrics)
	state.SetChannels(channels)

	config := &config.Device{
		ID:          1,
		Type:        "test",
		Description: "Fake device used for unit testing",
		Settings: []*config.DeviceSetting{
			{
				Key:   "enable",
				Value: "true"},
			{
				Key:   "notify",
				Value: "true"},
			{
				Key:   "uri",
				Value: ""}},
		Metrics: []*config.Metric{
			{
				ID:        1,
				DeviceID:  2,
				Name:      "Available Memory",
				Key:       "mem",
				Enable:    true,
				Notify:    true,
				Unit:      "bytes",
				AlarmLow:  1000,
				AlarmHigh: 10000},
			{
				ID:        2,
				DeviceID:  2,
				Name:      "Fake Temp Sensor",
				Key:       "sensor1",
				Enable:    true,
				Notify:    true,
				Unit:      "°",
				AlarmLow:  92.34,
				AlarmHigh: 65.12}},
		Channels: []*config.Channel{
			{
				ID:          1,
				DeviceID:    2,
				ChannelID:   0,
				Name:        "Test Channel 1",
				Enable:      true,
				Notify:      true,
				Conditions:  nil,
				Schedule:    nil,
				Duration:    1,
				Debounce:    2,
				Backoff:     3,
				AlgorithmID: 4},
			{
				ID:          2,
				DeviceID:    3,
				ChannelID:   1,
				Name:        "Test Channel 2",
				Enable:      false,
				Notify:      false,
				Conditions:  nil,
				Schedule:    nil,
				Duration:    1,
				Debounce:    2,
				Backoff:     3,
				AlgorithmID: 4},
			{
				ID:          3,
				DeviceID:    4,
				ChannelID:   2,
				Name:        "Test Channel 3",
				Enable:      false,
				Notify:      false,
				Conditions:  nil,
				Schedule:    nil,
				Duration:    1,
				Debounce:    2,
				Backoff:     3,
				AlgorithmID: 4}}}

	device, err := mapper.MapStateToDevice(state, config)
	assert.Nil(t, err)
	assert.Equal(t, config.ID, device.GetID())
	assert.Equal(t, config.GetType(), device.GetType())
	assert.Equal(t, config.GetDescription(), device.GetDescription())
	/*assert.Equal(t, config.IsEnabled(), device.IsEnabled())
	assert.Equal(t, config.IsNotify(), device.IsNotify())
	assert.Equal(t, config.GetURI(), device.GetURI())*/
	//assert.Equal(t, config.GetHardwareVersion(), device.GetHardwareVersion())
	//assert.Equal(t, config.GetFirmwareVersion(), device.GetFirmwareVersion())

	memMetric, err := device.GetMetric("mem")
	assert.Nil(t, err)
	assert.Equal(t, metrics["mem"], memMetric.GetValue())

	sensor1Metric, err := device.GetMetric("sensor1")
	assert.Nil(t, err)
	assert.Equal(t, metrics["sensor1"], sensor1Metric.GetValue())

	channel0, err := device.GetChannel(0)
	assert.Nil(t, err)
	assert.Equal(t, channels[0], channel0.GetValue())

	channel1, err := device.GetChannel(1)
	assert.Nil(t, err)
	assert.Equal(t, channels[1], channel1.GetValue())

	channel2, err := device.GetChannel(2)
	assert.Nil(t, err)
	assert.Equal(t, channels[2], channel2.GetValue())
}

func TestDeviceMapConfigToModel(t *testing.T) {

	mapper := NewDeviceMapper(NewMetricMapper(), NewChannelMapper())

	deviceEntity := &config.Device{
		ID: 2,
		//OrganizationID: 2,
		Type:        "test",
		Description: "Fake microdevice used for testing"}

	settingEntities := []*config.DeviceSetting{
		{
			Key:   "test.enable",
			Value: "true"},
		{
			Key:   "test.notify",
			Value: "true"},
		{
			Key:   "test.uri",
			Value: "true"},
		{
			Key:   "key1",
			Value: "key1.value1"},
		{
			Key:   "key2",
			Value: "key2.value"},
		{
			Key:   "test.key2",
			Value: "test.key2.value"}}

	deviceModel, err := mapper.MapConfigToModel(deviceEntity, settingEntities)
	assert.Nil(t, err)
	assert.Equal(t, deviceEntity.ID, deviceModel.GetID())
	assert.Equal(t, deviceEntity.GetType(), deviceModel.GetType())
	assert.Equal(t, deviceEntity.GetDescription(), deviceModel.GetDescription())
	//assert.Equal(t, deviceEntity.GetHardwareVersion(), deviceModel.GetHardwareVersion())
	//assert.Equal(t, deviceEntity.GetFirmwareVersion(), deviceModel.GetFirmwareVersion())

	configMap := deviceModel.GetConfigs()
	assert.Equal(t, len(settingEntities), len(configMap))
	assert.Equal(t, configMap["key1"], "key1.value1")
	assert.Equal(t, configMap["key2"], "key2.value")
	assert.Equal(t, configMap["test.key2"], "test.key2.value")
}
