package mapper

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/stretchr/testify/assert"
)

func TestChannelMapperMapEntityToModel(t *testing.T) {
	mapper := NewChannelMapper()
	entity := &config.Channel{
		ID:          1,
		DeviceID:    2,
		ChannelID:   3,
		Name:        "Test Channel",
		Enable:      true,
		Notify:      true,
		Duration:    1,
		Debounce:    2,
		Backoff:     3,
		AlgorithmID: 4}
	model := mapper.MapEntityToModel(entity)
	assert.Equal(t, model.GetID(), entity.GetID())
	assert.Equal(t, model.GetDeviceID(), entity.GetDeviceID())
	assert.Equal(t, model.GetChannelID(), entity.GetChannelID())
	assert.Equal(t, model.GetName(), entity.GetName())
	assert.Equal(t, model.IsEnabled(), entity.IsEnabled())
	assert.Equal(t, model.IsNotify(), entity.IsNotify())
	assert.Equal(t, model.GetDuration(), entity.GetDuration())
	assert.Equal(t, model.GetDebounce(), entity.GetDebounce())
	assert.Equal(t, model.GetBackoff(), entity.GetBackoff())
	assert.Equal(t, model.GetAlgorithmID(), entity.GetAlgorithmID())
}

func TestChannelMapperMapConfigToModel(t *testing.T) {

	mapper := NewChannelMapper()

	var channelConfig config.ChannelConfig = &model.Channel{}
	channelConfig.SetID(1)
	channelConfig.SetDeviceID(2)
	channelConfig.SetName("Test Channel")
	channelConfig.SetEnable(true)
	channelConfig.SetNotify(true)
	channelConfig.SetDuration(1)
	channelConfig.SetDebounce(2)
	channelConfig.SetBackoff(3)
	channelConfig.SetAlgorithmID(4)

	channel, ok := channelConfig.(common.Channel)
	assert.True(t, ok)
	assert.ObjectsAreEqual(channelConfig, channel)

	assert.Equal(t, channelConfig.GetID(), channel.GetID())
	assert.Equal(t, channelConfig.GetDeviceID(), channel.GetDeviceID())
	assert.Equal(t, channelConfig.GetChannelID(), channel.GetChannelID())
	assert.Equal(t, channelConfig.GetName(), channel.GetName())
	assert.Equal(t, channelConfig.IsEnabled(), channel.IsEnabled())
	assert.Equal(t, channelConfig.IsNotify(), channel.IsNotify())
	assert.Equal(t, channelConfig.GetDuration(), channel.GetDuration())
	assert.Equal(t, channelConfig.GetDebounce(), channel.GetDebounce())
	assert.Equal(t, channelConfig.GetBackoff(), channel.GetBackoff())

	// Mapper must return new model object to prevent config objects from being updated by model pointer.
	channel.SetID(2)
	assert.Equal(t, 2, channelConfig.GetID())

	model := mapper.MapConfigToModel(channelConfig)
	assert.ObjectsAreEqual(channelConfig, model)
	model.SetID(3)
	assert.Equal(t, 2, channelConfig.GetID()) // This is the desired behavior
}

func TestChannelMapperMapModelToEntity(t *testing.T) {

	mapper := NewChannelMapper()

	var channel common.Channel = &model.Channel{}
	channel.SetID(1)
	channel.SetDeviceID(2)
	channel.SetName("Test Channel")
	channel.SetEnable(true)
	channel.SetNotify(true)
	channel.SetDuration(1)
	channel.SetDebounce(2)
	channel.SetBackoff(3)
	channel.SetAlgorithmID(4)

	entity := mapper.MapModelToConfig(channel)
	assert.ObjectsAreEqual(channel, entity)

	assert.Equal(t, channel.GetID(), entity.GetID())
	assert.Equal(t, channel.GetDeviceID(), entity.GetDeviceID())
	assert.Equal(t, channel.GetChannelID(), entity.GetChannelID())
	assert.Equal(t, channel.GetName(), entity.GetName())
	assert.Equal(t, channel.IsEnabled(), entity.IsEnabled())
	assert.Equal(t, channel.IsNotify(), entity.IsNotify())
	assert.Equal(t, channel.GetDuration(), entity.GetDuration())
	assert.Equal(t, channel.GetDebounce(), entity.GetDebounce())
	assert.Equal(t, channel.GetBackoff(), entity.GetBackoff())

	channel.SetDeviceID(20)
	assert.NotEqual(t, channel.GetDeviceID(), entity.GetDeviceID())
}

func TestChannelMapperMapEntityToConfig(t *testing.T) {

	mapper := NewChannelMapper()

	entity := &config.Channel{
		ID:          1,
		DeviceID:    2,
		ChannelID:   3,
		Name:        "Test Channel",
		Enable:      true,
		Notify:      true,
		Duration:    1,
		Debounce:    2,
		Backoff:     3,
		AlgorithmID: 4}

	config := mapper.MapEntityToConfig(entity)
	assert.Equal(t, entity.GetID(), config.GetID())
	assert.Equal(t, entity.GetDeviceID(), config.GetDeviceID())
	assert.Equal(t, entity.GetChannelID(), config.GetChannelID())
	assert.Equal(t, entity.GetName(), config.GetName())
	assert.Equal(t, entity.IsEnabled(), config.IsEnabled())
	assert.Equal(t, entity.IsNotify(), config.IsNotify())
	assert.Equal(t, entity.GetDuration(), config.GetDuration())
	assert.Equal(t, entity.GetDebounce(), config.GetDebounce())
	assert.Equal(t, entity.GetBackoff(), config.GetBackoff())
	assert.Equal(t, entity.GetAlgorithmID(), config.GetAlgorithmID())

	entity.SetDeviceID(20)
	assert.NotEqual(t, entity.GetDeviceID(), config.GetDeviceID())
}
