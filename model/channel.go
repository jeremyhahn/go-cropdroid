package model

import (
	"github.com/jeremyhahn/go-cropdroid/config"
)

type Channel interface {
	AddCondition(condition config.Condition)
	GetConditions() []config.Condition
	SetConditions(conditions []config.Condition)
	SetCondition(condition config.Condition)
	GetSchedule() []config.Schedule
	SetSchedule(schedule []config.Schedule)
	SetScheduleItem(schedule config.Schedule)
	SetValue(value int)
	GetValue() int
	config.CommonChannel
}

// The Channel model is a fully populated Channel that contains
// the config and current state.
type ChannelStruct struct {
	ID          uint64             `yaml:"id" json:"id"`
	DeviceID    uint64             `yaml:"deviceID" json:"deviceId"`
	BoardID     int                `yaml:"boardId" json:"boardId"`
	Name        string             `yaml:"name" json:"name"`
	Enable      bool               `yaml:"enable" json:"enable"`
	Notify      bool               `yaml:"notify" json:"notify"`
	Conditions  []config.Condition `yaml:"conditions" json:"conditions"`
	Schedule    []config.Schedule  `yaml:"schedule" json:"schedule"`
	Duration    int                `yaml:"duration" json:"duration"`
	Debounce    int                `yaml:"debounce" json:"debounce"`
	Backoff     int                `yaml:"backoff" json:"backoff"`
	AlgorithmID uint64             `yaml:"algorithm" json:"algorithmId"`
	Value       int                `yaml:"value" json:"value"`
	Channel     `json:"-" yaml:"-"`
}

func NewChannel() Channel {
	return &ChannelStruct{
		Conditions: make([]config.Condition, 0),
		Schedule:   make([]config.Schedule, 0)}
}

func (channel *ChannelStruct) SetID(id uint64) {
	channel.ID = id
}

func (channel *ChannelStruct) Identifier() uint64 {
	return channel.ID
}

func (channel *ChannelStruct) GetDeviceID() uint64 {
	return channel.DeviceID
}

func (channel *ChannelStruct) SetDeviceID(id uint64) {
	channel.DeviceID = id
}

func (channel *ChannelStruct) SetBoardID(id int) {
	channel.BoardID = id
}

func (channel *ChannelStruct) GetBoardID() int {
	return channel.BoardID
}

func (channel *ChannelStruct) SetName(name string) {
	channel.Name = name
}

func (channel *ChannelStruct) GetName() string {
	return channel.Name
}

func (channel *ChannelStruct) SetEnable(enable bool) {
	channel.Enable = enable
}

func (channel *ChannelStruct) IsEnabled() bool {
	return channel.Enable
}

func (channel *ChannelStruct) SetNotify(notify bool) {
	channel.Notify = notify
}

func (channel *ChannelStruct) IsNotify() bool {
	return channel.Notify
}

func (channel *ChannelStruct) AddCondition(condition config.Condition) {
	channel.Conditions = append(channel.Conditions, condition)
}

func (channel *ChannelStruct) SetCondition(condition config.Condition) {
	for i, c := range channel.Conditions {
		if c.Identifier() == condition.Identifier() {
			channel.Conditions[i] = condition
			return
		}
	}
	channel.Conditions = append(channel.Conditions, condition)
}

func (channel *ChannelStruct) SetConditions(conditions []config.Condition) {
	channel.Conditions = conditions
}

func (channel *ChannelStruct) GetConditions() []config.Condition {
	return channel.Conditions
}

func (channel *ChannelStruct) SetSchedule(schedule []config.Schedule) {
	channel.Schedule = schedule
}

func (channel *ChannelStruct) GetSchedule() []config.Schedule {
	scheduleConfigs := make([]config.Schedule, len(channel.Schedule))
	copy(scheduleConfigs, channel.Schedule)
	// for i, sched := range channel.Schedule {
	// 	scheduleConfigs[i] = sched
	// }
	return scheduleConfigs
}

func (channel *ChannelStruct) SetScheduleItem(schedule config.Schedule) {
	for i, s := range channel.Schedule {
		if s.Identifier() == schedule.Identifier() {
			channel.Schedule[i] = schedule
			return
		}
	}
	channel.Schedule = append(channel.Schedule, schedule)
}

func (channel *ChannelStruct) SetDuration(duration int) {
	channel.Duration = duration
}

func (channel *ChannelStruct) GetDuration() int {
	return channel.Duration
}

func (channel *ChannelStruct) SetDebounce(debounce int) {
	channel.Debounce = debounce
}

func (channel *ChannelStruct) GetDebounce() int {
	return channel.Debounce
}

func (channel *ChannelStruct) SetBackoff(backoff int) {
	channel.Backoff = backoff
}

func (channel *ChannelStruct) GetBackoff() int {
	return channel.Backoff
}

func (channel *ChannelStruct) SetAlgorithmID(id uint64) {
	channel.AlgorithmID = id
}

func (channel *ChannelStruct) GetAlgorithmID() uint64 {
	return channel.AlgorithmID
}

func (channel *ChannelStruct) SetValue(value int) {
	channel.Value = value
}

func (channel *ChannelStruct) GetValue() int {
	return channel.Value
}
