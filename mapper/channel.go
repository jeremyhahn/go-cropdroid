package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type ChannelMapper interface {
	MapEntityToModel(entity config.ChannelConfig) common.Channel
	MapConfigToModel(config config.ChannelConfig) common.Channel
	MapModelToConfig(model common.Channel) config.ChannelConfig
	MapEntityToConfig(entity config.ChannelConfig) config.ChannelConfig
}

type DefaultChannelMapper struct {
}

func NewChannelMapper() ChannelMapper {
	return &DefaultChannelMapper{}
}

func (mapper *DefaultChannelMapper) MapEntityToModel(entity config.ChannelConfig) common.Channel {
	return &model.Channel{
		ID:           entity.GetID(),
		ControllerID: entity.GetControllerID(),
		ChannelID:    entity.GetChannelID(),
		Name:         entity.GetName(),
		Enable:       entity.IsEnabled(),
		Notify:       entity.IsNotify(),
		Duration:     entity.GetDuration(),
		Debounce:     entity.GetDebounce(),
		Backoff:      entity.GetBackoff(),
		AlgorithmID:  entity.GetAlgorithmID()}
}

func (mapper *DefaultChannelMapper) MapConfigToModel(config config.ChannelConfig) common.Channel {
	return &model.Channel{
		ID:           config.GetID(),
		ControllerID: config.GetControllerID(),
		ChannelID:    config.GetChannelID(),
		Name:         config.GetName(),
		Enable:       config.IsEnabled(),
		Notify:       config.IsNotify(),
		Duration:     config.GetDuration(),
		Debounce:     config.GetDebounce(),
		Backoff:      config.GetBackoff(),
		AlgorithmID:  config.GetAlgorithmID()}
}

func (mapper *DefaultChannelMapper) MapModelToConfig(model common.Channel) config.ChannelConfig {
	return &config.Channel{
		ID:           model.GetID(),
		ControllerID: model.GetControllerID(),
		ChannelID:    model.GetChannelID(),
		Name:         model.GetName(),
		Enable:       model.IsEnabled(),
		Notify:       model.IsNotify(),
		Duration:     model.GetDuration(),
		Debounce:     model.GetDebounce(),
		Backoff:      model.GetBackoff(),
		AlgorithmID:  model.GetAlgorithmID()}
}

func (mapper *DefaultChannelMapper) MapEntityToConfig(entity config.ChannelConfig) config.ChannelConfig {
	return &model.Channel{
		ID:           entity.GetID(),
		ControllerID: entity.GetControllerID(),
		ChannelID:    entity.GetChannelID(),
		Name:         entity.GetName(),
		Enable:       entity.IsEnabled(),
		Notify:       entity.IsNotify(),
		Duration:     entity.GetDuration(),
		Debounce:     entity.GetDebounce(),
		Backoff:      entity.GetBackoff(),
		AlgorithmID:  entity.GetAlgorithmID()}
}
