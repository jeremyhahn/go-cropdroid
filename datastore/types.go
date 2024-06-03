package datastore

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
)

const (
	GORM_STORE = iota
	RAFT_STORE
	REDIS_TS
)

var (
	ErrRecordNotFound    = errors.New("record not found")
	ErrUnexpectedQuery   = errors.New("unexpected query")
	ErrMetricKeyNotFound = errors.New("metric key not found")
	ErrNullEntityId      = errors.New("null entity id")
	//ErrOrganizationNotFound = errors.New("organization not found")
	//ErrOrganizationsNotFound = errors.New("organizations not found")
)

type Initializer interface {
	Initialize(includeFarm bool) error
	BuildConfig(orgID uint64, user *config.UserStruct, role config.Role) (config.Farm, error)
}

type DeviceDataStore interface {
	Save(deviceID uint64, deviceState state.DeviceStateMap) error
	GetLast30Days(deviceID uint64, metric string) ([]float64, error)
}
