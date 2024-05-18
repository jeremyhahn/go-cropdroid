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
	GetUsers(id uint64) ([]*config.User, error)
	GenericDAO[*config.Organization]
}

type FarmDAO interface {
	GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) ([]*config.Farm, error)
	GetByUserID(userID uint64, CONSISTENCY_LEVEL int) ([]*config.Farm, error)
	GenericDAO[*config.Farm]
}

type CustomerDAO interface {
	GetByEmail(name string, CONSISTENCY_LEVEL int) (*config.Customer, error)
	GenericDAO[*config.Customer]
}

type AlgorithmDAO interface {
	GenericDAO[*config.Algorithm]
}

type UserDAO interface {
	GenericDAO[*config.User]
}

type ChannelDAO interface {
	Save(farmID uint64, channel *config.Channel) error
	GetByDevice(orgID, farmID, deviceID uint64, CONSISTENCY_LEVEL int) ([]*config.Channel, error)
	Get(orgID, farmID, channelID uint64, CONSISTENCY_LEVEL int) (*config.Channel, error)
}

type RoleDAO interface {
	GetByName(name string, CONSISTENCY_LEVEL int) (*config.Role, error)
	GenericDAO[*config.Role]
}

type DeviceDAO interface {
	Save(device *config.Device) error
	Get(farmID, deviceID uint64, CONSISTENCY_LEVEL int) (*config.Device, error)
}

type EventLogDAO interface {
	GenericDAO[*entity.EventLog]
}

type PermissionDAO interface {
	Delete(permission *config.Permission) error
	GetFarms(orgID uint64, CONSISTENCY_LEVEL int) ([]*config.Farm, error)
	GetOrganizations(userID uint64, CONSISTENCY_LEVEL int) ([]*config.Organization, error)
	GetUsers(orgID uint64, CONSISTENCY_LEVEL int) ([]*config.User, error)
	Save(permission *config.Permission) error
	Update(permission *config.Permission) error
}

type DeviceSettingDAO interface {
	Save(farmID uint64, deviceSetting *config.DeviceSetting) error
	Get(farmID, deviceID uint64, name string, CONSISTENCY_LEVEL int) (*config.DeviceSetting, error)
}

type ConditionDAO interface {
	Save(farmID, deviceID uint64, condition *config.Condition) error
	Delete(farmID, deviceID uint64, condition *config.Condition) error
	Get(farmID, deviceID, channelID, conditionID uint64, CONSISTENCY_LEVEL int) (*config.Condition, error)
	GetByChannelID(farmID, deviceID, channelID uint64, CONSISTENCY_LEVEL int) ([]*config.Condition, error)
}

type MetricDAO interface {
	Save(farmID uint64, metric *config.Metric) error
	Get(farmID, deviceID uint64, metricID uint64, CONSISTENCY_LEVEL int) (*config.Metric, error)
	GetByDevice(farmID, deviceID uint64, CONSISTENCY_LEVEL int) ([]*config.Metric, error)
}

type ScheduleDAO interface {
	Save(farmID, deviceID uint64, schedule *config.Schedule) error
	Delete(farmID, deviceID uint64, schedule *config.Schedule) error
	GetByChannelID(farmID, deviceID, channelID uint64, CONSISTENCY_LEVEL int) ([]*config.Schedule, error)
}

type WorkflowDAO interface {
	Save(workflow *config.Workflow) error
	Delete(workflow *config.Workflow) error
	Get(farmID, workflowID uint64, CONSISTENCY_LEVEL int) (*config.Workflow, error)
	GetByFarmID(farmID uint64, CONSISTENCY_LEVEL int) ([]*config.Workflow, error)
}

type WorkflowStepDAO interface {
	Save(farmID uint64, workflowStep *config.WorkflowStep) error
	Delete(farmID uint64, workflowStep *config.WorkflowStep) error
	Get(farmID, workflowID, workflowStepID uint64, CONSISTENCY_LEVEL int) (*config.WorkflowStep, error)
	GetByWorkflowID(farmID, workflowID uint64, CONSISTENCY_LEVEL int) ([]*config.WorkflowStep, error)
}

type RegistrationDAO interface {
	GenericDAO[*config.Registration]
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
