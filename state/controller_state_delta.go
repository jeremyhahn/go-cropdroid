package state

import (
	"time"
)

// ControllerStateDeltaMap stores only the current data points that have changed
// since the last time the state was collected from a controller.
type ControllerStateDeltaMap interface {
	GetMetrics() map[string]float64
	SetMetrics(map[string]float64)
	GetChannels() map[int]int
	SetChannels(channels map[int]int)
	GetTimestamp() time.Time
	SetTimestamp(time.Time)
}

type ControllerStateDelta struct {
	Metrics                 map[string]float64 `yaml:"metrics" json:"metrics"`
	Channels                map[int]int        `yaml:"channels" json:"channels"`
	Timestamp               time.Time          `yaml:"timestamp" json:"timestamp"`
	ControllerStateDeltaMap `yaml:"-" json:"-"`
}

func NewControllerStateDeltaMap() ControllerStateDeltaMap {
	return &ControllerStateDelta{
		Metrics:   make(map[string]float64, 0),
		Channels:  make(map[int]int, 0),
		Timestamp: time.Now()}
}

func CreateControllerStateDeltaMap(metrics map[string]float64, channels map[int]int) ControllerStateDeltaMap {
	return &ControllerStateDelta{
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: time.Now()}
}

func (state *ControllerStateDelta) GetMetrics() map[string]float64 {
	return state.Metrics
}

func (state *ControllerStateDelta) SetMetrics(metrics map[string]float64) {
	state.Metrics = metrics
}

func (state *ControllerStateDelta) GetChannels() map[int]int {
	return state.Channels
}

func (state *ControllerStateDelta) SetChannels(channels map[int]int) {
	state.Channels = channels
}

func (state *ControllerStateDelta) GetTimestamp() time.Time {
	return state.Timestamp
}

func (state *ControllerStateDelta) SetTimestamp(timestamp time.Time) {
	state.Timestamp = timestamp
}
