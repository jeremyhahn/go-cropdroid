package state

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/config"
)

const (
	MEMORY_STORE = iota
	GORM_STORE
	RAFT_STORE
)

var (
	ErrFarmNotFound = errors.New("farm not found in state store")
)

// Used by farm state storage implementations
type FarmStorer interface {
	Len() int
	Put(farmID uint64, v FarmStateMap) error
	Get(farmID uint64) (FarmStateMap, error)
	GetAll() []*StoreViewItem
}

// Used by device state storage implementations
type DeviceStorer interface {
	Len() int
	Put(deviceID uint64, v DeviceStateMap) error
	Get(deviceID uint64) (DeviceStateMap, error)
	GetAll() []*DeviceStoreViewItem
}

// Used by generic state store implementation
type StateStore interface {
	Len() int
	Put(id int, v interface{})
	Get(id int) (interface{}, bool)
	GetAll() []interface{}
}

// Used by big generic state store implementation
type BigStateStore interface {
	Len() int
	Put(id uint64, v interface{})
	Get(id uint64) (interface{}, bool)
	GetAll() []interface{}
}

// Device index stores a map of id to object pointers
// to provide O(n) constant time lookups for fast retrievals.
//
// Without this index, searches have to loop over every
// device in every farm to find the it.
type DeviceIndex interface {
	Len() int
	Put(id uint64, v config.DeviceConfig)
	Get(id uint64) (config.DeviceConfig, bool)
	GetAll() []config.DeviceConfig
}

// Channel index stores a map of id to object pointers
// to provide O(n) constant time lookups for fast retrievals.
//
// Without this index, searches have to loop over every
// channel in every device to find it.
type ChannelIndex interface {
	Len() int
	Put(id int, v config.ChannelConfig)
	Get(id int) (config.ChannelConfig, bool)
	GetAll() []config.ChannelConfig
}
