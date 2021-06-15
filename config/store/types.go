package store

import "github.com/jeremyhahn/go-cropdroid/config"

const (
	MEMORY_STORE = iota
	GORM_STORE
	RAFT_STORE
)

type FarmConfigStorer interface {
	Cache(farmID uint64, farmConfig config.FarmConfig)
	Get(farmID uint64, CONSISTENCY_LEVEL int) (config.FarmConfig, error)
	GetAll() []config.FarmConfig
	Len() int
	Put(farmID uint64, farmConfig config.FarmConfig) error
}

type DeviceConfigStorer interface {
	Cache(deviceID uint64, farmConfig config.DeviceConfig)
	Get(deviceID uint64, CONSISTENCY_LEVEL int) (config.DeviceConfig, error)
	GetAll(farmID uint64) []config.DeviceConfig
	Len() int
	Put(deviceID uint64, farmConfig config.DeviceConfig) error
}
