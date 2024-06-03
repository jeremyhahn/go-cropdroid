package viewmodel

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type DeviceView interface {
	GetMetrics() []model.Metric
	GetChannels() []model.Channel
	GetTimestamp() time.Time
}

type DeviceViewModel struct {
	Metrics    []model.Metric  `json:"metrics"`
	Channels   []model.Channel `json:"channels"`
	Timestamp  time.Time       `json:"timestamp"`
	DeviceView `json:"-"`
}

func NewDeviceView(app *app.App, metrics []model.Metric, channels []model.Channel) DeviceView {
	return &DeviceViewModel{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: time.Now().In(app.Location)}
}

func CreateDeviceView(metrics []model.Metric, channels []model.Channel, timestamp time.Time) DeviceView {
	return &DeviceViewModel{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: timestamp}
}

func (device *DeviceViewModel) SetMetrics(metrics []model.Metric) {
	device.Metrics = metrics
}

func (device *DeviceViewModel) GetMetrics() []model.Metric {
	return device.Metrics
}

func (device *DeviceViewModel) SetChannels(channels []model.Channel) {
	device.Channels = channels
}

func (device *DeviceViewModel) GetChannels() []model.Channel {
	return device.Channels
}

func (device *DeviceViewModel) GetTimestamp() time.Time {
	return device.Timestamp
}
