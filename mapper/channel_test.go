package mapper

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/stretchr/testify/assert"
)

func TestChannelMapperMapConfigToModel(t *testing.T) {
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
	model := mapper.MapConfigToModel(entity)
	assert.Equal(t, model.GetID(), entity.ID)
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

func TestChannelMapperMapModelToConfig(t *testing.T) {

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

	assert.Equal(t, channel.GetID(), entity.ID)
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
