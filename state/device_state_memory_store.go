package state

import (
	"fmt"
	"sync"
	"time"

	logging "github.com/op/go-logging"
)

var deviceStoreMutex = &sync.RWMutex{}

type DeviceStoreViewItem struct {
	ID         uint64         `json:"id"`
	State      DeviceStateMap `json:"state"`
	LastAccess int64          `json:"lastAccess"`
}

type deviceStoreItem struct {
	state      DeviceStateMap
	lastAccess int64
}

type DeviceStore struct {
	logger       *logging.Logger
	devices      map[uint64]deviceStoreItem
	mutex        *sync.RWMutex
	gcTickerStop bool
	DeviceStorer
}

// Creates a new memory based state store for devices. This store holds
// a single DeviceStateMap that represents the current state of a single device.
func NewMemoryDeviceStore(logger *logging.Logger, len, ttl int,
	gcTicker time.Duration) DeviceStorer {

	deviceStoreMutex.Lock()
	defer deviceStoreMutex.Unlock()

	appstate := &DeviceStore{
		logger:       logger,
		devices:      make(map[uint64]deviceStoreItem, len),
		mutex:        deviceStoreMutex,
		gcTickerStop: false}

	if ttl > 0 {
		go func() {
			logger.Debugf("Device state store gcTicker started")
			for now := range time.Tick(gcTicker) {
				appstate.mutex.Lock()
				for k, v := range appstate.devices {
					if now.Unix()-v.lastAccess > int64(ttl) {
						delete(appstate.devices, k)
					}
				}
				appstate.mutex.Unlock()
				if appstate.gcTickerStop {
					logger.Debugf("Device state store gcTicker stopped")
					return
				}
			}
		}()
	}
	return appstate
}

func (store *DeviceStore) Len() int {
	return len(store.devices)
}

func (store *DeviceStore) Put(deviceID uint64, v DeviceStateMap) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	item, ok := store.devices[deviceID]
	if !ok {
		item = deviceStoreItem{state: v}
		store.devices[deviceID] = item
		//store.logger.Errorf("Storing device: device.id=%d", deviceID)
	}
	item.lastAccess = time.Now().Unix()
	return nil
}

func (store *DeviceStore) Get(deviceID uint64) (DeviceStateMap, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	if device, ok := store.devices[deviceID]; ok {
		state := device.state
		device.lastAccess = time.Now().Unix()
		//store.logger.Errorf("Returning stored device: device.id=%d", deviceID)
		return state, nil
	}
	return nil, fmt.Errorf("DeviceID not found in app state: %d", deviceID)
}

func (store *DeviceStore) GetAll() []*DeviceStoreViewItem {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	devices := make([]*DeviceStoreViewItem, len(store.devices))
	i := 0
	for k, v := range store.devices {
		devices[i] = &DeviceStoreViewItem{
			ID:         k,
			State:      v.state,
			LastAccess: v.lastAccess}
	}
	return devices
}

func (store *DeviceStore) Close() {
	store.gcTickerStop = true
	store.logger.Debugf("Stopping device state memory store gcTicker")
}
