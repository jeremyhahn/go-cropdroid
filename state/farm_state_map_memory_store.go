package state

import (
	"fmt"
	"sync"
	"time"

	logging "github.com/op/go-logging"
)

var farmStoreMutex = &sync.RWMutex{}

type StoreViewItem struct {
	ID         int          `json:"id"`
	State      FarmStateMap `json:"state"`
	LastAccess int64        `json:"lastAccess"`
}

type storeItem struct {
	state      FarmStateMap
	lastAccess int64
}

type FarmStore struct {
	logger *logging.Logger
	farms  map[int]storeItem
	mutex  *sync.RWMutex
	FarmStorer
}

func NewMemoryFarmStore(logger *logging.Logger, len, ttl int, gcTicker time.Duration) FarmStorer {
	farmStoreMutex.Lock()
	defer farmStoreMutex.Unlock()
	appstate := &FarmStore{
		logger: logger,
		farms:  make(map[int]storeItem, len),
		mutex:  farmStoreMutex}
	if ttl > 0 {
		go func() {
			for now := range time.Tick(gcTicker) {
				appstate.mutex.Lock()
				for k, v := range appstate.farms {
					if now.Unix()-v.lastAccess > int64(ttl) {
						delete(appstate.farms, k)
					}
				}
				appstate.mutex.Unlock()
			}
		}()
	}
	return appstate
}

func (store *FarmStore) Len() int {
	return len(store.farms)
}

func (store *FarmStore) Put(farmID int, v FarmStateMap) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	item, ok := store.farms[farmID]
	if !ok {
		item = storeItem{state: v}
		store.farms[farmID] = item
		//store.logger.Errorf("Storing farm: farm.id=%d", farmID)
	}
	item.lastAccess = time.Now().Unix()
	return nil
}

func (store *FarmStore) Get(farmID int) (FarmStateMap, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	if farm, ok := store.farms[farmID]; ok {
		state := farm.state
		farm.lastAccess = time.Now().Unix()
		//store.logger.Errorf("Returning stored farm: farm.id=%d", farmID)
		return state, nil
	}
	return nil, fmt.Errorf("FarmID not found in app state: %d", farmID)
}

func (store *FarmStore) GetAll() []*StoreViewItem {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	farms := make([]*StoreViewItem, len(store.farms))
	i := 0
	for k, v := range store.farms {
		farms[i] = &StoreViewItem{
			ID:         k,
			State:      v.state,
			LastAccess: v.lastAccess}
	}
	return farms
}
