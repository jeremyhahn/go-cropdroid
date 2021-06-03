package dao

import (
	"github.com/jeremyhahn/cropdroid/config"
)

type OrganizationDAO interface {
	First() (config.OrganizationConfig, error)
	GetAll() ([]config.Organization, error)
	Get(orgID int) (config.OrganizationConfig, error)
	GetByUserID(userID int) ([]config.OrganizationConfig, error)
	//Find(orgID int) ([]config.Organization, error)
	Create(organization config.OrganizationConfig) error
	CreateUserRole(org config.OrganizationConfig, user config.UserConfig, role config.RoleConfig) error
}

type FarmDAO interface {
	Create(farm *config.Farm) error
	Save(farm *config.Farm) error
	First() (config.FarmConfig, error)
	Get(farmID int) (config.FarmConfig, error)
	GetAll() ([]config.Farm, error)
	GetByOrgAndUserID(orgID, userID int) ([]config.Farm, error)
}

type ControllerDAO interface {
	Save(controller config.ControllerConfig) error // Used only by integration test
	//GetByOrgId(orgId int) ([]config.Controller, error)
	GetByFarmId(orgId int) ([]config.Controller, error)
}

type ControllerConfigDAO interface {
	Save(config config.ControllerConfigConfig) error
	Get(controllerID int, name string) (*config.ControllerConfigItem, error)
	GetAll(controllerID int) ([]config.ControllerConfigItem, error)
}

type UserDAO interface {
	GetByEmail(email string) (config.UserConfig, error)
	Create(user config.UserConfig) error
	Save(user config.UserConfig) error // used by integration test only
}

type RoleDAO interface {
	Create(role config.RoleConfig) error // Used by integration test only
	Save(role config.RoleConfig) error   // Used by integration test only
	//GetByUserAndOrgID(userID, orgID int) (config.RoleConfig, error)
	GetByUserAndOrgID(userID, orgID int) ([]config.Role, error)
}

type AlgorithmDAO interface {
	Create(config.AlgorithmConfig) error // used by integration test only
	GetAll() ([]config.Algorithm, error)
}

type ChannelDAO interface {
	Save(channel config.ChannelConfig) error
	Get(channelID int) (config.ChannelConfig, error)
	GetByControllerID(controllerID int) ([]config.Channel, error)
	GetByOrgUserAndControllerID(orgID, userID, controllerID int) ([]config.Channel, error)
}

type ConditionDAO interface {
	Create(condition config.ConditionConfig) error
	Save(condition config.ConditionConfig) error
	Delete(condition config.ConditionConfig) error
	Get(id int) (config.ConditionConfig, error)
	GetByChannelID(id int) ([]config.Condition, error)
	GetByOrgUserAndChannelID(orgID, userID, channelID int) ([]config.Condition, error)
}

type ConfigDAO interface {
	Save(config config.ControllerConfigConfig) error
	Get(controllerID int, name string) (*config.ControllerConfigItem, error)
	GetAll(controllerID int) ([]config.ControllerConfigItem, error)
}

type MetricDAO interface {
	Save(metric config.MetricConfig) error
	Get(metricID int) (config.MetricConfig, error)
	GetByControllerID(controllerID int) ([]config.Metric, error) // Used to bootstrap sqlite config (configService.buildMetrics())
	GetByOrgUserAndControllerID(orgID, userID, controllerID int) ([]config.Metric, error)
}

type ScheduleDAO interface {
	Create(schedule config.ScheduleConfig) error
	Save(schedule config.ScheduleConfig) error
	Delete(schedule config.ScheduleConfig) error
	GetByChannelID(id int) ([]config.Schedule, error)
}
