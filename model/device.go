package model

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
)

type Device interface {
	GetMetric(key string) (Metric, error)
	GetMetrics() []Metric
	SetMetrics(metrics []Metric)
	GetChannels() []Channel
	GetChannel(id int) (Channel, error)
	SetChannels(channels []Channel)
	config.CommonDevice
}

type DeviceStruct struct {
	ID              uint64 `yaml:"id" json:"id"`
	OrgID           int    `yaml:"orgId" json:"orgId"`
	Type            string `yaml:"type" json:"type"`
	Description     string `yaml:"description" json:"description"`
	Enable          bool   `yaml:"enable" json:"enable"`
	Notify          bool   `yaml:"notify" json:"notify"`
	URI             string `yaml:"uri" json:"uri"`
	HardwareVersion string `yaml:"hwVersion" json:"hwVersion"`
	FirmwareVersion string `yaml:"fwVersion" json:"fwVersion"`
	//Configs         map[string]string `yaml:"configs" json:"configs"`
	Settings map[string]string `yaml:"configMap" json:"configMap"`
	Metrics  []Metric          `yaml:"metrics" json:"metrics"`
	Channels []Channel         `yaml:"channels" json:"channels"`
	Device   `yaml:"-" json:"-"`
}

func (device *DeviceStruct) Identifier() uint64 {
	return device.ID
}

func (device *DeviceStruct) SetID(id uint64) {
	device.ID = id
}

func (device *DeviceStruct) GetOrgID() int {
	return device.OrgID
}

func (device *DeviceStruct) SetOrgID(id int) {
	device.OrgID = id
}

func (device *DeviceStruct) GetType() string {
	return device.Type
}

func (device *DeviceStruct) SetType(deviceType string) {
	device.Type = deviceType
}

func (device *DeviceStruct) GetDescription() string {
	return device.Description
}

func (device *DeviceStruct) SetDescription(description string) {
	device.Description = description
}

func (device *DeviceStruct) IsEnabled() bool {
	return device.Enable
}

func (device *DeviceStruct) SetEnabled(enabled bool) {
	device.Enable = enabled
}

func (device *DeviceStruct) IsNotify() bool {
	return device.Notify
}

func (device *DeviceStruct) SetNotify(notify bool) {
	device.Notify = notify
}

func (device *DeviceStruct) GetURI() string {
	return device.URI
}

func (device *DeviceStruct) SetURI(uri string) {
	device.URI = uri
}

func (device *DeviceStruct) GetHardwareVersion() string {
	return device.HardwareVersion
}

func (device *DeviceStruct) SetHardwareVersion(version string) {
	device.HardwareVersion = version
}

func (device *DeviceStruct) GetFirmwareVersion() string {
	return device.FirmwareVersion
}

func (device *DeviceStruct) SetFirmwareVersion(version string) {
	device.FirmwareVersion = version
}

func (device *DeviceStruct) GetMetric(key string) (Metric, error) {
	for _, metric := range device.Metrics {
		if metric.GetKey() == key {
			return metric, nil
		}
	}
	return nil, fmt.Errorf("Metric key not found: %s", key)
}

func (device *DeviceStruct) GetMetrics() []Metric {
	return device.Metrics
}

func (device *DeviceStruct) SetMetrics(metrics []Metric) {
	device.Metrics = metrics
}

func (device *DeviceStruct) GetChannels() []Channel {
	return device.Channels
}

func (device *DeviceStruct) GetChannel(id int) (Channel, error) {
	if id < 0 || id > len(device.Channels) {
		return nil, fmt.Errorf("Channel ID not found: %d", id)
	}
	return device.Channels[id], nil
}

func (device *DeviceStruct) SetChannels(channels []Channel) {
	device.Channels = channels
}

func (device *DeviceStruct) GetSettings() map[string]string {
	return device.Settings
}

func (device *DeviceStruct) SetSettings(configs map[string]string) {
	device.Settings = configs
}

func (device *DeviceStruct) GetSettingsMap() map[string]string {
	return device.Settings
}
