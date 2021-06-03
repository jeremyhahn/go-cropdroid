package state

import (
	"time"
)

// ControllerStateMap stores the current data points collected from a controller
type ControllerStateMap interface {
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
}

type ControllerState struct {
	//HardwareVersion    string             `yaml:"hardwareVersion" json:"hardwareVersion"`
	//FirmwareVersion    string             `yaml:"firmwareVersion" json:"firmwareVersion"`
	ControllerID       uint64             `yaml:"controller_id" json:"controller_id"`
	Metrics            map[string]float64 `yaml:"metrics" json:"metrics"`
	Channels           []int              `yaml:"channels" json:"channels"`
	Timestamp          time.Time          `yaml:"timestamp" json:"timestamp"`
	ControllerStateMap `yaml:"-" json:"-"`
}

func NewControllerStateMap() ControllerStateMap {
	return &ControllerState{
		Metrics:  make(map[string]float64, 0),
		Channels: make([]int, 0)}
}

func CreateControllerStateMap(metrics map[string]float64, channels []int) ControllerStateMap {
	return &ControllerState{
		Metrics:  metrics,
		Channels: channels}
}

/*
func (state *ControllerState) GetHardwareVersion() string {
	return state.HardwareVersion
}

func (state *ControllerState) SetHardwareVersion(version string) {
	state.HardwareVersion = version
}

func (state *ControllerState) GetFirmwareVersion() string {
	return state.FirmwareVersion
}

func (state *ControllerState) SetFirmwareVersion(version string) {
	state.FirmwareVersion = version
}
*/
func (state *ControllerState) GetID() uint64 {
	return state.ControllerID
}

func (state *ControllerState) SetID(id uint64) {
	state.ControllerID = id
}

func (state *ControllerState) GetMetrics() map[string]float64 {
	return state.Metrics
}

func (state *ControllerState) SetMetrics(metrics map[string]float64) {
	state.Metrics = metrics
}

func (state *ControllerState) GetChannels() []int {
	return state.Channels
}

func (state *ControllerState) SetChannels(channels []int) {
	state.Channels = channels
}

func (state *ControllerState) GetTimestamp() time.Time {
	return state.Timestamp
}

func (state *ControllerState) SetTimestamp(timestamp time.Time) {
	state.Timestamp = timestamp
}
