package dao

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/entity"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
)

type Pager[E any] interface {
	GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (PageResult[E], error)
	ForEachPage(pageQuery query.PageQuery, pagerProcFunc query.PagerProcFunc[E], CONSISTENCY_LEVEL int) error
}

type GenericDAO[E any] interface {
	Save(entity E) error
	Get(id uint64, CONSISTENCY_LEVEL int) (E, error)
	Delete(entity E) error
	Count(CONSISTENCY_LEVEL int) (int64, error)
	// Removing this from the Generic interface and will only
	// implement Update on GORM DAO's that require it. Raft does
	// not use updates
	// Update(entity E) error
	Pager[E]
}

type ServerDAO interface {
	GenericDAO[*config.Server]
}

type OrganizationDAO interface {
	GetUsers(id uint64) ([]*config.UserStruct, error)
	GenericDAO[*config.OrganizationStruct]
}

type FarmDAO interface {
	GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) ([]*config.FarmStruct, error)
	GetByUserID(userID uint64, CONSISTENCY_LEVEL int) ([]*config.FarmStruct, error)
	GenericDAO[*config.FarmStruct]
}

type CustomerDAO interface {
	GetByEmail(name string, CONSISTENCY_LEVEL int) (*config.CustomerStruct, error)
	GenericDAO[*config.CustomerStruct]
}

type AlgorithmDAO interface {
	GenericDAO[*config.AlgorithmStruct]
}

type UserDAO interface {
	GenericDAO[*config.UserStruct]
}

type ChannelDAO interface {
	Save(farmID uint64, channel *config.ChannelStruct) error
	GetByDevice(orgID, farmID, deviceID uint64, CONSISTENCY_LEVEL int) ([]*config.ChannelStruct, error)
	Get(orgID, farmID, channelID uint64, CONSISTENCY_LEVEL int) (*config.ChannelStruct, error)
}

type RoleDAO interface {
	GetByName(name string, CONSISTENCY_LEVEL int) (*config.RoleStruct, error)
	GenericDAO[*config.RoleStruct]
}

type DeviceDAO interface {
	Save(device *config.DeviceStruct) error
	Get(farmID, deviceID uint64, CONSISTENCY_LEVEL int) (*config.DeviceStruct, error)
}

type EventLogDAO interface {
	GenericDAO[*entity.EventLog]
}

type PermissionDAO interface {
	Delete(permission *config.PermissionStruct) error
	GetFarms(orgID uint64, CONSISTENCY_LEVEL int) ([]*config.FarmStruct, error)
	GetOrganizations(userID uint64, CONSISTENCY_LEVEL int) ([]*config.OrganizationStruct, error)
	GetUsers(orgID uint64, CONSISTENCY_LEVEL int) ([]*config.UserStruct, error)
	Save(permission *config.PermissionStruct) error
	Update(permission *config.PermissionStruct) error
}

type DeviceSettingDAO interface {
	Save(farmID uint64, deviceSetting *config.DeviceSettingStruct) error
	Get(farmID, deviceID uint64, name string, CONSISTENCY_LEVEL int) (*config.DeviceSettingStruct, error)
}

type ConditionDAO interface {
	Save(farmID, deviceID uint64, condition *config.ConditionStruct) error
	Delete(farmID, deviceID uint64, condition *config.ConditionStruct) error
	Get(farmID, deviceID, channelID, conditionID uint64, CONSISTENCY_LEVEL int) (*config.ConditionStruct, error)
	GetByChannelID(farmID, deviceID, channelID uint64, CONSISTENCY_LEVEL int) ([]*config.ConditionStruct, error)
}

type MetricDAO interface {
	Save(farmID uint64, metric *config.MetricStruct) error
	Get(farmID, deviceID uint64, metricID uint64, CONSISTENCY_LEVEL int) (*config.MetricStruct, error)
	GetByDevice(farmID, deviceID uint64, CONSISTENCY_LEVEL int) ([]*config.MetricStruct, error)
}

type ScheduleDAO interface {
	Save(farmID, deviceID uint64, schedule *config.ScheduleStruct) error
	Delete(farmID, deviceID uint64, schedule *config.ScheduleStruct) error
	GetByChannelID(farmID, deviceID, channelID uint64, CONSISTENCY_LEVEL int) ([]*config.ScheduleStruct, error)
}

type WorkflowDAO interface {
	Save(workflow *config.WorkflowStruct) error
	Delete(workflow *config.WorkflowStruct) error
	Get(farmID, workflowID uint64, CONSISTENCY_LEVEL int) (*config.WorkflowStruct, error)
	GetByFarmID(farmID uint64, CONSISTENCY_LEVEL int) ([]*config.WorkflowStruct, error)
}

type WorkflowStepDAO interface {
	Save(farmID uint64, workflowStep *config.WorkflowStepStruct) error
	Delete(farmID uint64, workflowStep *config.WorkflowStepStruct) error
	Get(farmID, workflowID, workflowStepID uint64, CONSISTENCY_LEVEL int) (*config.WorkflowStepStruct, error)
	GetByWorkflowID(farmID, workflowID uint64, CONSISTENCY_LEVEL int) ([]*config.WorkflowStepStruct, error)
}

type RegistrationDAO interface {
	GenericDAO[*config.RegistrationStruct]
}

// type EventLogDAO interface {
// 	Save(eventLog *entity.EventLog) error
// 	GetAll(CONSISTENCY_LEVEL int) ([]*entity.EventLog, error)
// 	GetAllDesc(CONSISTENCY_LEVEL int) ([]*entity.EventLog, error)
// 	GetPage(page, size, CONSISTENCY_LEVEL int) ([]*entity.EventLog, error)
// 	Count(CONSISTENCY_LEVEL int) (int64, error)
// }

type Registry interface {
	GetOrganizationDAO() OrganizationDAO
	SetOrganizationDAO(dao OrganizationDAO)
	GetFarmDAO() FarmDAO
	SetFarmDAO(dao FarmDAO)
	NewFarmDAO() FarmDAO
	GetDeviceDAO() DeviceDAO
	NewDeviceDAO() DeviceDAO
	SetDeviceDAO(dao DeviceDAO)
	GetDeviceSettingDAO() DeviceSettingDAO
	SetDeviceSettingDAO(dao DeviceSettingDAO)
	GetMetricDAO() MetricDAO
	SetMetricDAO(dao MetricDAO)
	GetChannelDAO() ChannelDAO
	SetChannelDAO(dao ChannelDAO)
	GetScheduleDAO() ScheduleDAO
	SetScheduleDAO(dao ScheduleDAO)
	GetConditionDAO() ConditionDAO
	SetConditionDAO(dao ConditionDAO)
	GetAlgorithmDAO() AlgorithmDAO
	SetAlgorithmDAO(dao AlgorithmDAO)
	GetUserDAO() UserDAO
	SetUserDAO(UserDAO)
	GetPermissionDAO() PermissionDAO
	SetPermissiondAO(regDAO PermissionDAO)
	GetRegistrationDAO() RegistrationDAO
	SetRegistrationDAO(regDAO RegistrationDAO)
	GetRoleDAO() RoleDAO
	SetRoleDAO(RoleDAO)
	GetCustomerDAO() CustomerDAO
	SetCustomerDAO(CustomerDAO)
	GetWorkflowDAO() WorkflowDAO
	SetWorkflowDAO(WorkflowDAO)
	GetWorkflowStepDAO() WorkflowStepDAO
	SetWorkflowStepDAO(WorkflowStepDAO)
	GetEventLogDAO() EventLogDAO
	SetEventLogDAO(dao EventLogDAO)
}
