package mapper

import (
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/model"
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
		ControllerID: config.GetControllerID(),
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
		ControllerID: entity.GetControllerID(),
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
		ControllerID: entity.GetControllerID(),
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
		ControllerID: model.GetControllerID(),
		DataType:     model.GetDataType(),
		Name:         model.GetName(),
		Key:          model.GetKey(),
		Enable:       model.IsEnabled(),
		Notify:       model.IsNotify(),
		Unit:         model.GetUnit(),
		AlarmLow:     model.GetAlarmLow(),
		AlarmHigh:    model.GetAlarmHigh()}
}
