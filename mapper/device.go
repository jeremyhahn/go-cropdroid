package mapper

import (
	"fmt"
	"sort"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/state"
)

type MicroDeviceMapper struct {
	metricMapper  MetricMapper
	channelMapper ChannelMapper
	DeviceMapper
}

type DeviceMapper interface {
	GetMetricMapper() MetricMapper
	GetChannelMapper() ChannelMapper
	MapStateToDevice(state state.DeviceStateMap,
		configuration *config.Device) (common.Device, error)
	MapConfigToModel(deviceEntity *config.Device,
		configEntities []*config.DeviceSetting) (common.Device, error)
}

func NewDeviceMapper(metricMapper MetricMapper, channelMapper ChannelMapper) DeviceMapper {
	return &MicroDeviceMapper{
		metricMapper:  metricMapper,
		channelMapper: channelMapper}
}

func (mapper *MicroDeviceMapper) GetMetricMapper() MetricMapper {
	return mapper.metricMapper
}

func (mapper *MicroDeviceMapper) GetChannelMapper() ChannelMapper {
	return mapper.channelMapper
}

func (mapper *MicroDeviceMapper) MapStateToDevice(state state.DeviceStateMap, device *config.Device) (common.Device, error) {
	configuredMetrics := device.GetMetrics()
	metrics := make([]common.Metric, len(configuredMetrics))
	for i, metricConfig := range configuredMetrics {
		if value, ok := state.GetMetrics()[metricConfig.GetKey()]; ok {
			metric := mapper.metricMapper.MapConfigToModel(metricConfig)
			metric.SetValue(value)
			metrics[i] = metric
		} else {
			return nil, fmt.Errorf("Unable to locate configured metric in device state! (device=%s, metric=%s)", device.GetType(), metricConfig.GetKey())
		}
	}
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].GetName() < metrics[j].GetName()
	})
	configuredChannels := device.GetChannels()
	channels := make([]common.Channel, len(configuredChannels))
	channelState := state.GetChannels()
	for i, channelConfig := range configuredChannels {
		channel := mapper.channelMapper.MapConfigToModel(channelConfig)
		channel.SetValue(channelState[channelConfig.GetChannelID()])
		channels[i] = channel
	}
	return &model.Device{
		ID:              device.GetID(),
		Type:            device.GetType(),
		Description:     device.GetDescription(),
		Enable:          device.IsEnabled(),
		Notify:          device.IsNotify(),
		URI:             device.GetURI(),
		HardwareVersion: device.GetHardwareVersion(),
		FirmwareVersion: device.GetFirmwareVersion(),
		Configs:         device.GetConfigMap(),
		Metrics:         metrics,
		Channels:        channels}, nil
}

func (mapper *MicroDeviceMapper) MapConfigToModel(deviceEntity *config.Device,
	settingEntities []*config.DeviceSetting) (common.Device, error) {

	configs := make(map[string]string, len(settingEntities))
	for _, entity := range settingEntities {
		configs[entity.GetKey()] = entity.GetValue()
	}
	return &model.Device{
		ID: deviceEntity.GetID(),
		//FarmID:      deviceEntity.GetFarmID(),
		Type:        deviceEntity.GetType(),
		Description: deviceEntity.GetDescription(),
		Configs:     configs,
		Metrics:     make([]common.Metric, 0),
		Channels:    make([]common.Channel, 0)}, nil
}
