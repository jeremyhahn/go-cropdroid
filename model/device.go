package model

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
)

type Device struct {
	ID              uint64            `yaml:"id" json:"id"`
	OrgID           int               `yaml:"orgId" json:"orgId"`
	Type            string            `yaml:"type" json:"type"`
	Description     string            `yaml:"description" json:"description"`
	Enable          bool              `yaml:"enable" json:"enable"`
	Notify          bool              `yaml:"notify" json:"notify"`
	URI             string            `yaml:"uri" json:"uri"`
	HardwareVersion string            `yaml:"hwVersion" json:"hwVersion"`
	FirmwareVersion string            `yaml:"fwVersion" json:"fwVersion"`
	Configs         map[string]string `yaml:"configs" json:"configs"`
	Metrics         []common.Metric   `yaml:"metrics" json:"metrics"`
	Channels        []common.Channel  `yaml:"channels" json:"channels"`
	common.Device   `yaml:"-" json:"-"`
}

func (device *Device) GetID() uint64 {
	return device.ID
}

func (device *Device) SetID(id uint64) {
	device.ID = id
}

func (device *Device) GetOrgID() int {
	return device.OrgID
}

func (device *Device) SetOrgID(id int) {
	device.OrgID = id
}

func (device *Device) GetType() string {
	return device.Type
}

func (device *Device) SetType(deviceType string) {
	device.Type = deviceType
}

func (device *Device) GetDescription() string {
	return device.Description
}

func (device *Device) SetDescription(description string) {
	device.Description = description
}

func (device *Device) IsEnabled() bool {
	return device.Enable
}

func (device *Device) SetEnabled(enabled bool) {
	device.Enable = enabled
}

func (device *Device) IsNotify() bool {
	return device.Notify
}

func (device *Device) SetNotify(notify bool) {
	device.Notify = notify
}

func (device *Device) GetURI() string {
	return device.URI
}

func (device *Device) SetURI(uri string) {
	device.URI = uri
}

func (device *Device) GetHardwareVersion() string {
	return device.HardwareVersion
}

func (device *Device) SetHardwareVersion(version string) {
	device.HardwareVersion = version
}

func (device *Device) GetFirmwareVersion() string {
	return device.FirmwareVersion
}

func (device *Device) SetFirmwareVersion(version string) {
	device.FirmwareVersion = version
}

func (device *Device) GetMetric(key string) (common.Metric, error) {
	for _, metric := range device.Metrics {
		if metric.GetKey() == key {
			return metric, nil
		}
	}
	return nil, fmt.Errorf("Metric key not found: %s", key)
}

func (device *Device) GetMetrics() []common.Metric {
	return device.Metrics
}

func (device *Device) SetMetrics(metrics []common.Metric) {
	device.Metrics = metrics
}

func (device *Device) GetChannels() []common.Channel {
	return device.Channels
}

func (device *Device) GetChannel(id int) (common.Channel, error) {
	if id < 0 || id > len(device.Channels) {
		return nil, fmt.Errorf("Channel ID not found: %d", id)
	}
	return device.Channels[id], nil
}

func (device *Device) SetChannels(channels []common.Channel) {
	device.Channels = channels
}

func (device *Device) GetConfigs() map[string]string {
	return device.Configs
}

func (device *Device) SetConfigs(configs map[string]string) {
	device.Configs = configs
}
