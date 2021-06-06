package model

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
)

type Controller struct {
	ID              int               `yaml:"id" json:"id"`
	OrgID           int               `yaml:"orgId" json:"orgId"`
	Type            string            `yaml:"type" json:"type"`
	Description     string            `yaml:"description" json:"description"`
	Enable          bool              `yaml:"enable" json:"enable"`
	Notify          bool              `yaml:"notify" json:"notify"`
	URI             string            `yaml:"uri" json:"uri"`
	HardwareVersion string            `yaml:"hardwareVersion" json:"hardwareVersion"`
	FirmwareVersion string            `yaml:"firmwareVersion" json:"firmwareVersion"`
	Configs         map[string]string `yaml:"configs" json:"configs"`
	Metrics         []common.Metric   `yaml:"metrics" json:"metrics"`
	Channels        []common.Channel  `yaml:"channels" json:"channels"`
	common.Controller
}

func (controller *Controller) GetID() int {
	return controller.ID
}

func (controller *Controller) SetID(id int) {
	controller.ID = id
}

func (controller *Controller) GetOrgID() int {
	return controller.OrgID
}

func (controller *Controller) SetOrgID(id int) {
	controller.OrgID = id
}

func (controller *Controller) GetType() string {
	return controller.Type
}

func (controller *Controller) SetType(controllerType string) {
	controller.Type = controllerType
}

func (controller *Controller) GetDescription() string {
	return controller.Description
}

func (controller *Controller) SetDescription(description string) {
	controller.Description = description
}

func (controller *Controller) IsEnabled() bool {
	return controller.Enable
}

func (controller *Controller) SetEnabled(enabled bool) {
	controller.Enable = enabled
}

func (controller *Controller) IsNotify() bool {
	return controller.Notify
}

func (controller *Controller) SetNotify(notify bool) {
	controller.Notify = notify
}

func (controller *Controller) GetURI() string {
	return controller.URI
}

func (controller *Controller) SetURI(uri string) {
	controller.URI = uri
}

func (controller *Controller) GetHardwareVersion() string {
	return controller.HardwareVersion
}

func (controller *Controller) SetHardwareVersion(version string) {
	controller.HardwareVersion = version
}

func (controller *Controller) GetFirmwareVersion() string {
	return controller.FirmwareVersion
}

func (controller *Controller) SetFirmwareVersion(version string) {
	controller.FirmwareVersion = version
}

func (controller *Controller) GetMetric(key string) (common.Metric, error) {
	for _, metric := range controller.Metrics {
		if metric.GetKey() == key {
			return metric, nil
		}
	}
	return nil, fmt.Errorf("Metric key not found: %s", key)
}

func (controller *Controller) GetMetrics() []common.Metric {
	return controller.Metrics
}

func (controller *Controller) SetMetrics(metrics []common.Metric) {
	controller.Metrics = metrics
}

func (controller *Controller) GetChannels() []common.Channel {
	return controller.Channels
}

func (controller *Controller) GetChannel(id int) (common.Channel, error) {
	if id < 0 || id > len(controller.Channels) {
		return nil, fmt.Errorf("Channel ID not found: %d", id)
	}
	return controller.Channels[id], nil
}

func (controller *Controller) SetChannels(channels []common.Channel) {
	controller.Channels = channels
}

func (controller *Controller) GetConfigs() map[string]string {
	return controller.Configs
}

func (controller *Controller) SetConfigs(configs map[string]string) {
	controller.Configs = configs
}
