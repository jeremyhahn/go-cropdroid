package viewmodel

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
)

type DeviceViewModel struct {
	Metrics               []common.Metric  `json:"metrics"`
	Channels              []common.Channel `json:"channels"`
	Timestamp             time.Time        `json:"timestamp"`
	common.DeviceView `json:"-"`
}

func NewDeviceView(app *app.App, metrics []common.Metric, channels []common.Channel) common.DeviceView {
	return &DeviceViewModel{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: time.Now().In(app.Location)}
}

func CreateDeviceView(metrics []common.Metric, channels []common.Channel, timestamp time.Time) common.DeviceView {
	return &DeviceViewModel{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: timestamp}
}

func (device *DeviceViewModel) SetMetrics(metrics []common.Metric) {
	device.Metrics = metrics
}

func (device *DeviceViewModel) GetMetrics() []common.Metric {
	return device.Metrics
}

func (device *DeviceViewModel) SetChannels(channels []common.Channel) {
	device.Channels = channels
}

func (device *DeviceViewModel) GetChannels() []common.Channel {
	return device.Channels
}

func (device *DeviceViewModel) GetTimestamp() time.Time {
	return device.Timestamp
}
