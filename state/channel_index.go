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

func CreateChannelIndex(items map[uint64]config.Channel) ChannelIndex {
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

func (store *ChannelIndexMap) Put(id uint64, v config.Channel) {
	store.GenericStateStore.Put(uint64(id), v)
}

func (store *ChannelIndexMap) Get(id uint64) (config.Channel, bool) {
	if item, ok := store.GenericStateStore.Get(id); ok {
		return item.(config.Channel), true
	}
	return &config.ChannelStruct{}, false
}

func (store *ChannelIndexMap) GetAll() []config.Channel {
	items := make([]config.Channel, len(store.items))
	for k, v := range store.GenericStateStore.items {
		items[k] = v.(config.Channel)
	}
	return items
}
