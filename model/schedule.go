package model

import (
	"time"
)

type Schedule interface {
	GetWorkflowID()
	GetChannelID()
	GetStartDate()
	GetEndDate()
	GetFrequency()
	GetInterval()
	GetCount()
	GetDays()
	GetLastExecuted()
	GetExecutionCount()
}

type ScheduleStruct struct {
	ID             int        `yaml:"id" json:"id"`
	ChannelID      int        `yaml:"channelId" json:"channelId"`
	StartDate      time.Time  `yaml:"startDate" json:"startDate"`
	EndDate        *time.Time `yaml:"endDate" json:"endDate"`
	Frequency      int        `yaml:"frequency" json:"frequency"`
	Interval       int        `yaml:"interval" json:"interval"`
	Count          int        `yaml:"count" json:"count"`
	Days           []string   `yaml:"days" json:"days"`
	LastExecuted   time.Time  `yaml:"lastExecuted" json:"lastExecuted"`
	ExecutionCount int        `yaml:"executionCount" json:"executionCount"`
	Schedule       `yaml:"-" json:"-"`
}

func (schedule *ScheduleStruct) Identifier() int {
	return schedule.ID
}

func (schedule *ScheduleStruct) SetID(id int) {
	schedule.ID = id
}

func (schedule *ScheduleStruct) GetChannelID() int {
	return schedule.ChannelID
}

func (schedule *ScheduleStruct) GetStartDate() time.Time {
	return schedule.StartDate
}

func (schedule *ScheduleStruct) GetEndDate() *time.Time {
	return schedule.EndDate
}

func (schedule *ScheduleStruct) GetFrequency() int {
	return schedule.Frequency
}

func (schedule *ScheduleStruct) GetInterval() int {
	return schedule.Interval
}

func (schedule *ScheduleStruct) GetCount() int {
	return schedule.Count
}

func (schedule *ScheduleStruct) GetDays() []string {
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
