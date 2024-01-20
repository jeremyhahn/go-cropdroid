package datastore

import (
	"encoding/json"
	"errors"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
)

const (
	GORM_STORE = iota
	RAFT_STORE
	REDIS_TS
)

var (
	ErrNotFound          = errors.New("not found")
	ErrUnexpectedQuery   = errors.New("unexpected query")
	ErrMetricKeyNotFound = errors.New("metric key not found")
	//ErrOrganizationNotFound = errors.New("organization not found")
	//ErrOrganizationsNotFound = errors.New("organizations not found")
)

type Initializer interface {
	Initialize(includeFarm bool) error
	BuildConfig(orgID uint64, user *config.User,
		role common.Role) (*config.Farm, error)
}

type ChangefeedCallback func(Changefeed)

type Changefeeder interface {
	Subscribe(callback ChangefeedCallback)
}

type Changefeed interface {
	GetTable() string
	GetKey() int64
	GetValue() interface{}
	GetUpdated() string
	GetBytes() []byte
	GetRawMessage() map[string]*json.RawMessage
}

type DeviceDataStore interface {
	//CreateTable(tableName string, deviceState state.DeviceStateMap) error
	Save(deviceID uint64, deviceState state.DeviceStateMap) error
	GetLast30Days(deviceID uint64, metric string) ([]float64, error)
}
