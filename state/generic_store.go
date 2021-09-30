package state

import (
	"sync"
)

type GenericStateStore struct {
	items map[uint64]interface{}
	mutex *sync.RWMutex
	StateStore
}

func NewGenericStore(len int) StateStore {
	return &GenericStateStore{
		items: make(map[uint64]interface{}, len),
		mutex: &sync.RWMutex{}}
}

func (store *GenericStateStore) Len() int {
	return len(store.items)
}

func (store *GenericStateStore) Put(id uint64, v interface{}) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.items[id] = v
}

func (store *GenericStateStore) Get(id uint64) (interface{}, bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	if item, ok := store.items[id]; ok {
		return item, true
	}
	return nil, false
}

func (store *GenericStateStore) GetAll() []interface{} {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	items := make([]interface{}, len(store.items))
	for k, v := range store.items {
		items[k] = v
	}
	return items
}
