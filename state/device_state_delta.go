package state

import (
	"time"
)

// DeviceStateDeltaMap stores only the current data points that have changed
// since the last time the state was collected from a device.
type DeviceStateDeltaMap interface {
	GetMetrics() map[string]float64
	SetMetrics(map[string]float64)
	GetChannels() map[int]int
	SetChannels(channels map[int]int)
	GetTimestamp() time.Time
	SetTimestamp(time.Time)
}

type DeviceStateDelta struct {
	Metrics             map[string]float64 `yaml:"metrics" json:"metrics"`
	Channels            map[int]int        `yaml:"channels" json:"channels"`
	Timestamp           time.Time          `yaml:"timestamp" json:"timestamp"`
	DeviceStateDeltaMap `yaml:"-" json:"-"`
}

func NewDeviceStateDeltaMap() DeviceStateDeltaMap {
	return &DeviceStateDelta{
		Metrics:   make(map[string]float64, 0),
		Channels:  make(map[int]int, 0),
		Timestamp: time.Now()}
}

func CreateDeviceStateDeltaMap(metrics map[string]float64, channels map[int]int) DeviceStateDeltaMap {
	return &DeviceStateDelta{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: time.Now()}
}

func (state *DeviceStateDelta) GetMetrics() map[string]float64 {
	return state.Metrics
}

func (state *DeviceStateDelta) SetMetrics(metrics map[string]float64) {
	state.Metrics = metrics
}

func (state *DeviceStateDelta) GetChannels() map[int]int {
	return state.Channels
}

func (state *DeviceStateDelta) SetChannels(channels map[int]int) {
	state.Channels = channels
}

func (state *DeviceStateDelta) GetTimestamp() time.Time {
	return state.Timestamp
}

func (state *DeviceStateDelta) SetTimestamp(timestamp time.Time) {
	state.Timestamp = timestamp
}
