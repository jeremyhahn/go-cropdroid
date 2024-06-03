package config

import (
	"fmt"
	"hash/fnv"
)

type Condition interface {
	SetWorkflowID(id uint64)
	GetWorkflowID() uint64
	SetChannelID(uint64)
	GetChannelID() uint64
	GetMetricID() uint64
	SetMetricID(uint64)
	SetComparator(string)
	GetComparator() string
	SetThreshold(float64)
	GetThreshold() float64
	KeyValueEntity
}

type ConditionStruct struct {
	ID         uint64  `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	WorkflowID uint64  `yaml:"workflow" json:"workflow_id"`
	ChannelID  uint64  `yaml:"channel" json:"channel_id"`
	MetricID   uint64  `yaml:"metric" json:"metric_id"`
	Comparator string  `yaml:"comparator" json:"comparator"`
	Threshold  float64 `yaml:"threshold" json:"threshold"`
	Condition  `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewCondition() *ConditionStruct {
	return &ConditionStruct{}
}

func (condition *ConditionStruct) TableName() string {
	return "conditions"
}

func (condition *ConditionStruct) Identifier() uint64 {
	return condition.ID
}

func (condition *ConditionStruct) SetID(id uint64) {
	condition.ID = id
}

func (condition *ConditionStruct) SetWorkflowID(id uint64) {
	condition.WorkflowID = id
}

func (condition *ConditionStruct) GetWorkflowID() uint64 {
	return condition.WorkflowID
}

func (condition *ConditionStruct) SetChannelID(id uint64) {
	condition.ChannelID = id
}

func (condition *ConditionStruct) GetChannelID() uint64 {
	return condition.ChannelID
}

func (condition *ConditionStruct) SetMetricID(metric uint64) {
	condition.MetricID = metric
}

func (condition *ConditionStruct) GetMetricID() uint64 {
	return condition.MetricID
}

func (condition *ConditionStruct) SetComparator(comp string) {
	condition.Comparator = comp
}

func (condition *ConditionStruct) GetComparator() string {
	return condition.Comparator
}

func (condition *ConditionStruct) SetThreshold(threshold float64) {
	condition.Threshold = threshold
}

func (condition *ConditionStruct) GetThreshold() float64 {
	return condition.Threshold
}

func (condition *ConditionStruct) Hash() uint64 {
	key := fmt.Sprintf("%d-%d-%s-%f", condition.GetChannelID(), condition.GetMetricID(),
		condition.GetComparator(), condition.GetThreshold())
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(key))
	return clusterHash.Sum64()
}

func (condition *ConditionStruct) String() string {
	return fmt.Sprintf("%d-%d-%d-%d-%s-%f",
		condition.ID, condition.WorkflowID, condition.ChannelID,
		condition.MetricID, condition.Comparator, condition.Threshold)
}
