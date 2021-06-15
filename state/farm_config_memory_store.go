// +build ignore

// This has been deprecated in favor of config.store.FarmConfigStorer

package state

import (
	"errors"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/config"
)

type MemoryConfigStore struct {
	configs map[uint64]config.FarmConfig
	mutex   *sync.RWMutex
	ConfigStorer
}

func NewMemoryConfigStore(len int) ConfigStorer {
	return &MemoryConfigStore{
		configs: make(map[uint64]config.FarmConfig, len),
		mutex:   &sync.RWMutex{}}
}

func (store *MemoryConfigStore) Len() int {
	return len(store.configs)
}

func (store *MemoryConfigStore) Put(id uint64, c config.FarmConfig) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.configs[id] = c
	return nil
}

func (store *MemoryConfigStore) Get(id uint64) (config.FarmConfig, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	if config, ok := store.configs[id]; ok {
		return config, nil
	}
	return nil, errors.New("Config id not found")
}

func (store *MemoryConfigStore) GetAll() []config.FarmConfig {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	configs := make([]config.FarmConfig, len(store.configs))
	for k, v := range store.configs {
		configs[k] = v
	}
	return configs
}
