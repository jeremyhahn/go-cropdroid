package mapper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type ConditionMapper interface {
	MapModelToEntity(model config.ConditionConfig) *config.Condition
	MapEntityToModel(entity *config.Condition) *config.Condition
	MapEntityToView(entity config.ConditionConfig, deviceType string, metric config.MetricConfig, channelID int) *viewmodel.Condition
	MapViewToConfig(viewModel viewmodel.Condition) config.ConditionConfig
}

type DefaultConditionMapper struct {
}

func NewConditionMapper() ConditionMapper {
	return &DefaultConditionMapper{}
}

func (mapper *DefaultConditionMapper) MapModelToEntity(model config.ConditionConfig) *config.Condition {
	return &config.Condition{
		ID:         model.GetID(),
		ChannelID:  model.GetChannelID(),
		MetricID:   model.GetMetricID(),
		Comparator: model.GetComparator(),
		Threshold:  model.GetThreshold()}
}

func (mapper *DefaultConditionMapper) MapEntityToModel(entity *config.Condition) *config.Condition {
	return &config.Condition{
		ID:         entity.GetID(),
		ChannelID:  entity.GetChannelID(),
		MetricID:   entity.GetMetricID(),
		Comparator: entity.GetComparator(),
		Threshold:  entity.GetThreshold()}
}

func (mapper *DefaultConditionMapper) MapEntityToView(entity config.ConditionConfig, deviceType string,
	metric config.MetricConfig, channelID int) *viewmodel.Condition {

	text := fmt.Sprintf("%s %s %s %.2f",
		strings.Title(deviceType),
		strings.ToLower(metric.GetName()),
		//mapper.comparatorToText(entity.GetComparator()),
		entity.GetComparator(),
		entity.GetThreshold())
	return &viewmodel.Condition{
		ID:             fmt.Sprintf("%d", entity.GetID()),
		DeviceType: deviceType,
		MetricID:       metric.GetID(),
		MetricName:     metric.GetName(),
		ChannelID:      channelID,
		Comparator:     entity.GetComparator(),
		Threshold:      entity.GetThreshold(),
		Text:           text}
}

func (mapper *DefaultConditionMapper) MapViewToConfig(viewModel viewmodel.Condition) config.ConditionConfig {
	id, _ := strconv.ParseUint(viewModel.GetID(), 10, 64)
	return &config.Condition{
		ID:         id,
		MetricID:   viewModel.GetMetricID(),
		ChannelID:  viewModel.GetChannelID(),
		Comparator: viewModel.GetComparator(),
		Threshold:  viewModel.GetThreshold()}
}

/*
func (mapper *DefaultConditionMapper) comparatorToText(comparator string) string {
	switch comparator {
	case ">":
		return "is greater than"
	case ">=":
		return "is greater than or equal to"
	case "<":
		return "is less than"
	case "<=":
		return "is less than or equal to"
	case "=":
		return "is equal to"
	}
	return ""
}
*/
