package config

import (
	"fmt"
	"strconv"
)

type Device struct {
	ID              uint64            `gorm:"primaryKey" yaml:"id" json:"id"`
	FarmID          uint64            `yaml:"farmId" json:"farmId"`
	Type            string            `yaml:"type" json:"type"`
	Interval        int               `yaml:"interval" json:"interval"`
	Description     string            `yaml:"description" json:"description"`
	Enable          bool              `gorm:"-" yaml:"enable" json:"enable"`
	Notify          bool              `gorm:"-" yaml:"notify" json:"notify"`
	URI             string            `gorm:"-" yaml:"uri" json:"uri"`
	HardwareVersion string            `gorm:"hw_version" yaml:"hwVersion" json:"hwVersion"`
	FirmwareVersion string            `gorm:"fw_version" yaml:"fwVersion" json:"fwVersion"`
	ConfigMap       map[string]string `gorm:"-" yaml:"configMap" json:"configMap"`
	//Configs         []DeviceConfigItem `yaml:"-" json:"-"`
	Settings       []*DeviceSetting `yaml:"settings" json:"settings"`
	Metrics        []*Metric        `yaml:"metrics" json:"metrics"`
	Channels       []*Channel       `yaml:"channels" json:"channels"`
	KeyValueEntity `gorm:"-" yaml:"-" json:"-"`
}

func NewDevice() *Device {
	return &Device{
		ConfigMap: make(map[string]string, 0),
		Settings:  make([]*DeviceSetting, 0),
		Metrics:   make([]*Metric, 0),
		Channels:  make([]*Channel, 0)}
}

func (device *Device) Identifier() uint64 {
	return device.ID
}

func (device *Device) SetID(id uint64) {
	device.ID = id
}

func (device *Device) GetFarmID() uint64 {
	return device.FarmID
}

func (device *Device) SetFarmID(id uint64) {
	device.FarmID = id
}

func (device *Device) GetType() string {
	return device.Type
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

func (device *Device) GetURI() string {
	return device.URI
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

func (device *Device) GetInterval() int {
	return device.Interval
}

func (device *Device) SetInterval(interval int) {
	device.Interval = interval
}

func (device *Device) GetConfigMap() map[string]string {
	return device.ConfigMap
}

func (device *Device) GetSettings() []*DeviceSetting {
	return device.Settings
}

// Only used by test/datastore to allow looking up configs
// in an order agnostic manner.
func (device *Device) GetSetting(key string) *DeviceSetting {
	for _, setting := range device.Settings {
		if setting.GetKey() == key {
			return setting
		}
	}
	return nil
}

func (device *Device) SetSettings(settings []*DeviceSetting) {
	device.Settings = settings
}

func (device *Device) SetSetting(deviceSetting *DeviceSetting) {
	id := deviceSetting.ID
	value := deviceSetting.GetValue()
	for i, configItem := range device.Settings {
		if configItem.ID == id {
			device.Settings[i].Value = value
			device.ConfigMap[configItem.GetKey()] = value
			return
		}
	}
	device.Settings = append(device.Settings, deviceSetting)
	device.ConfigMap[deviceSetting.GetKey()] = deviceSetting.GetValue()
}

func (device *Device) GetMetric(key string) (*Metric, error) {
	for _, metric := range device.Metrics {
		if metric.GetKey() == key {
			return metric, nil
		}
	}
	return nil, fmt.Errorf("Metric key not found: %s", key)
}

func (device *Device) GetMetrics() []*Metric {
	return device.Metrics
}

func (device *Device) SetMetric(metric *Metric) {
	for i, m := range device.Metrics {
		if m.ID == metric.ID {
			device.Metrics[i] = metric
			return
		}
	}
	device.Metrics = append(device.Metrics, metric)
}

func (device *Device) SetMetrics(metrics []*Metric) {
	device.Metrics = metrics
}

func (device *Device) GetChannel(id int) (*Channel, error) {
	if id < 0 || id > len(device.Channels) {
		return nil, fmt.Errorf("Channel ID not found: %d", id)
	}
	return device.Channels[id], nil
}

func (device *Device) GetChannels() []*Channel {
	return device.Channels
}

func (device *Device) SetChannel(channel *Channel) {
	for i, c := range device.Channels {
		if c.ID == channel.ID {
			device.Channels[i] = channel
			return
		}
	}
	device.Channels = append(device.Channels, channel)
}

func (device *Device) SetChannels(channels []*Channel) {
	device.Channels = channels
}

func (device *Device) ParseSettings() error {
	configs := device.GetSettings()
	if device.ConfigMap == nil {
		device.ConfigMap = make(map[string]string, len(configs))
	}
	enableKey := fmt.Sprintf("%s.enable", device.Type)
	notifyKey := fmt.Sprintf("%s.notify", device.Type)
	uriKey := fmt.Sprintf("%s.uri", device.Type)
	for _, config := range configs {
		key := config.GetKey()
		value := config.GetValue()
		device.ConfigMap[key] = value
		switch key {
		case enableKey:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			device.Enable = b
		case notifyKey:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			device.Notify = b
		case uriKey:
			device.URI = value
		}
	}
	return nil
}

// // HydrateConfigs populates the device config items from the ConfigMap. This is
// // used when unmarshalling from JSON or YAML since device.Configs json:"-" and yaml:"-"
// // is set so the results are returned as key/value pairs by the API. Probably best to refactor
// // this so the API returns a dedicated view and device.Configs doesn't get ignored.
// func (device *Device) HydrateConfigs() error {
// 	configs := device.GetConfigMap()
// 	enableKey := fmt.Sprintf("%s.enable", device.Type)
// 	notifyKey := fmt.Sprintf("%s.notify", device.Type)
// 	uriKey := fmt.Sprintf("%s.uri", device.Type)
// 	for key, value := range configs {
// 		switch key {
// 		case enableKey:
// 			b, err := strconv.ParseBool(value)
// 			if err != nil {
// 				return err
// 			}
// 			device.Enable = b
// 		case notifyKey:
// 			b, err := strconv.ParseBool(value)
// 			if err != nil {
// 				return err
// 			}
// 			device.Notify = b
// 		case uriKey:
// 			device.URI = value
// 		}
// 	}
// 	return nil
// }

/*
func (device *Device) getBoolConfig(key string) (bool, error) {
	if device.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", device.Type, key)
		if boolean, ok := device.ConfigMap[configKey]; ok {
			b, err := strconv.ParseBool(boolean)
			if err != nil {
				return false, err
			}
			return b, nil
		}
		return false, fmt.Errorf("Device config (bool) not found: %s", configKey)
	}
	return false, nil
}

func (device *Device) getStringConfig(key string) (string, error) {
	if device.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", device.Type, key)
		if value, ok := device.ConfigMap[configKey]; ok {
			return value, nil
		}
		return "", fmt.Errorf("Device config (string) not found: %s", configKey)
	}
	return "", nil
}
*/

/*
func (device *Device) getBoolConfig(key string) (bool, error) {
	if device.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", device.Type, key)
		for _, config := range device.Configs {
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
	return false, fmt.Errorf("Device config (bool) not found: %s", key)
}

func (device *Device) getStringConfig(key string) (string, error) {
	if device.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", device.Type, key)
		for _, config := range device.Configs {
			if config.GetKey() == configKey {
				return config.GetValue(), nil
			}
		}
		return "", fmt.Errorf("Device config (string) not found: %s", key)
	}
	return "", fmt.Errorf("Device config (string) not found: %s", key)
}
*/

/*
func (device *Device) Get() (bool, error) {
	if device.Configs != nil {
		key := fmt.Sprintf("%s.%s", device.Type, "enabled")
		if enabled, ok := device.Configs[key]; ok {
			b, err := strconv.ParseBool(enabled)
			if err != nil {
				return false, err
			}
			return b, nil
		} else {
			return false, fmt.Errorf("Device missing 'enabled' config key")
		}
	}
	return false, nil
}*/
