package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormDaoRegistry struct {
	orgDAO              dao.OrganizationDAO
	farmDAO             dao.FarmDAO
	deviceDAO       dao.DeviceDAO
	deviceConfigDAO dao.DeviceConfigDAO
	metricDAO           dao.MetricDAO
	channelDAO          dao.ChannelDAO
	scheduleDAO         dao.ScheduleDAO
	conditionDAO        dao.ConditionDAO
	algorithmDAO        dao.AlgorithmDAO
	eventLogDAO         EventLogDAO
	userDAO             dao.UserDAO
	roleDAO             dao.RoleDAO
	datastore.DatastoreRegistry
}

func NewGormRegistry(logger *logging.Logger, db *gorm.DB) datastore.DatastoreRegistry {
	return &GormDaoRegistry{
		orgDAO:              NewOrganizationDAO(logger, db),
		farmDAO:             NewFarmDAO(logger, db),
		deviceDAO:       NewDeviceDAO(logger, db),
		deviceConfigDAO: NewDeviceConfigDAO(logger, db),
		metricDAO:           NewMetricDAO(logger, db),
		channelDAO:          NewChannelDAO(logger, db),
		scheduleDAO:         NewScheduleDAO(logger, db),
		conditionDAO:        NewConditionDAO(logger, db),
		algorithmDAO:        NewAlgorithmDAO(logger, db),
		eventLogDAO:         NewEventLogDAO(logger, db),
		userDAO:             NewUserDAO(logger, db),
		roleDAO:             NewRoleDAO(logger, db)}
}

func (registry *GormDaoRegistry) GetOrganizationDAO() dao.OrganizationDAO {
	return registry.orgDAO
}

func (registry *GormDaoRegistry) SetOrganizationDAO(dao dao.OrganizationDAO) {
	registry.orgDAO = dao
}

func (registry *GormDaoRegistry) GetFarmDAO() dao.FarmDAO {
	return registry.farmDAO
}

func (registry *GormDaoRegistry) SetFarmDAO(dao dao.FarmDAO) {
	registry.farmDAO = dao
}

func (registry *GormDaoRegistry) GetDeviceDAO() dao.DeviceDAO {
	return registry.deviceDAO
}

func (registry *GormDaoRegistry) SetDeviceDAO(dao dao.DeviceDAO) {
	registry.deviceDAO = dao
}

func (registry *GormDaoRegistry) GetDeviceConfigDAO() dao.DeviceConfigDAO {
	return registry.deviceConfigDAO
}

func (registry *GormDaoRegistry) SetDeviceConfigDAO(dao dao.DeviceConfigDAO) {
	registry.deviceConfigDAO = dao
}

func (registry *GormDaoRegistry) GetMetricDAO() dao.MetricDAO {
	return registry.metricDAO
}

func (registry *GormDaoRegistry) SetMetricDAO(dao dao.MetricDAO) {
	registry.metricDAO = dao
}

func (registry *GormDaoRegistry) GetChannelDAO() dao.ChannelDAO {
	return registry.channelDAO
}

func (registry *GormDaoRegistry) SetChannelDAO(dao dao.ChannelDAO) {
	registry.channelDAO = dao
}

func (registry *GormDaoRegistry) GetScheduleDAO() dao.ScheduleDAO {
	return registry.scheduleDAO
}

func (registry *GormDaoRegistry) SetScheduleDAO(dao dao.ScheduleDAO) {
	registry.scheduleDAO = dao
}

func (registry *GormDaoRegistry) GetConditionDAO() dao.ConditionDAO {
	return registry.conditionDAO
}

func (registry *GormDaoRegistry) SetConditionDAO(dao dao.ConditionDAO) {
	registry.conditionDAO = dao
}

func (registry *GormDaoRegistry) GetAlgorithmDAO() dao.AlgorithmDAO {
	return registry.algorithmDAO
}

func (registry *GormDaoRegistry) SetAlgorithmDAO(dao dao.AlgorithmDAO) {
	registry.algorithmDAO = dao
}

func (registry *GormDaoRegistry) GetEventLogDAO() EventLogDAO {
	return registry.eventLogDAO
}

func (registry *GormDaoRegistry) SetEventLogDAO(dao EventLogDAO) {
	registry.eventLogDAO = dao
}

func (registry *GormDaoRegistry) GetUserDAO() dao.UserDAO {
	return registry.userDAO
}

func (registry *GormDaoRegistry) SetUserDAO(dao dao.UserDAO) {
	registry.userDAO = dao
}

func (registry *GormDaoRegistry) GetRoleDAO() dao.RoleDAO {
	return registry.roleDAO
}

func (registry *GormDaoRegistry) SetRoleDAO(dao dao.RoleDAO) {
	registry.roleDAO = dao
}
