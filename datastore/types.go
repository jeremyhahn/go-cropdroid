package datastore

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/state"
)

const (
	GORM_STORE = iota
	RAFT_STORE
	REDIS_TS
)

type Initializer interface {
	Initialize(includeFarmConfig bool) error
	BuildConfig(config.UserConfig, config.RoleConfig) (config.FarmConfig, error)
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

type DatastoreRegistry interface {
	GetOrganizationDAO() dao.OrganizationDAO
	SetOrganizationDAO(dao dao.OrganizationDAO)
	GetFarmDAO() dao.FarmDAO
	SetFarmDAO(dao dao.FarmDAO)
	NewFarmDAO() dao.FarmDAO
	GetDeviceDAO() dao.DeviceDAO
	NewDeviceDAO() dao.DeviceDAO
	SetDeviceDAO(dao dao.DeviceDAO)
	GetDeviceConfigDAO() dao.DeviceConfigDAO
	SetDeviceConfigDAO(dao dao.DeviceConfigDAO)
	GetMetricDAO() dao.MetricDAO
	SetMetricDAO(dao dao.MetricDAO)
	GetChannelDAO() dao.ChannelDAO
	SetChannelDAO(dao dao.ChannelDAO)
	GetScheduleDAO() dao.ScheduleDAO
	SetScheduleDAO(dao dao.ScheduleDAO)
	GetConditionDAO() dao.ConditionDAO
	SetConditionDAO(dao dao.ConditionDAO)
	GetAlgorithmDAO() dao.AlgorithmDAO
	SetAlgorithmDAO(dao dao.AlgorithmDAO)
	GetUserDAO() dao.UserDAO
	SetUserDAO(dao.UserDAO)
	GetPermissionDAO() dao.PermissionDAO
	SetPermissiondAO(regDAO dao.PermissionDAO)
	GetRegistrationDAO() dao.RegistrationDAO
	SetRegistrationDAO(regDAO dao.RegistrationDAO)
	GetRoleDAO() dao.RoleDAO
	SetRoleDAO(dao.RoleDAO)
	GetWorkflowDAO() dao.WorkflowDAO
	SetWorkflowDAO(dao.WorkflowDAO)
	GetWorkflowStepDAO() dao.WorkflowStepDAO
	SetWorkflowStepDAO(dao.WorkflowStepDAO)
}
