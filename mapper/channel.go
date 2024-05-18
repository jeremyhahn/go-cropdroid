package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type ChannelMapper interface {
	MapConfigToModel(config *config.Channel) common.Channel
	MapModelToConfig(model common.Channel) *config.Channel
}

type DefaultChannelMapper struct {
}

func NewChannelMapper() ChannelMapper {
	return &DefaultChannelMapper{}
}

func (mapper *DefaultChannelMapper) MapConfigToModel(channel *config.Channel) common.Channel {
	return &model.Channel{
		ID:          channel.ID,
		DeviceID:    channel.GetDeviceID(),
		ChannelID:   channel.GetChannelID(),
		Name:        channel.GetName(),
		Enable:      channel.IsEnabled(),
		Notify:      channel.IsNotify(),
		Duration:    channel.GetDuration(),
		Debounce:    channel.GetDebounce(),
		Backoff:     channel.GetBackoff(),
		AlgorithmID: channel.GetAlgorithmID(),
		Conditions:  make([]config.Condition, 0),
		Schedule:    make([]config.Schedule, 0)}
}

func (mapper *DefaultChannelMapper) MapModelToConfig(model common.Channel) *config.Channel {
	return &config.Channel{
		ID:          model.GetID(),
		DeviceID:    model.GetDeviceID(),
		ChannelID:   model.GetChannelID(),
		Name:        model.GetName(),
		Enable:      model.IsEnabled(),
		Notify:      model.IsNotify(),
		Duration:    model.GetDuration(),
		Debounce:    model.GetDebounce(),
		Backoff:     model.GetBackoff(),
		AlgorithmID: model.GetAlgorithmID(),
		Conditions:  make([]*config.Condition, 0),
		Schedule:    make([]*config.Schedule, 0)}
}
