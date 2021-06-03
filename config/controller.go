package config

import (
	"fmt"
	"strconv"
)

type Controller struct {
	ID               int                    `gorm:"primary_key;auto_increment" yaml:"id" json:"id"`
	FarmID           int                    `yaml:"farmId" json:"farmId"`
	Type             string                 `yaml:"type" json:"type"`
	Interval         int                    `yaml:"interval" json:"interval"`
	Description      string                 `yaml:"description" json:"description"`
	Enable           bool                   `gorm:"-" yaml:"-" json:"enable"`
	Notify           bool                   `gorm:"-" yaml:"-" json:"notify"`
	URI              string                 `gorm:"-" yaml:"-" json:"uri"`
	HardwareVersion  string                 `gorm:"hw_version" yaml:"hwVersion" json:"hwVersion"`
	FirmwareVersion  string                 `gorm:"fw_version" yaml:"fwVersion" json:"fwVersion"`
	ConfigMap        map[string]string      `gorm:"-" yaml:"configs" json:"configs"`
	Configs          []ControllerConfigItem `yaml:"-" json:"-"`
	Metrics          []Metric               `yaml:"metrics" json:"metrics"`
	Channels         []Channel              `yaml:"channels" json:"channels"`
	ControllerConfig `yaml:"-" json:"-"`
}

func NewController() *Controller {
	return &Controller{
		ConfigMap: make(map[string]string, 0),
		Configs:   make([]ControllerConfigItem, 0),
		Metrics:   make([]Metric, 0),
		Channels:  make([]Channel, 0)}
}

func (controller *Controller) GetID() int {
	return controller.ID
}

func (controller *Controller) SetID(id int) {
	controller.ID = id
}

func (controller *Controller) GetFarmID() int {
	return controller.FarmID
}

func (controller *Controller) SetFarmID(id int) {
	controller.FarmID = id
}

func (controller *Controller) GetType() string {
	return controller.Type
}

func (controller *Controller) IsEnabled() bool {
	return controller.Enable
}

func (controller *Controller) IsNotify() bool {
	return controller.Notify
}

func (controller *Controller) GetURI() string {
	return controller.URI
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

func (controller *Controller) GetInterval() int {
	return controller.Interval
}

func (controller *Controller) SetInterval(interval int) {
	controller.Interval = interval
}

func (controller *Controller) GetConfigMap() map[string]string {
	return controller.ConfigMap
}

func (controller *Controller) GetConfigs() []ControllerConfigItem {
	return controller.Configs
}

func (controller *Controller) SetConfigs(configs []ControllerConfigItem) {
	controller.Configs = configs
}

func (controller *Controller) SetConfig(controllerConfig ControllerConfigConfig) {
	id := controllerConfig.GetID()
	value := controllerConfig.GetValue()
	for i, configItem := range controller.Configs {
		if configItem.GetID() == id {
			controller.Configs[i].Value = value
			controller.ConfigMap[configItem.GetKey()] = value
			return
		}
	}
	controller.Configs = append(controller.Configs, *controllerConfig.(*ControllerConfigItem))
	controller.ConfigMap[controllerConfig.GetKey()] = controllerConfig.GetValue()
}

func (controller *Controller) GetMetric(key string) (*Metric, error) {
	for _, metric := range controller.Metrics {
		if metric.GetKey() == key {
			return &metric, nil
		}
	}
	return nil, fmt.Errorf("Metric key not found: %s", key)
}

func (controller *Controller) GetMetrics() []Metric {
	return controller.Metrics
}

func (controller *Controller) SetMetric(metric MetricConfig) {
	for i, m := range controller.Metrics {
		if m.GetID() == metric.GetID() {
			controller.Metrics[i] = *metric.(*Metric)
			return
		}
	}
	controller.Metrics = append(controller.Metrics, *metric.(*Metric))
}

func (controller *Controller) SetMetrics(metrics []Metric) {
	controller.Metrics = metrics
}

func (controller *Controller) GetChannel(id int) (*Channel, error) {
	if id < 0 || id > len(controller.Channels) {
		return nil, fmt.Errorf("Channel ID not found: %d", id)
	}
	return &controller.Channels[id], nil
}

func (controller *Controller) GetChannels() []Channel {
	return controller.Channels
}

func (controller *Controller) SetChannel(channel ChannelConfig) {
	for i, c := range controller.Channels {
		if c.GetID() == channel.GetID() {
			controller.Channels[i] = *channel.(*Channel)
			return
		}
	}
	controller.Channels = append(controller.Channels, *channel.(*Channel))
}

func (controller *Controller) SetChannels(channels []Channel) {
	controller.Channels = channels
}

func (controller *Controller) ParseConfigs() error {
	configs := controller.GetConfigs()
	if controller.ConfigMap == nil {
		controller.ConfigMap = make(map[string]string, len(configs))
	}
	enableKey := fmt.Sprintf("%s.enable", controller.Type)
	notifyKey := fmt.Sprintf("%s.notify", controller.Type)
	uriKey := fmt.Sprintf("%s.uri", controller.Type)
	for _, config := range configs {
		key := config.GetKey()
		value := config.GetValue()
		controller.ConfigMap[key] = value
		switch key {
		case enableKey:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			controller.Enable = b
		case notifyKey:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			controller.Notify = b
		case uriKey:
			controller.URI = value
		}
	}
	return nil
}

// HydrateConfigs populates the controller config items from the ConfigMap. This is
// used when unmarshalling from JSON or YAML since controller.Configs json:"-" and yaml:"-"
// is set so the results are returned as key/value pairs by the API. Probably best to refactor
// this so the API returns a dedicated view and controller.Configs doesn't get ignored.
func (controller *Controller) HydrateConfigs() error {
	configs := controller.GetConfigMap()
	enableKey := fmt.Sprintf("%s.enable", controller.Type)
	notifyKey := fmt.Sprintf("%s.notify", controller.Type)
	uriKey := fmt.Sprintf("%s.uri", controller.Type)
	for key, value := range configs {
		switch key {
		case enableKey:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			controller.Enable = b
		case notifyKey:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			controller.Notify = b
		case uriKey:
			controller.URI = value
		}
	}
	return nil
}

/*
func (controller *Controller) getBoolConfig(key string) (bool, error) {
	if controller.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", controller.Type, key)
		if boolean, ok := controller.ConfigMap[configKey]; ok {
			b, err := strconv.ParseBool(boolean)
			if err != nil {
				return false, err
			}
			return b, nil
		}
		return false, fmt.Errorf("Controller config (bool) not found: %s", configKey)
	}
	return false, nil
}

func (controller *Controller) getStringConfig(key string) (string, error) {
	if controller.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", controller.Type, key)
		if value, ok := controller.ConfigMap[configKey]; ok {
			return value, nil
		}
		return "", fmt.Errorf("Controller config (string) not found: %s", configKey)
	}
	return "", nil
}
*/

/*
func (controller *Controller) getBoolConfig(key string) (bool, error) {
	if controller.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", controller.Type, key)
		for _, config := range controller.Configs {
			if config.GetKey() == configKey {
				b, err := strconv.ParseBool(config.GetValue())
				if err != nil {
					return false, err
				}
				return b, nil
			}
		}
		return false, nil
	}
	return false, fmt.Errorf("Controller config (bool) not found: %s", key)
}

func (controller *Controller) getStringConfig(key string) (string, error) {
	if controller.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", controller.Type, key)
		for _, config := range controller.Configs {
			if config.GetKey() == configKey {
				return config.GetValue(), nil
			}
		}
		return "", fmt.Errorf("Controller config (string) not found: %s", key)
	}
	return "", fmt.Errorf("Controller config (string) not found: %s", key)
}
*/

/*
func (controller *Controller) Get() (bool, error) {
	if controller.Configs != nil {
		key := fmt.Sprintf("%s.%s", controller.Type, "enabled")
		if enabled, ok := controller.Configs[key]; ok {
			b, err := strconv.ParseBool(enabled)
			if err != nil {
				return false, err
			}
			return b, nil
		} else {
			return false, fmt.Errorf("Controller missing 'enabled' config key")
		}
	}
	return false, nil
}*/
