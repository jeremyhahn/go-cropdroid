package model

import (
	"github.com/jeremyhahn/go-cropdroid/common"
)

type Farm struct {
	ID       int    `yaml:"id" json:"id"`
	OrgID    int    `yaml:"orgId" json:"orgId"`
	Mode     string `yaml:"mode" json:"mode"`
	Name     string `yaml:"name" json:"name"`
	Interval int    `yaml:"interval" json:"interval"`
	//Consistency int                 `yaml:"consistency" json:"consistency"`
	Devices []common.Device `yaml:"devices" json:"devices"`
	common.Farm `yaml:"-" json:"-"`
}

func NewFarm() common.Farm {
	return &Farm{Devices: make([]common.Device, 0)}
}

// func CreateFarm(name string, orgID, interval, consistency int,
// 	devices []common.Device) common.Farm {

// 	return &Farm{
// 		Name:        name,
// 		OrgID:       orgID,
// 		Interval:    interval,
// 		Consistency: consistency,
// 		Devices: devices}
// }

func CreateFarm(name string, orgID, interval int, devices []common.Device) common.Farm {
	return &Farm{
		Name:        name,
		OrgID:       orgID,
		Interval:    interval,
		Devices: devices}
}

func (farm *Farm) SetID(id int) {
	farm.ID = id
}

func (farm *Farm) GetID() int {
	return farm.ID
}

func (farm *Farm) SetOrgID(id int) {
	farm.OrgID = id
}

func (farm *Farm) GetOrgID() int {
	return farm.OrgID
}

func (farm *Farm) SetMode(mode string) {
	farm.Mode = mode
}

func (farm *Farm) GetMode() string {
	return farm.Mode
}

// func (farm *Farm) SetConsistency(level int) {
// 	farm.Consistency = level
// }

// func (farm *Farm) GetConsistency() int {
// 	return farm.Consistency
// }

func (farm *Farm) SetName(name string) {
	farm.Name = name
}

func (farm *Farm) GetName() string {
	return farm.Name
}

func (farm *Farm) SetInterval(interval int) {
	farm.Interval = interval
}

func (farm *Farm) GetInterval() int {
	return farm.Interval
}

func (farm *Farm) GetDevices() []common.Device {
	return farm.Devices
}

func (farm *Farm) SetDevices(devices []common.Device) {
	farm.Devices = devices
}

/*
func (farm *Farm) AddDevice(device common.Device) {
	farm.Devices = append(farm.Devices, device)
}

func (farm *Farm) GetDevice(deviceType string) (common.Device, error) {
	for _, device := range farm.Devices {
		if device.GetType() == deviceType {
			return device, nil
		}
	}
	return nil, fmt.Errorf("Device type not found: %s", deviceType)
}

func (farm *Farm) GetDeviceById(id int) (common.Device, error) {
	farmSize := len(farm.Devices)
	if farmSize < id {
		return nil, fmt.Errorf("Device ID out of bounds: %d. Farm size: %d", id, farmSize)
	}
	return farm.Devices[id], nil
}
*/
