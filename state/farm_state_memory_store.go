package state

import (
	"sync"
	"time"

	logging "github.com/op/go-logging"
)

var (
	farmStoreMutex = &sync.RWMutex{}
)

type StoreViewItem struct {
	ID         uint64       `json:"id"`
	State      FarmStateMap `json:"state"`
	LastAccess int64        `json:"lastAccess"`
}

type storeItem struct {
	state      FarmStateMap
	lastAccess int64
}

type FarmStore struct {
	logger       *logging.Logger
	farms        map[uint64]storeItem
	mutex        *sync.RWMutex
	gcTickerStop bool
	FarmStateStorer
}

// Creates a new memory based state store to hold farms. This store holds
// a single FarmStateMap that represents the current state of a farm and
// it's associated device states.
func NewMemoryFarmStore(logger *logging.Logger, len, ttl int, gcTicker time.Duration) FarmStateStorer {
	farmStoreMutex.Lock()
	defer farmStoreMutex.Unlock()
	appstate := &FarmStore{
		logger:       logger,
		farms:        make(map[uint64]storeItem, len),
		mutex:        farmStoreMutex,
		gcTickerStop: false}
	if ttl > 0 {
		go func() {
			logger.Debugf("Farm state store gcTicker started")
			for now := range time.Tick(gcTicker) {
				appstate.mutex.Lock()
				for k, v := range appstate.farms {
					if now.Unix()-v.lastAccess > int64(ttl) {
						delete(appstate.farms, k)
					}
				}
				appstate.mutex.Unlock()
				if appstate.gcTickerStop {
					logger.Debugf("Farm state store gcTicker stopped")
					return
				}
			}
		}()
	}
	return appstate
}

func (store *FarmStore) Len() int {
	return len(store.farms)
}

func (store *FarmStore) Put(farmID uint64, v FarmStateMap) error {
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

func (store *FarmStore) Get(farmID uint64) (FarmStateMap, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	if farm, ok := store.farms[farmID]; ok {
		state := farm.state
		farm.lastAccess = time.Now().Unix()
		//store.logger.Errorf("Returning stored farm: farm.id=%d", farmID)
		return state, nil
	}
	return nil, ErrFarmNotFound
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

func (store *FarmStore) Close() {
	store.gcTickerStop = true
	store.logger.Debugf("Stopping farm state memory store gcTicker")
}
