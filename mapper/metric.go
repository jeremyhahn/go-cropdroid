package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type MetricMapper interface {
	MapConfigToModel(config *config.Metric) common.Metric
	MapModelToConfig(model common.Metric) *config.Metric
}

type DefaultMetricMapper struct {
}

func NewMetricMapper() MetricMapper {
	return &DefaultMetricMapper{}
}

func (mapper *DefaultMetricMapper) MapConfigToModel(config *config.Metric) common.Metric {
	return &model.Metric{
		ID:        config.GetID(),
		DeviceID:  config.GetDeviceID(),
		DataType:  config.GetDataType(),
		Name:      config.GetName(),
		Key:       config.GetKey(),
		Enable:    config.IsEnabled(),
		Notify:    config.IsNotify(),
		Unit:      config.GetUnit(),
		AlarmLow:  config.GetAlarmLow(),
		AlarmHigh: config.GetAlarmHigh()}
}

func (mapper *DefaultMetricMapper) MapModelToConfig(model common.Metric) *config.Metric {
	return &config.Metric{
		ID:        model.GetID(),
		DeviceID:  model.GetDeviceID(),
		DataType:  model.GetDataType(),
		Name:      model.GetName(),
		Key:       model.GetKey(),
		Enable:    model.IsEnabled(),
		Notify:    model.IsNotify(),
		Unit:      model.GetUnit(),
		AlarmLow:  model.GetAlarmLow(),
		AlarmHigh: model.GetAlarmHigh()}
}
