package state

import (
	"time"
)

// DeviceStateMap stores the current data points collected from a device
type DeviceStateMap interface {
	/*
		GetHardwareVersion() string
		SetHardwareVersion(string)
		GetFirmwareVersion() string
		SetFirmwareVersion(string)
	*/
	GetID() uint64
	SetID(uint64)
	GetMetrics() map[string]float64
	SetMetrics(map[string]float64)
	GetChannels() []int
	SetChannels(channels []int)
	GetTimestamp() time.Time
	SetTimestamp(time.Time)
	Clone() DeviceStateMap
}

type DeviceState struct {
	//HardwareVersion    string             `yaml:"hardwareVersion" json:"hardwareVersion"`
	//FirmwareVersion    string             `yaml:"firmwareVersion" json:"firmwareVersion"`
	DeviceID       uint64             `yaml:"device_id" json:"device_id"`
	Metrics        map[string]float64 `yaml:"metrics" json:"metrics"`
	Channels       []int              `yaml:"channels" json:"channels"`
	Timestamp      time.Time          `yaml:"timestamp" json:"timestamp"`
	DeviceStateMap `yaml:"-" json:"-"`
}

func NewDeviceStateMap() DeviceStateMap {
	return &DeviceState{
		Metrics:  make(map[string]float64, 0),
		Channels: make([]int, 0)}
}

func CreateDeviceStateMap(metrics map[string]float64, channels []int) DeviceStateMap {
	return &DeviceState{
		Metrics:  metrics,
		Channels: channels}
}

func CreateEmptyDeviceStateMap(deviceID uint64, numMetrics, numChannels int) DeviceStateMap {
	return &DeviceState{
		DeviceID: deviceID,
		Metrics:  make(map[string]float64, numMetrics),
		Channels: make([]int, numChannels)}
}

/*
func (state *DeviceState) GetHardwareVersion() string {
	return state.HardwareVersion
}

func (state *DeviceState) SetHardwareVersion(version string) {
	state.HardwareVersion = version
}

func (state *DeviceState) GetFirmwareVersion() string {
	return state.FirmwareVersion
}

func (state *DeviceState) SetFirmwareVersion(version string) {
	state.FirmwareVersion = version
}
*/
func (state *DeviceState) GetID() uint64 {
	return state.DeviceID
}

func (state *DeviceState) SetID(id uint64) {
	state.DeviceID = id
}

func (state *DeviceState) GetMetrics() map[string]float64 {
	return state.Metrics
}

func (state *DeviceState) SetMetrics(metrics map[string]float64) {
	state.Metrics = metrics
}

func (state *DeviceState) GetChannels() []int {
	return state.Channels
}

func (state *DeviceState) SetChannels(channels []int) {
	state.Channels = channels
}

func (state *DeviceState) GetTimestamp() time.Time {
	return state.Timestamp
}

func (state *DeviceState) SetTimestamp(timestamp time.Time) {
	state.Timestamp = timestamp
}

func (state *DeviceState) Clone() DeviceStateMap {
	metrics := make(map[string]float64, len(state.Metrics))
	channels := make([]int, len(state.Channels))
	for i, metric := range state.Metrics {
		metrics[i] = metric
	}
	for i, channel := range state.Channels {
		channels[i] = channel
	}
	return &DeviceState{
		DeviceID:  state.DeviceID,
		Metrics:   metrics,
		Channels:  channels,
		Timestamp: state.Timestamp}
}
