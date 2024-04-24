package dao

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
)

type CustomerDAO interface {
	Delete(user *config.Customer) error
	Get(userID uint64, CONSISTENCY_LEVEL int) (*config.Customer, error)
	GetByEmail(name string, CONSISTENCY_LEVEL int) (*config.Customer, error)
	GetByProcessorID(id string, CONSISTENCY_LEVEL int) (*config.Customer, error)
	GetAll(CONSISTENCY_LEVEL int) ([]*config.Customer, error)
	Update(customer *config.Customer) error
	Save(user *config.Customer) error
}

type UserDAO interface {
	Delete(user *config.User) error
	Get(userID uint64, CONSISTENCY_LEVEL int) (*config.User, error)
	Save(user *config.User) error
}

type RoleDAO interface {
	Delete(role *config.Role) error
	Get(roleID uint64, CONSISTENCY_LEVEL int) (*config.Role, error)
	GetAll(CONSISTENCY_LEVEL int) ([]*config.Role, error)
	GetByName(name string, CONSISTENCY_LEVEL int) (*config.Role, error)
	Save(role *config.Role) error
}

type PermissionDAO interface {
	Delete(permission *config.Permission) error
	GetFarms(orgID uint64, CONSISTENCY_LEVEL int) ([]*config.Farm, error)
	GetOrganizations(userID uint64, CONSISTENCY_LEVEL int) ([]*config.Organization, error)
	GetUsers(orgID uint64, CONSISTENCY_LEVEL int) ([]*config.User, error)
	Save(permission *config.Permission) error
	Update(permission *config.Permission) error
}

type OrganizationDAO interface {
	Delete(organization *config.Organization) error
	Get(id uint64, CONSISTENCY_LEVEL int) (*config.Organization, error)
	GetAll(CONSISTENCY_LEVEL int) ([]*config.Organization, error)
	GetUsers(id uint64) ([]*config.User, error)
	Save(organization *config.Organization) error
}

type FarmDAO interface {
	Delete(farm *config.Farm) error
	Get(farmID uint64, CONSISTENCY_LEVEL int) (*config.Farm, error)
	GetAll(CONSISTENCY_LEVEL int) ([]*config.Farm, error)
	GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) ([]*config.Farm, error)
	GetByUserID(userID uint64, CONSISTENCY_LEVEL int) ([]*config.Farm, error)
	Save(farm *config.Farm) error
}

type DeviceDAO interface {
	Save(device *config.Device) error
	Get(farmID, deviceID uint64, CONSISTENCY_LEVEL int) (*config.Device, error)
}

type DeviceSettingDAO interface {
	Save(farmID uint64, deviceSetting *config.DeviceSetting) error
	Get(farmID, deviceID uint64, name string, CONSISTENCY_LEVEL int) (*config.DeviceSetting, error)
}

type ChannelDAO interface {
	Save(farmID uint64, channel *config.Channel) error
	GetByDevice(orgID, farmID, deviceID uint64, CONSISTENCY_LEVEL int) ([]*config.Channel, error)
	Get(orgID, farmID, channelID uint64, CONSISTENCY_LEVEL int) (*config.Channel, error)
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

type AlgorithmDAO interface {
	Save(clgorithm *config.Algorithm) error // used by test only
	GetAll(CONSISTENCY_LEVEL int) ([]*config.Algorithm, error)
}

type RegistrationDAO interface {
	Save(registration *config.Registration) error
	Get(registrationID uint64, CONSISTENCY_LEVEL int) (*config.Registration, error)
	Delete(registration *config.Registration) error
}

type EventLogDAO interface {
	Save(eventLog *entity.EventLog) error
	GetAll(CONSISTENCY_LEVEL int) ([]*entity.EventLog, error)
	GetAllDesc(CONSISTENCY_LEVEL int) ([]*entity.EventLog, error)
	GetPage(CONSISTENCY_LEVEL int, page, size int64) ([]*entity.EventLog, error)
	Count(CONSISTENCY_LEVEL int) (int64, error)
}

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
