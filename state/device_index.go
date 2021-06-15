package state

import (
	"sync"

	"github.com/jeremyhahn/go-cropdroid/config"
)

type DeviceIndexMap struct {
	BigGenericStateStore
}

func NewDeviceIndex(len int) DeviceIndex {
	return &DeviceIndexMap{
		BigGenericStateStore{
			items: make(map[uint64]interface{}, len),
			mutex: &sync.RWMutex{}}}
}

func CreateDeviceIndex(items map[uint64]config.DeviceConfig) DeviceIndex {
	genericItems := make(map[uint64]interface{}, len(items))
	for i := range items {
		genericItems[i] = items[i]
	}
	return &DeviceIndexMap{
		BigGenericStateStore{
			items: genericItems,
			mutex: &sync.RWMutex{}}}
}

func (store *DeviceIndexMap) Len() int {
	return len(store.items)
}

func (store *DeviceIndexMap) Put(id uint64, v config.DeviceConfig) {
	store.BigGenericStateStore.Put(id, v)
}

func (store *DeviceIndexMap) Get(id uint64) (config.DeviceConfig, bool) {
	if item, ok := store.BigGenericStateStore.Get(id); ok {
		return item.(config.DeviceConfig), true
	}
	return nil, false
}

func (store *DeviceIndexMap) GetAll() []config.DeviceConfig {
	items := make([]config.DeviceConfig, len(store.BigGenericStateStore.items))
	i := 0
	for _, v := range store.BigGenericStateStore.items {
		items[i] = v.(config.DeviceConfig)
		i++
	}
	return items
}
