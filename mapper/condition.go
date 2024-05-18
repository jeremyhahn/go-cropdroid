package mapper

import (
	"fmt"
	"strings"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type ConditionMapper interface {
	MapConfigToView(entity *config.Condition, deviceType string,
		metric *config.Metric, channelID uint64) *viewmodel.Condition
	MapViewToConfig(viewModel viewmodel.Condition) *config.Condition
}

type DefaultConditionMapper struct {
}

func NewConditionMapper() ConditionMapper {
	return &DefaultConditionMapper{}
}

func (mapper *DefaultConditionMapper) MapConfigToView(entity *config.Condition, deviceType string,
	metric *config.Metric, channelID uint64) *viewmodel.Condition {

	text := fmt.Sprintf("%s %s %s %.2f",
		strings.Title(deviceType),
		strings.ToLower(metric.GetName()),
		//mapper.comparatorToText(entity.GetComparator()),
		entity.GetComparator(),
		entity.GetThreshold())
	return &viewmodel.Condition{
		ID:         entity.ID,
		DeviceType: deviceType,
		MetricID:   metric.ID,
		MetricName: metric.GetName(),
		WorkflowID: entity.GetWorkflowID(),
		ChannelID:  channelID,
		Comparator: entity.GetComparator(),
		Threshold:  entity.GetThreshold(),
		Text:       text}
}

func (mapper *DefaultConditionMapper) MapViewToConfig(viewModel viewmodel.Condition) *config.Condition {
	return &config.Condition{
		ID:         viewModel.ID,
		WorkflowID: viewModel.GetWorkflowID(),
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
