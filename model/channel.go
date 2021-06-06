package model

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
)

type Channel struct {
	ID             int                `yaml:"id" json:"id"`
	ControllerID   int                `yaml:"controllerID" json:"controllerId"`
	ChannelID      int                `yaml:"channelId" json:"channelId"`
	Name           string             `yaml:"name" json:"name"`
	Enable         bool               `yaml:"enable" json:"enable"`
	Notify         bool               `yaml:"notify" json:"notify"`
	Conditions     []config.Condition `yaml:"conditions" json:"conditions"`
	Schedule       []config.Schedule  `yaml:"schedule" json:"schedule"`
	Duration       int                `yaml:"duration" json:"duration"`
	Debounce       int                `yaml:"debounce" json:"debounce"`
	Backoff        int                `yaml:"backoff" json:"backoff"`
	AlgorithmID    int                `yaml:"algorithm" json:"algorithmId"`
	Value          int                `yaml:"value" json:"value"`
	config.Channel `json:"-"`
}

func NewChannel() common.Channel {
	return &Channel{}
}

func (channel *Channel) SetID(id int) {
	channel.ID = id
}

func (channel *Channel) GetID() int {
	return channel.ID
}

func (channel *Channel) GetControllerID() int {
	return channel.ControllerID
}

func (channel *Channel) SetControllerID(id int) {
	channel.ControllerID = id
}

func (channel *Channel) SetChannelID(id int) {
	channel.ChannelID = id
}

func (channel *Channel) GetChannelID() int {
	return channel.ChannelID
}

func (channel *Channel) SetName(name string) {
	channel.Name = name
}

func (channel *Channel) GetName() string {
	return channel.Name
}

func (channel *Channel) SetEnable(enable bool) {
	channel.Enable = enable
}

func (channel *Channel) IsEnabled() bool {
	return channel.Enable
}

func (channel *Channel) SetNotify(notify bool) {
	channel.Notify = notify
}

func (channel *Channel) IsNotify() bool {
	return channel.Notify
}

func (channel *Channel) SetConditions(conditions []config.Condition) {
	channel.Conditions = conditions
}

func (channel *Channel) GetConditions() []config.Condition {
	return channel.Conditions
}

func (channel *Channel) SetSchedule(schedule []config.Schedule) {
	channel.Schedule = schedule
}

func (channel *Channel) GetSchedule() []config.Schedule {
	return channel.Schedule
}

func (channel *Channel) SetDuration(duration int) {
	channel.Duration = duration
}

func (channel *Channel) GetDuration() int {
	return channel.Duration
}

func (channel *Channel) SetDebounce(debounce int) {
	channel.Debounce = debounce
}

func (channel *Channel) GetDebounce() int {
	return channel.Debounce
}

func (channel *Channel) SetBackoff(backoff int) {
	channel.Backoff = backoff
}

func (channel *Channel) GetBackoff() int {
	return channel.Backoff
}

func (channel *Channel) SetAlgorithmID(id int) {
	channel.AlgorithmID = id
}

func (channel *Channel) GetAlgorithmID() int {
	return channel.AlgorithmID
}

func (channel *Channel) SetValue(value int) {
	channel.Value = value
}

func (channel *Channel) GetValue() int {
	return channel.Value
}
