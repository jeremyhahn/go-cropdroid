// +build ignore

// This is not used, provided as a reference or in case it
// becomes useful at some point.
// See datastore.gorm.store

package store

import (
	"fmt"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/config"
)

type DeviceMemoryConfigStorer struct {
	configs map[uint64]config.DeviceConfig
	mutex   *sync.RWMutex
	DeviceConfigStorer
}

func NewDeviceMemoryConfigStorer(len int) DeviceConfigStorer {
	return &DeviceMemoryConfigStorer{
		configs: make(map[uint64]config.DeviceConfig, len),
		mutex:   &sync.RWMutex{}}
}

func (store *DeviceMemoryConfigStorer) Len() int {
	return len(store.configs)
}

func (store *DeviceMemoryConfigStorer) Put(farmID uint64, c config.DeviceConfig) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.configs[uint64(farmID)] = c
	return nil
}

func (store *DeviceMemoryConfigStorer) Get(id uint64, CONSISTENCY_LEVEL int) (config.DeviceConfig, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	if config, ok := store.configs[id]; ok {
		return config, nil
	}
	return nil, fmt.Errorf("DeviceMemoryConfigStorer id not found: %d", id)
}

func (store *DeviceMemoryConfigStorer) GetAll(farmID int) []config.DeviceConfig {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	configs := make([]config.DeviceConfig, len(store.configs))
	for k, v := range store.configs {
		configs[k] = v
	}
	return configs
}
