package model

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/config"
)

type Schedule struct {
	ID                    int        `yaml:"id" json:"id"`
	ChannelID             int        `yaml:"channelId" json:"channelId"`
	StartDate             time.Time  `yaml:"startDate" json:"startDate"`
	EndDate               *time.Time `yaml:"endDate" json:"endDate"`
	Frequency             int        `yaml:"frequency" json:"frequency"`
	Interval              int        `yaml:"interval" json:"interval"`
	Count                 int        `yaml:"count" json:"count"`
	Days                  []string   `yaml:"days" json:"days"`
	LastExecuted          time.Time  `yaml:"lastExecuted" json:"lastExecuted"`
	ExecutionCount        int        `yaml:"executionCount" json:"executionCount"`
	config.ScheduleConfig `yaml:"-" json:"-"`
}

func (schedule *Schedule) GetID() int {
	return schedule.ID
}

func (schedule *Schedule) SetID(id int) {
	schedule.ID = id
}

func (schedule *Schedule) GetChannelID() int {
	return schedule.ChannelID
}

func (schedule *Schedule) GetStartDate() time.Time {
	return schedule.StartDate
}

func (schedule *Schedule) GetEndDate() *time.Time {
	return schedule.EndDate
}

func (schedule *Schedule) GetFrequency() int {
	return schedule.Frequency
}

func (schedule *Schedule) GetInterval() int {
	return schedule.Interval
}

func (schedule *Schedule) GetCount() int {
	return schedule.Count
}

func (schedule *Schedule) GetDays() []string {
	return schedule.Days
}

func (schedule *Schedule) GetLastExecuted() time.Time {
	return schedule.LastExecuted
}

func (schedule *Schedule) SetLastExecuted(dateTime time.Time) {
	schedule.LastExecuted = dateTime
}

func (schedule *Schedule) GetExecutionCount() int {
	return schedule.ExecutionCount
}

func (schedule *Schedule) SetExecutionCount(count int) {
	schedule.ExecutionCount = count
}
