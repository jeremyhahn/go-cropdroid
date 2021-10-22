package config

import (
	"fmt"
	"hash/fnv"
)

type Condition struct {
	ID              uint64  `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	WorkflowID      uint64  `yaml:"workflow" json:"workflow_id"`
	ChannelID       uint64  `yaml:"channel" json:"channel_id"`
	MetricID        uint64  `yaml:"metric" json:"metric_id"`
	Comparator      string  `yaml:"comparator" json:"comparator"`
	Threshold       float64 `yaml:"threshold" json:"threshold"`
	ConditionConfig `yaml:"-" json:"-"`
}

func NewCondition() *Condition {
	return &Condition{}
}

func (condition *Condition) GetID() uint64 {
	return condition.ID
}

func (condition *Condition) SetID(id uint64) {
	condition.ID = id
}

func (condition *Condition) SetWorkflowID(id uint64) {
	condition.WorkflowID = id
}

func (condition *Condition) GetWorkflowID() uint64 {
	return condition.WorkflowID
}

func (condition *Condition) SetChannelID(id uint64) {
	condition.ChannelID = id
}

func (condition *Condition) GetChannelID() uint64 {
	return condition.ChannelID
}

func (condition *Condition) SetMetricID(metric uint64) {
	condition.MetricID = metric
}

func (condition *Condition) GetMetricID() uint64 {
	return condition.MetricID
}

func (condition *Condition) SetComparator(comp string) {
	condition.Comparator = comp
}

func (condition *Condition) GetComparator() string {
	return condition.Comparator
}

func (condition *Condition) SetThreshold(threshold float64) {
	condition.Threshold = threshold
}

func (condition *Condition) GetThreshold() float64 {
	return condition.Threshold
}

func (condition *Condition) Hash() uint64 {
	key := fmt.Sprintf("%d-%d-%s-%f", condition.GetChannelID(), condition.GetMetricID(),
		condition.GetComparator(), condition.GetThreshold())
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(key))
	return clusterHash.Sum64()
}
