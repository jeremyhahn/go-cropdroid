package mapper

import (
	"fmt"
	"strings"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type ConditionMapper interface {
	MapConfigToView(entity *config.ConditionStruct, deviceType string,
		metric *config.MetricStruct, channelID uint64) *viewmodel.Condition
	MapViewToConfig(viewModel viewmodel.Condition) *config.ConditionStruct
}

type ConditionMapperStruct struct {
}

func NewConditionMapper() ConditionMapper {
	return &ConditionMapperStruct{}
}

func (mapper *ConditionMapperStruct) MapConfigToView(entity *config.ConditionStruct, deviceType string,
	metric *config.MetricStruct, channelID uint64) *viewmodel.Condition {

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

func (mapper *ConditionMapperStruct) MapViewToConfig(viewModel viewmodel.Condition) *config.ConditionStruct {
	return &config.ConditionStruct{
		ID:         viewModel.ID,
		WorkflowID: viewModel.GetWorkflowID(),
		MetricID:   viewModel.GetMetricID(),
		ChannelID:  viewModel.GetChannelID(),
		Comparator: viewModel.GetComparator(),
		Threshold:  viewModel.GetThreshold()}
}

// func (mapper *ConditionMapperStruct) comparatorToText(comparator string) string {
// 	switch comparator {
// 	case ">":
// 		return "is greater than"
// 	case ">=":
// 		return "is greater than or equal to"
// 	case "<":
// 		return "is less than"
// 	case "<=":
// 		return "is less than or equal to"
// 	case "=":
// 		return "is equal to"
// 	}
// 	return ""
// }
