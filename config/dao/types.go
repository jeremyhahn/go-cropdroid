package dao

import (
	"github.com/jeremyhahn/go-cropdroid/config"
)

type OrganizationDAO interface {
	First() (config.OrganizationConfig, error)
	GetAll() ([]config.Organization, error)
	Get(orgID uint64) (config.OrganizationConfig, error)
	GetByUserID(userID uint64) ([]config.OrganizationConfig, error)
	//Find(orgID int) ([]config.Organization, error)
	Create(organization config.OrganizationConfig) error
	CreateUserRole(org config.OrganizationConfig, user config.UserConfig, role config.RoleConfig) error
}

type FarmDAO interface {
	Count() (int64, error)
	Create(farm config.FarmConfig) error
	Delete(farm config.FarmConfig) error
	//DeleteById(farmID uint64) error
	First() (config.FarmConfig, error)
	Get(farmID uint64, CONSISTENCY_LEVEL int) (config.FarmConfig, error)
	GetAll() ([]config.Farm, error)
	GetByOrgAndUserID(orgID, userID uint64) ([]config.Farm, error)
	Save(farm config.FarmConfig) error
}

type DeviceDAO interface {
	Save(device config.DeviceConfig) error // Used only by integration test
	Get(id uint64) (config.DeviceConfig, error)
	//GetByOrgId(orgId int) ([]config.Device, error)
	GetByFarmId(orgId uint64) ([]config.Device, error)
	Count() (int64, error)
}

type DeviceConfigDAO interface {
	Save(config config.DeviceConfigConfig) error
	Get(deviceID uint64, name string) (*config.DeviceConfigItem, error)
	GetAll(deviceID uint64) ([]config.DeviceConfigItem, error)
}

type UserDAO interface {
	GetByID(userID uint64) (config.UserConfig, error)
	GetByEmail(email string) (config.UserConfig, error)
	Create(user config.UserConfig) error
	Save(user config.UserConfig) error // used by integration test only
}

type RoleDAO interface {
	Create(role config.RoleConfig) error // Used by integration test only
	Save(role config.RoleConfig) error   // Used by integration test only
	//GetByUserAndOrgID(userID, orgID int) (config.RoleConfig, error)
	GetByUserAndOrgID(userID, orgID int) ([]config.Role, error)
	GetByName(name string) (config.RoleConfig, error)
}

type AlgorithmDAO interface {
	Create(config.AlgorithmConfig) error // used by integration test only
	GetAll() ([]config.Algorithm, error)
}

type ChannelDAO interface {
	Save(channel config.ChannelConfig) error
	Get(channelID uint64) (config.ChannelConfig, error)
	GetByDeviceID(deviceID uint64) ([]config.Channel, error)
	GetByOrgUserAndDeviceID(orgID, userID, deviceID uint64) ([]config.Channel, error)
}

type ConditionDAO interface {
	Create(condition config.ConditionConfig) error
	Save(condition config.ConditionConfig) error
	Delete(condition config.ConditionConfig) error
	Get(id uint64) (config.ConditionConfig, error)
	GetByChannelID(id uint64) ([]config.Condition, error)
	GetByOrgUserAndChannelID(orgID, userID, channelID uint64) ([]config.Condition, error)
}

type ConfigDAO interface {
	Save(config config.DeviceConfigConfig) error
	Get(deviceID int, name string) (*config.DeviceConfigItem, error)
	GetAll(deviceID int) ([]config.DeviceConfigItem, error)
}

type MetricDAO interface {
	Save(metric config.MetricConfig) error
	Get(metricID int) (config.MetricConfig, error)
	GetByDeviceID(deviceID uint64) ([]config.Metric, error) // Used to bootstrap sqlite config (configService.buildMetrics())
	GetByOrgUserAndDeviceID(orgID, userID, deviceID uint64) ([]config.Metric, error)
}

type ScheduleDAO interface {
	Create(schedule config.ScheduleConfig) error
	Save(schedule config.ScheduleConfig) error
	Delete(schedule config.ScheduleConfig) error
	GetByChannelID(id uint64) ([]config.Schedule, error)
}

type WorkflowDAO interface {
	Create(condition config.WorkflowConfig) error
	Save(condition config.WorkflowConfig) error
	Delete(condition config.WorkflowConfig) error
	Get(id uint64) (config.WorkflowConfig, error)
	//GetAll(farmID uint64) ([]config.Workflow, error)
	GetByFarmID(id uint64) ([]config.Workflow, error)
}

type WorkflowStepDAO interface {
	Create(condition config.WorkflowStepConfig) error
	Save(condition config.WorkflowStepConfig) error
	Delete(condition config.WorkflowStepConfig) error
	Get(id uint64) (config.WorkflowStepConfig, error)
	GetByWorkflowID(id uint64) ([]config.WorkflowStep, error)
}
