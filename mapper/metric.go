package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type MetricMapper interface {
	MapConfigToModel(config config.MetricConfig) common.Metric
	//MapEntityToConfig(entity config.MetricConfig) config.MetricConfig
	MapEntityToModel(entity config.MetricConfig) common.Metric
	MapModelToConfig(model common.Metric) config.Metric
}

type DefaultMetricMapper struct {
}

func NewMetricMapper() MetricMapper {
	return &DefaultMetricMapper{}
}

func (mapper *DefaultMetricMapper) MapConfigToModel(config config.MetricConfig) common.Metric {
	return &model.Metric{
		ID:           config.GetID(),
		DeviceID: config.GetDeviceID(),
		DataType:     config.GetDataType(),
		Name:         config.GetName(),
		Key:          config.GetKey(),
		Enable:       config.IsEnabled(),
		Notify:       config.IsNotify(),
		Unit:         config.GetUnit(),
		AlarmLow:     config.GetAlarmLow(),
		AlarmHigh:    config.GetAlarmHigh()}
}

/*
func (mapper *DefaultMetricMapper) MapEntityToConfig(entity config.MetricConfig) config.MetricConfig {
	return &model.Metric{
		ID:           entity.GetID(),
		DeviceID: entity.GetDeviceID(),
		Name:         entity.GetName(),
		Key:          entity.GetKey(),
		Enable:       entity.IsEnabled(),
		Notify:       entity.IsNotify(),
		Unit:         entity.GetUnit(),
		AlarmLow:     entity.GetAlarmLow(),
		AlarmHigh:    entity.GetAlarmHigh()}
}*/

func (mapper *DefaultMetricMapper) MapEntityToModel(entity config.MetricConfig) common.Metric {
	return &model.Metric{
		ID:           entity.GetID(),
		DeviceID: entity.GetDeviceID(),
		DataType:     entity.GetDataType(),
		Name:         entity.GetName(),
		Key:          entity.GetKey(),
		Enable:       entity.IsEnabled(),
		Notify:       entity.IsNotify(),
		Unit:         entity.GetUnit(),
		AlarmLow:     entity.GetAlarmLow(),
		AlarmHigh:    entity.GetAlarmHigh()}
}

func (mapper *DefaultMetricMapper) MapModelToConfig(model common.Metric) config.Metric {
	return config.Metric{
		ID:           model.GetID(),
		DeviceID: model.GetDeviceID(),
		DataType:     model.GetDataType(),
		Name:         model.GetName(),
		Key:          model.GetKey(),
		Enable:       model.IsEnabled(),
		Notify:       model.IsNotify(),
		Unit:         model.GetUnit(),
		AlarmLow:     model.GetAlarmLow(),
		AlarmHigh:    model.GetAlarmHigh()}
}
