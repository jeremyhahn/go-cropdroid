package state

import (
	"sync"

	"github.com/jeremyhahn/cropdroid/config"
)

type ControllerIndexMap struct {
	GenericStateStore
}

func NewControllerIndex(len int) ControllerIndex {
	return &ControllerIndexMap{
		GenericStateStore{
			items: make(map[int]interface{}, len),
			mutex: &sync.RWMutex{}}}
}

func CreateControllerIndex(items map[int]config.ControllerConfig) ControllerIndex {
	genericItems := make(map[int]interface{}, len(items))
	for i := range items {
		genericItems[i] = items[i]
	}
	return &ControllerIndexMap{
		GenericStateStore{
			items: genericItems,
			mutex: &sync.RWMutex{}}}
}

func (store *ControllerIndexMap) Len() int {
	return len(store.items)
}

func (store *ControllerIndexMap) Put(id int, v config.ControllerConfig) {
	store.GenericStateStore.Put(id, v)
}

func (store *ControllerIndexMap) Get(id int) (config.ControllerConfig, bool) {
	if item, ok := store.GenericStateStore.Get(id); ok {
		return item.(config.ControllerConfig), true
	}
	return nil, false
}

func (store *ControllerIndexMap) GetAll() []config.ControllerConfig {
	items := make([]config.ControllerConfig, len(store.GenericStateStore.items))
	i := 0
	for _, v := range store.GenericStateStore.items {
		items[i] = v.(config.ControllerConfig)
		i++
	}
	return items
}
