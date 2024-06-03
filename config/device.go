package config

import (
	"fmt"
	"strconv"
)

type CommonDevice interface {
	GetFarmID() uint64
	SetFarmID(uint64)
	GetType() string
	SetType(string)
	GetInterval() int
	SetInterval(int)
	GetDescription() string
	SetDescription(string)
	GetHardwareVersion() string
	SetHardwareVersion(string)
	GetFirmwareVersion() string
	SetFirmwareVersion(string)
	// GetConfigs() map[string]string
	// SetConfigs(map[string]string)
	GetSettingsMap() map[string]string
	IsEnabled() bool
	SetEnabled(enabled bool)
	IsNotify() bool
	SetNotify(notify bool)
	GetURI() string
	SetURI(uri string)
	KeyValueEntity
	//HydrateConfigs() error
}

type Device interface {
	GetSettings() []*DeviceSettingStruct
	GetSetting(key string) *DeviceSettingStruct
	SetSettings(settings []*DeviceSettingStruct)
	SetSetting(deviceSetting *DeviceSettingStruct)
	GetMetric(key string) (*MetricStruct, error)
	GetMetrics() []*MetricStruct
	SetMetric(metric *MetricStruct)
	SetMetrics([]*MetricStruct)
	GetChannel(id int) (*ChannelStruct, error)
	GetChannels() []*ChannelStruct
	SetChannel(channel *ChannelStruct)
	SetChannels([]*ChannelStruct)
	CommonDevice
}

type DeviceWithSettings interface {
	ParseSettings() error
}

type DeviceStruct struct {
	ID                 uint64                 `gorm:"primaryKey" yaml:"id" json:"id"`
	FarmID             uint64                 `yaml:"farmId" json:"farmId"`
	Type               string                 `yaml:"type" json:"type"`
	Interval           int                    `yaml:"interval" json:"interval"`
	Description        string                 `yaml:"description" json:"description"`
	Enable             bool                   `gorm:"-" yaml:"enable" json:"enable"`
	Notify             bool                   `gorm:"-" yaml:"notify" json:"notify"`
	URI                string                 `gorm:"-" yaml:"uri" json:"uri"`
	HardwareVersion    string                 `gorm:"hw_version" yaml:"hwVersion" json:"hwVersion"`
	FirmwareVersion    string                 `gorm:"fw_version" yaml:"fwVersion" json:"fwVersion"`
	SettingsMap        map[string]string      `gorm:"-" yaml:"configMap" json:"configMap"`
	Settings           []*DeviceSettingStruct `gorm:"foreignKey:DeviceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"settings" json:"settings"`
	Metrics            []*MetricStruct        `gorm:"foreignKey:DeviceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"metrics" json:"metrics"`
	Channels           []*ChannelStruct       `gorm:"foreignKey:DeviceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"channels" json:"channels"`
	Device             `sql:"-" gorm:"-" yaml:"-" json:"-"`
	DeviceWithSettings `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewDevice() *DeviceStruct {
	return &DeviceStruct{
		SettingsMap: make(map[string]string, 0),
		Settings:    make([]*DeviceSettingStruct, 0),
		Metrics:     make([]*MetricStruct, 0),
		Channels:    make([]*ChannelStruct, 0)}
}

func (ddevice *DeviceStruct) TableName() string {
	return "devices"
}

func (device *DeviceStruct) Identifier() uint64 {
	return device.ID
}

func (device *DeviceStruct) SetID(id uint64) {
	device.ID = id
}

func (device *DeviceStruct) GetFarmID() uint64 {
	return device.FarmID
}

func (device *DeviceStruct) SetFarmID(id uint64) {
	device.FarmID = id
}

func (device *DeviceStruct) GetType() string {
	return device.Type
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

func (device *DeviceStruct) GetURI() string {
	return device.URI
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

func (device *DeviceStruct) GetInterval() int {
	return device.Interval
}

func (device *DeviceStruct) SetInterval(interval int) {
	device.Interval = interval
}

func (device *DeviceStruct) GetSettingsMap() map[string]string {
	return device.SettingsMap
}

func (device *DeviceStruct) GetSettings() []*DeviceSettingStruct {
	return device.Settings
}

// Only used by test/datastore to allow looking up configs
// in an order agnostic manner.
func (device *DeviceStruct) GetSetting(key string) *DeviceSettingStruct {
	for _, setting := range device.Settings {
		if setting.GetKey() == key {
			return setting
		}
	}
	return nil
}

func (device *DeviceStruct) SetSettings(settings []*DeviceSettingStruct) {
	device.Settings = settings
}

func (device *DeviceStruct) SetSetting(deviceSetting *DeviceSettingStruct) {
	id := deviceSetting.ID
	value := deviceSetting.GetValue()
	for i, configItem := range device.Settings {
		if configItem.ID == id {
			device.Settings[i].Value = value
			device.SettingsMap[configItem.GetKey()] = value
			return
		}
	}
	device.Settings = append(device.Settings, deviceSetting)
	device.SettingsMap[deviceSetting.GetKey()] = deviceSetting.GetValue()
}

func (device *DeviceStruct) GetMetric(key string) (*MetricStruct, error) {
	for _, metric := range device.Metrics {
		if metric.GetKey() == key {
			return metric, nil
		}
	}
	return nil, fmt.Errorf("Metric key not found: %s", key)
}

func (device *DeviceStruct) GetMetrics() []*MetricStruct {
	return device.Metrics
}

func (device *DeviceStruct) SetMetric(metric *MetricStruct) {
	for i, m := range device.Metrics {
		if m.ID == metric.ID {
			device.Metrics[i] = metric
			return
		}
	}
	device.Metrics = append(device.Metrics, metric)
}

func (device *DeviceStruct) SetMetrics(metrics []*MetricStruct) {
	device.Metrics = metrics
}

func (device *DeviceStruct) GetChannel(id int) (*ChannelStruct, error) {
	if id < 0 || id > len(device.Channels) {
		return nil, fmt.Errorf("Channel ID not found: %d", id)
	}
	return device.Channels[id], nil
}

func (device *DeviceStruct) GetChannels() []*ChannelStruct {
	return device.Channels
}

func (device *DeviceStruct) SetChannel(channel *ChannelStruct) {
	for i, c := range device.Channels {
		if c.ID == channel.ID {
			device.Channels[i] = channel
			return
		}
	}
	device.Channels = append(device.Channels, channel)
}

func (device *DeviceStruct) SetChannels(channels []*ChannelStruct) {
	device.Channels = channels
}

func (device *DeviceStruct) ParseSettings() error {
	configs := device.GetSettings()
	if device.SettingsMap == nil {
		device.SettingsMap = make(map[string]string, len(configs))
	}
	enableKey := fmt.Sprintf("%s.enable", device.Type)
	notifyKey := fmt.Sprintf("%s.notify", device.Type)
	uriKey := fmt.Sprintf("%s.uri", device.Type)
	for _, config := range configs {
		key := config.GetKey()
		value := config.GetValue()
		device.SettingsMap[key] = value
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

// // HydrateConfigs populates the device config items from the SettingsMap. This is
// // used when unmarshalling from JSON or YAML since device.Configs json:"-" and yaml:"-"
// // is set so the results are returned as key/value pairs by the API. Probably best to refactor
// // this so the API returns a dedicated view and device.Configs doesn't get ignored.
// func (device *DeviceStruct) HydrateConfigs() error {
// 	configs := device.GetSettingsMap()
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
func (device *DeviceStruct) getBoolConfig(key string) (bool, error) {
	if device.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", device.Type, key)
		if boolean, ok := device.SettingsMap[configKey]; ok {
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

func (device *DeviceStruct) getStringConfig(key string) (string, error) {
	if device.Configs != nil {
		configKey := fmt.Sprintf("%s.%s", device.Type, key)
		if value, ok := device.SettingsMap[configKey]; ok {
			return value, nil
		}
		return "", fmt.Errorf("Device config (string) not found: %s", configKey)
	}
	return "", nil
}
*/

/*
func (device *DeviceStruct) getBoolConfig(key string) (bool, error) {
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

func (device *DeviceStruct) getStringConfig(key string) (string, error) {
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
func (device *DeviceStruct) Get() (bool, error) {
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
