package state

import (
	"sync"

	"github.com/jeremyhahn/go-cropdroid/config"
)

type ChannelIndexMap struct {
	GenericStateStore
}

func NewChannelIndex(len int) ChannelIndex {
	return &ChannelIndexMap{
		GenericStateStore{
			items: make(map[uint64]interface{}, len),
			mutex: &sync.RWMutex{}}}
}

func CreateChannelIndex(items map[uint64]config.ChannelConfig) ChannelIndex {
	genericItems := make(map[uint64]interface{}, len(items))
	for i := range items {
		genericItems[i] = items[i]
	}
	return &ChannelIndexMap{
		GenericStateStore{
			items: genericItems,
			mutex: &sync.RWMutex{}}}
}

func (store *ChannelIndexMap) Len() int {
	return len(store.items)
}

func (store *ChannelIndexMap) Put(id uint64, v config.ChannelConfig) {
	store.GenericStateStore.Put(uint64(id), v)
}

func (store *ChannelIndexMap) Get(id uint64) (config.ChannelConfig, bool) {
	if item, ok := store.GenericStateStore.Get(id); ok {
		return item.(config.ChannelConfig), true
	}
	return nil, false
}

func (store *ChannelIndexMap) GetAll() []config.ChannelConfig {
	items := make([]config.ChannelConfig, len(store.items))
	for k, v := range store.GenericStateStore.items {
		items[k] = v.(config.ChannelConfig)
	}
	return items
}
