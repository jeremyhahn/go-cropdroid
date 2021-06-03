package config

import (
	"fmt"
	"hash/fnv"
	"time"
)

type Schedule struct {
	ID             uint64     `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	ChannelID      int        `yaml:"channelId" json:"channel_id"`
	StartDate      time.Time  `yaml:"startDate" json:"startDate"`
	EndDate        *time.Time `yaml:"endDate" json:"endDate"`
	Frequency      int        `yaml:"frequency" json:"frequency"`
	Interval       int        `yaml:"interval" json:"interval"`
	Count          int        `yaml:"count" json:"count"`
	Days           *string    `gorm:"type:varchar(50);default:NULL" yaml:"days" json:"days"`
	LastExecuted   time.Time  `gorm:"type:timestamp" yaml:"lastExecuted" json:"lastExecuted"`
	ExecutionCount int        `yaml:"executionCount" json:"executionCount"`
	ScheduleConfig `yaml:"-" json:"-"`
}

func NewSchedule() *Schedule {
	return &Schedule{}
}

func (schedule *Schedule) GetID() uint64 {
	return schedule.ID
}

func (schedule *Schedule) SetID(id uint64) {
	schedule.ID = id
}

func (schedule *Schedule) SetChannelID(id int) {
	schedule.ChannelID = id
}

func (schedule *Schedule) GetChannelID() int {
	return schedule.ChannelID
}

func (schedule *Schedule) SetStartDate(t time.Time) {
	schedule.StartDate = t
}

func (schedule *Schedule) GetStartDate() time.Time {
	return schedule.StartDate
}

func (schedule *Schedule) SetEndDate(t *time.Time) {
	schedule.EndDate = t
}

func (schedule *Schedule) GetEndDate() *time.Time {
	return schedule.EndDate
}

func (schedule *Schedule) SetFrequency(freq int) {
	schedule.Frequency = freq
}

func (schedule *Schedule) GetFrequency() int {
	return schedule.Frequency
}

func (schedule *Schedule) SetInterval(interval int) {
	schedule.Interval = interval
}

func (schedule *Schedule) GetInterval() int {
	return schedule.Interval
}

func (schedule *Schedule) SetCount(count int) {
	schedule.Count = count
}

func (schedule *Schedule) GetCount() int {
	return schedule.Count
}

func (schedule *Schedule) SetDays(days *string) {
	schedule.Days = days
}

func (schedule *Schedule) GetDays() *string {
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

func (schedule *Schedule) Hash() uint64 {
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
