package mapper

import (
	"fmt"
	"sort"

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
	MapStateToDevice(state state.DeviceStateMap, configuration config.Device) (model.Device, error)
	MapConfigToModel(deviceEntity config.Device) model.Device
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

func (mapper *MicroDeviceMapper) MapStateToDevice(state state.DeviceStateMap, device config.Device) (model.Device, error) {
	configuredMetrics := device.GetMetrics()
	metrics := make([]model.Metric, len(configuredMetrics))
	for i, metricConfig := range configuredMetrics {
		if value, ok := state.GetMetrics()[metricConfig.GetKey()]; ok {
			metric := mapper.metricMapper.MapConfigToModel(metricConfig)
			metric.SetValue(value)
			metrics[i] = metric
		} else {
			return nil, fmt.Errorf(
				"Unable to locate configured metric in device state! (device=%s, metric=%s)",
				device.GetType(),
				metricConfig.GetKey())
		}
	}
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].GetName() < metrics[j].GetName()
	})
	configuredChannels := device.GetChannels()
	channels := make([]model.Channel, len(configuredChannels))
	channelState := state.GetChannels()
	for i, channelConfig := range configuredChannels {
		channel := mapper.channelMapper.MapConfigToModel(channelConfig)
		channel.SetValue(channelState[channelConfig.GetBoardID()])
		channels[i] = channel
	}
	return &model.DeviceStruct{
		ID:              device.Identifier(),
		Type:            device.GetType(),
		Description:     device.GetDescription(),
		Enable:          device.IsEnabled(),
		Notify:          device.IsNotify(),
		URI:             device.GetURI(),
		HardwareVersion: device.GetHardwareVersion(),
		FirmwareVersion: device.GetFirmwareVersion(),
		Settings:        device.GetSettingsMap(),
		Metrics:         metrics,
		Channels:        channels}, nil
}

func (mapper *MicroDeviceMapper) MapConfigToModel(deviceEntity config.Device) model.Device {

	deviceEntitySettings := deviceEntity.GetSettings()
	settings := make(map[string]string, len(deviceEntitySettings))
	for _, entity := range deviceEntitySettings {
		settings[entity.GetKey()] = entity.GetValue()
	}
	return &model.DeviceStruct{
		ID: deviceEntity.Identifier(),
		//FarmID:      deviceEntity.GetFarmID(),
		Type:        deviceEntity.GetType(),
		Description: deviceEntity.GetDescription(),
		Settings:    settings,
		Metrics:     make([]model.Metric, 0),
		Channels:    make([]model.Channel, 0)}
}
