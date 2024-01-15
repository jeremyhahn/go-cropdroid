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

func (mapper *DefaultChannelMapper) MapConfigToModel(config *config.Channel) common.Channel {
	return &model.Channel{
		ID:          config.GetID(),
		DeviceID:    config.GetDeviceID(),
		ChannelID:   config.GetChannelID(),
		Name:        config.GetName(),
		Enable:      config.IsEnabled(),
		Notify:      config.IsNotify(),
		Duration:    config.GetDuration(),
		Debounce:    config.GetDebounce(),
		Backoff:     config.GetBackoff(),
		AlgorithmID: config.GetAlgorithmID()}
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
		AlgorithmID: model.GetAlgorithmID()}
}
