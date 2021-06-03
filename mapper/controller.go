package mapper

import (
	"fmt"
	"sort"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/state"
)

type MicroControllerMapper struct {
	metricMapper  MetricMapper
	channelMapper ChannelMapper
	ControllerMapper
}

type ControllerMapper interface {
	GetMetricMapper() MetricMapper
	GetChannelMapper() ChannelMapper
	MapStateToController(state state.ControllerStateMap, configuration config.Controller) (common.Controller, error)
	MapConfigToModel(controllerEntity config.ControllerConfig, configEntities []config.ControllerConfigItem) (common.Controller, error)
}

func NewControllerMapper(metricMapper MetricMapper, channelMapper ChannelMapper) ControllerMapper {
	return &MicroControllerMapper{
		metricMapper:  metricMapper,
		channelMapper: channelMapper}
}

func (mapper *MicroControllerMapper) GetMetricMapper() MetricMapper {
	return mapper.metricMapper
}

func (mapper *MicroControllerMapper) GetChannelMapper() ChannelMapper {
	return mapper.channelMapper
}

func (mapper *MicroControllerMapper) MapStateToController(state state.ControllerStateMap, controller config.Controller) (common.Controller, error) {
	configuredMetrics := controller.GetMetrics()
	metrics := make([]common.Metric, len(configuredMetrics))
	for i, metricConfig := range configuredMetrics {
		if value, ok := state.GetMetrics()[metricConfig.GetKey()]; ok {
			metric := mapper.metricMapper.MapConfigToModel(&metricConfig)
			metric.SetValue(value)
			metrics[i] = metric
		} else {
			return nil, fmt.Errorf("Unable to locate configured metric in controller state! (controller=%s, metric=%s)", controller.GetType(), metricConfig.GetKey())
		}
	}
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].GetName() < metrics[j].GetName()
	})
	configuredChannels := controller.GetChannels()
	channels := make([]common.Channel, len(configuredChannels))
	channelState := state.GetChannels()
	for i, channelConfig := range configuredChannels {
		channel := mapper.channelMapper.MapConfigToModel(&channelConfig)
		channel.SetValue(channelState[channelConfig.ChannelID])
		channels[i] = channel
	}
	return &model.Controller{
		ID:          controller.GetID(),
		Type:        controller.GetType(),
		Description: controller.GetDescription(),
		Enable:      controller.IsEnabled(),
		Notify:      controller.IsNotify(),
		URI:         controller.GetURI(),
		Configs:     controller.GetConfigMap(),
		Metrics:     metrics,
		Channels:    channels}, nil
}

func (mapper *MicroControllerMapper) MapConfigToModel(controllerEntity config.ControllerConfig,
	configEntities []config.ControllerConfigItem) (common.Controller, error) {

	configs := make(map[string]string, len(configEntities))
	for _, entity := range configEntities {
		configs[entity.GetKey()] = entity.GetValue()
	}
	return &model.Controller{
		ID: controllerEntity.GetID(),
		//FarmID:      controllerEntity.GetFarmID(),
		Type:        controllerEntity.GetType(),
		Description: controllerEntity.GetDescription(),
		Configs:     configs,
		Metrics:     make([]common.Metric, 0),
		Channels:    make([]common.Channel, 0)}, nil
}
