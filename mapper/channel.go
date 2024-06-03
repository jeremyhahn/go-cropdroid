package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type ChannelMapper interface {
	MapConfigToModel(channel config.Channel) model.Channel
	MapModelToConfig(model model.Channel) config.Channel
}

type ChannelMapperStruct struct {
	ChannelMapper
}

func NewChannelMapper() ChannelMapper {
	return &ChannelMapperStruct{}
}

func (mapper *ChannelMapperStruct) MapConfigToModel(channel config.Channel) model.Channel {
	return &model.ChannelStruct{
		ID:          channel.Identifier(),
		DeviceID:    channel.GetDeviceID(),
		BoardID:     channel.GetBoardID(),
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

func (mapper *ChannelMapperStruct) MapModelToConfig(model model.Channel) config.Channel {
	return &config.ChannelStruct{
		ID:          model.Identifier(),
		DeviceID:    model.GetDeviceID(),
		BoardID:     model.GetBoardID(),
		Name:        model.GetName(),
		Enable:      model.IsEnabled(),
		Notify:      model.IsNotify(),
		Duration:    model.GetDuration(),
		Debounce:    model.GetDebounce(),
		Backoff:     model.GetBackoff(),
		AlgorithmID: model.GetAlgorithmID(),
		Conditions:  make([]*config.ConditionStruct, 0),
		Schedule:    make([]*config.ScheduleStruct, 0)}
}
