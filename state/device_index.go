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

func CreateDeviceIndex(items map[uint64]config.Device) DeviceIndex {
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

func (store *DeviceIndexMap) Put(id uint64, v config.Device) {
	store.BigGenericStateStore.Put(id, v)
}

func (store *DeviceIndexMap) Get(id uint64) (config.Device, bool) {
	if item, ok := store.BigGenericStateStore.Get(id); ok {
		return item.(config.Device), true
	}
	return &config.DeviceStruct{}, false
}

func (store *DeviceIndexMap) GetAll() []config.Device {
	items := make([]config.Device, len(store.BigGenericStateStore.items))
	i := 0
	for _, v := range store.BigGenericStateStore.items {
		items[i] = v.(config.Device)
		i++
	}
	return items
}
