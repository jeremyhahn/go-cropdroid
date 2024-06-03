package config

import (
	"fmt"
	"hash/fnv"
	"time"
)

type Schedule interface {
	GetID() uint64
	SetID(uint64)
	SetWorkflowID(id uint64)
	GetWorkflowID() uint64
	GetChannelID() uint64
	SetStartDate(time.Time)
	GetStartDate() time.Time
	SetEndDate(*time.Time)
	GetEndDate() *time.Time
	SetFrequency(int)
	GetFrequency() int
	SetInterval(int)
	GetInterval() int
	SetCount(int)
	GetCount() int
	SetDays(*string)
	GetDays() *string
	SetLastExecuted(time.Time)
	GetLastExecuted() time.Time
	SetExecutionCount(int)
	GetExecutionCount() int
	KeyValueEntity
}

type ScheduleStruct struct {
	ID             uint64     `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	WorkflowID     uint64     `yaml:"workflow" json:"workflow_id"`
	ChannelID      uint64     `yaml:"channelId" json:"channel_id"`
	StartDate      time.Time  `yaml:"startDate" json:"startDate"`
	EndDate        *time.Time `yaml:"endDate" json:"endDate"`
	Frequency      int        `yaml:"frequency" json:"frequency"`
	Interval       int        `yaml:"interval" json:"interval"`
	Count          int        `yaml:"count" json:"count"`
	Days           *string    `gorm:"type:varchar(50);default:NULL" yaml:"days" json:"days"`
	LastExecuted   time.Time  `gorm:"type:timestamp" yaml:"lastExecuted" json:"lastExecuted"`
	ExecutionCount int        `yaml:"executionCount" json:"executionCount"`
	Schedule       `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewSchedule() *ScheduleStruct {
	return &ScheduleStruct{}
}

func (schedule *ScheduleStruct) TableName() string {
	return "schedules"
}

func (schedule *ScheduleStruct) Identifier() uint64 {
	return schedule.ID
}

func (schedule *ScheduleStruct) SetID(id uint64) {
	schedule.ID = id
}

func (schedule *ScheduleStruct) SetWorkflowID(id uint64) {
	schedule.WorkflowID = id
}

func (schedule *ScheduleStruct) GetWorkflowID() uint64 {
	return schedule.WorkflowID
}

func (schedule *ScheduleStruct) SetChannelID(id uint64) {
	schedule.ChannelID = id
}

func (schedule *ScheduleStruct) GetChannelID() uint64 {
	return schedule.ChannelID
}

func (schedule *ScheduleStruct) SetStartDate(t time.Time) {
	schedule.StartDate = t
}

func (schedule *ScheduleStruct) GetStartDate() time.Time {
	return schedule.StartDate
}

func (schedule *ScheduleStruct) SetEndDate(t *time.Time) {
	schedule.EndDate = t
}

func (schedule *ScheduleStruct) GetEndDate() *time.Time {
	return schedule.EndDate
}

func (schedule *ScheduleStruct) SetFrequency(freq int) {
	schedule.Frequency = freq
}

func (schedule *ScheduleStruct) GetFrequency() int {
	return schedule.Frequency
}

func (schedule *ScheduleStruct) SetInterval(interval int) {
	schedule.Interval = interval
}

func (schedule *ScheduleStruct) GetInterval() int {
	return schedule.Interval
}

func (schedule *ScheduleStruct) SetCount(count int) {
	schedule.Count = count
}

func (schedule *ScheduleStruct) GetCount() int {
	return schedule.Count
}

func (schedule *ScheduleStruct) SetDays(days *string) {
	schedule.Days = days
}

func (schedule *ScheduleStruct) GetDays() *string {
	return schedule.Days
}

func (schedule *ScheduleStruct) GetLastExecuted() time.Time {
	return schedule.LastExecuted
}

func (schedule *ScheduleStruct) SetLastExecuted(dateTime time.Time) {
	schedule.LastExecuted = dateTime
}

func (schedule *ScheduleStruct) GetExecutionCount() int {
	return schedule.ExecutionCount
}

func (schedule *ScheduleStruct) SetExecutionCount(count int) {
	schedule.ExecutionCount = count
}

func (schedule *ScheduleStruct) Hash() uint64 {
	endDate := ""
	if schedule.EndDate != nil {
		endDate = schedule.EndDate.String()
	}
	key := fmt.Sprintf("%d-%d-%s-%d-%d-%d-%s-%s", schedule.ChannelID, schedule.Count, endDate, schedule.ExecutionCount,
		schedule.Frequency, schedule.Interval, schedule.LastExecuted.String(), schedule.StartDate.String())
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(key))
	return clusterHash.Sum64()
}

// func (schedule *ScheduleStruct) String() string {
// 	days := ""
// 	if schedule.Days != nil {
// 		days = days
// 	}
// 	endDate := ""
// 	if schedule.EndDate != nil {
// 		endDate = schedule.EndDate.String()
// 	}
// 	return fmt.Sprintf("%d-%d-%s-%s-%d-%d-%d-%s-%s-%d",
// 		schedule.WorkflowID, schedule.ChannelID,
// 		schedule.StartDate.String(), endDate,
// 		schedule.Frequency, schedule.Interval,
// 		schedule.Count, days, schedule.LastExecuted.String(),
// 		schedule.ExecutionCount)
// }
