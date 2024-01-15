//go:build ignore
// +build ignore

package store

import (
	"sync"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/config/store"
)

type GormDeviceConfigStore struct {
	farmID       uint64
	deviceDAO    dao.DeviceDAO
	cachedConfig map[uint64]config.Device
	mutex        *sync.RWMutex
	store.DeviceConfigStorer
}

func NewGormDeviceConfigStore(deviceDAO dao.DeviceDAO, len int) store.DeviceConfigStorer {

	return &GormDeviceConfigStore{
		deviceDAO:    deviceDAO,
		mutex:        &sync.RWMutex{},
		cachedConfig: make(map[uint64]config.Device, 0)}
}

func (s *GormDeviceConfigStore) Len() int {
	count, _ := s.deviceDAO.Count()
	return int(count)
}

func (s *GormDeviceConfigStore) Cache(deviceID uint64, c config.DeviceConfig) {
	s.mutex.RLock()
	s.cachedConfig[deviceID] = *c.(*config.Device)
	s.mutex.RUnlock()
}

func (s *GormDeviceConfigStore) Put(deviceID uint64, c config.DeviceConfig) error {
	deviceConfig := c.(*config.Device)
	//s.Cache(deviceID, deviceConfig)
	return s.deviceDAO.Save(deviceConfig)
}

func (s *GormDeviceConfigStore) Get(deviceID uint64, CONSISTENCY_LEVEL int) (config.DeviceConfig, error) {
	// if CONSISTENCY_LEVEL == common.CONSISTENCY_CACHED {
	// 	if config, ok := s.cachedConfig[deviceID]; ok {
	// 		return &config, nil
	// 	}
	// }
	config, err := s.deviceDAO.Get(deviceID)
	//s.Cache(deviceID, config)
	return config, err
}

func (s *GormDeviceConfigStore) GetAll(farmID uint64) []config.DeviceConfig {
	configs, _ := s.deviceDAO.GetByFarmId(s.farmID)
	deviceConfigs := make([]config.DeviceConfig, len(configs))
	for i, conf := range configs {
		deviceConfigs[i] = &conf
	}
	return deviceConfigs
}
