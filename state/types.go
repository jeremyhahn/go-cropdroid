package state

import "github.com/jeremyhahn/cropdroid/config"

type FarmStorer interface {
	Len() int
	Put(farmID int, v FarmStateMap) error
	Get(farmID int) (FarmStateMap, error)
	GetAll() []*StoreViewItem
}

type ConfigStorer interface {
	Len() int
	Put(id uint64, v config.FarmConfig) error
	Get(id uint64) (config.FarmConfig, error)
	GetAll() []config.FarmConfig
}

type StateStore interface {
	Len() int
	Put(id int, v interface{})
	Get(id int) (interface{}, bool)
	GetAll() []interface{}
}

type ControllerIndex interface {
	Len() int
	Put(id int, v config.ControllerConfig)
	Get(id int) (config.ControllerConfig, bool)
	GetAll() []config.ControllerConfig
}

type ChannelIndex interface {
	Len() int
	Put(id int, v config.ChannelConfig)
	Get(id int) (config.ChannelConfig, bool)
	GetAll() []config.ChannelConfig
}
