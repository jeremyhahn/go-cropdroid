package viewmodel

import (
	"fmt"
)

type Condition struct {
	ID         uint64  `json:"id"`
	DeviceType string  `json:"deviceType"`
	MetricID   uint64  `json:"metricId"`
	MetricName string  `json:"metricName"`
	WorkflowID uint64  `json:"workflowId"`
	ChannelID  uint64  `json:"channelId"`
	Comparator string  `json:"comparator"`
	Threshold  float64 `json:"threshold"`
	Text       string  `json:"text"`
}

func (condition *Condition) GetID() uint64 {
	return condition.ID
}

func (condition *Condition) GetDeviceType() string {
	return condition.DeviceType
}

func (condition *Condition) GetMetricID() uint64 {
	return condition.MetricID
}

func (condition *Condition) GetMetricName() string {
	return condition.MetricName
}

func (condition *Condition) GetWorkflowID() uint64 {
	return condition.WorkflowID
}

func (condition *Condition) GetChannelID() uint64 {
	return condition.ChannelID
}

func (condition *Condition) GetComparator() string {
	return condition.Comparator
}

func (condition *Condition) GetThreshold() float64 {
	return condition.Threshold
}

func (condition *Condition) GetText() string {
	return condition.Text
}

func (condition *Condition) String() string {
	return fmt.Sprintf("%s %s %s %.2f", condition.DeviceType, condition.MetricName, condition.Comparator, condition.Threshold)
}
