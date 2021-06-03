package viewmodel

import (
	"time"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
)

type ControllerViewModel struct {
	Metrics               []common.Metric  `json:"metrics"`
	Channels              []common.Channel `json:"channels"`
	Timestamp             time.Time        `json:"timestamp"`
	common.ControllerView `json:"-"`
}

func NewControllerView(app *app.App, metrics []common.Metric, channels []common.Channel) common.ControllerView {
	return &ControllerViewModel{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: time.Now().In(app.Location)}
}

func CreateControllerView(metrics []common.Metric, channels []common.Channel, timestamp time.Time) common.ControllerView {
	return &ControllerViewModel{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: timestamp}
}

func (controller *ControllerViewModel) SetMetrics(metrics []common.Metric) {
	controller.Metrics = metrics
}

func (controller *ControllerViewModel) GetMetrics() []common.Metric {
	return controller.Metrics
}

func (controller *ControllerViewModel) SetChannels(channels []common.Channel) {
	controller.Channels = channels
}

func (controller *ControllerViewModel) GetChannels() []common.Channel {
	return controller.Channels
}

func (controller *ControllerViewModel) GetTimestamp() time.Time {
	return controller.Timestamp
}
