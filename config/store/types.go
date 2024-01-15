package store

import "github.com/jeremyhahn/go-cropdroid/config"

type OrganizationStorer interface {
	Get(organizationID uint64, CONSISTENCY_LEVEL int) (config.Organization, error)
	GetAll(CONSISTENCY_LEVEL int) []config.Farm
	GetByIds(organizationID []uint64, CONSISTENCY_LEVEL int) []config.Organization
	Len() int
	Put(organizationID uint64, organization config.Organization) error
}

type FarmStorer interface {
	Cache(farmID uint64, farm config.Farm)
	Get(farmID uint64, CONSISTENCY_LEVEL int) (config.Farm, error)
	GetAll() []config.Farm
	GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) []config.Farm
	Len() int
	Put(farmID uint64, farm config.Farm) error
}

type DeviceStorer interface {
	Cache(deviceID uint64, farm config.Device)
	Get(deviceID uint64, CONSISTENCY_LEVEL int) (config.Device, error)
	GetAll(deviceID uint64) []config.Device
	Len() int
	Put(deviceID uint64, device config.Device) error
}
