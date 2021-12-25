package store

import "github.com/jeremyhahn/go-cropdroid/config"

type FarmConfigStorer interface {
	Cache(farmID uint64, farmConfig config.FarmConfig)
	Get(farmID uint64, CONSISTENCY_LEVEL int) (config.FarmConfig, error)
	GetAll() []config.FarmConfig
	GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) []config.FarmConfig
	Len() int
	Put(farmID uint64, farmConfig config.FarmConfig) error
}

type DeviceConfigStorer interface {
	Cache(deviceID uint64, farmConfig config.DeviceConfig)
	Get(deviceID uint64, CONSISTENCY_LEVEL int) (config.DeviceConfig, error)
	GetAll(deviceID uint64) []config.DeviceConfig
	Len() int
	Put(deviceID uint64, deviceConfig config.DeviceConfig) error
}
