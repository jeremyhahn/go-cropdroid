package model

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
)

type DeviceState struct {
	Metrics               []common.Metric  `json:"metrics"`
	Channels              []common.Channel `json:"channels"`
	Timestamp             time.Time        `json:"timestamp"`
	common.DeviceView `json:"-"`
}

func NewDeviceState(app *app.App, metrics []common.Metric, channels []common.Channel) common.DeviceView {
	return &DeviceState{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: time.Now().In(app.Location)}
}

func CreateDeviceState(metrics []common.Metric, channels []common.Channel, timestamp time.Time) common.DeviceView {
	return &DeviceState{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: timestamp}
}

func (state *DeviceState) SetMetrics(metrics []common.Metric) {
	state.Metrics = metrics
}

func (state *DeviceState) GetMetrics() []common.Metric {
	return state.Metrics
}

func (state *DeviceState) SetChannels(channels []common.Channel) {
	state.Channels = channels
}

func (state *DeviceState) GetChannels() []common.Channel {
	return state.Channels
}

func (state *DeviceState) GetTimestamp() time.Time {
	return state.Timestamp
}
