package model

import (
	"time"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
)

type ControllerState struct {
	Metrics               []common.Metric  `json:"metrics"`
	Channels              []common.Channel `json:"channels"`
	Timestamp             time.Time        `json:"timestamp"`
	common.ControllerView `json:"-"`
}

func NewControllerState(app *app.App, metrics []common.Metric, channels []common.Channel) common.ControllerView {
	return &ControllerState{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: time.Now().In(app.Location)}
}

func CreateControllerState(metrics []common.Metric, channels []common.Channel, timestamp time.Time) common.ControllerView {
	return &ControllerState{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: timestamp}
}

func (state *ControllerState) SetMetrics(metrics []common.Metric) {
	state.Metrics = metrics
}

func (state *ControllerState) GetMetrics() []common.Metric {
	return state.Metrics
}

func (state *ControllerState) SetChannels(channels []common.Channel) {
	state.Channels = channels
}

func (state *ControllerState) GetChannels() []common.Channel {
	return state.Channels
}

func (state *ControllerState) GetTimestamp() time.Time {
	return state.Timestamp
}
