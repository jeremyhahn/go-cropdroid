package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type MetricMapper interface {
	MapConfigToModel(config *config.MetricStruct) model.Metric
	MapModelToConfig(model model.Metric) *config.MetricStruct
}

type DefaultMetricMapper struct {
}

func NewMetricMapper() MetricMapper {
	return &DefaultMetricMapper{}
}

func (mapper *DefaultMetricMapper) MapConfigToModel(config *config.MetricStruct) model.Metric {
	return &model.MetricStruct{
		ID:        config.Identifier(),
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

func (mapper *DefaultMetricMapper) MapModelToConfig(model model.Metric) *config.MetricStruct {
	return &config.MetricStruct{
		ID:        model.Identifier(),
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
