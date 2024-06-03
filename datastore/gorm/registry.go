package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"
)

type GormDaoRegistry struct {
	logger          *logging.Logger
	gormDB          GormDB
	idGenerator     util.IdGenerator
	permissionDAO   dao.PermissionDAO
	registrationDAO dao.RegistrationDAO
	orgDAO          dao.OrganizationDAO
	farmDAO         dao.FarmDAO
	deviceDAO       dao.DeviceDAO
	deviceConfigDAO dao.DeviceSettingDAO
	metricDAO       dao.MetricDAO
	channelDAO      dao.ChannelDAO
	scheduleDAO     dao.ScheduleDAO
	conditionDAO    dao.ConditionDAO
	algorithmDAO    dao.AlgorithmDAO
	eventLogDAO     dao.EventLogDAO
	userDAO         dao.UserDAO
	roleDAO         dao.RoleDAO
	customerDAO     dao.CustomerDAO
	workflowDAO     dao.WorkflowDAO
	workflowStepDAO dao.WorkflowStepDAO
	dao.Registry
}

func NewGormRegistry(logger *logging.Logger, gormDB GormDB) dao.Registry {
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_32BIT)
	return &GormDaoRegistry{
		logger:          logger,
		gormDB:          gormDB,
		idGenerator:     idGenerator,
		permissionDAO:   NewPermissionDAO(logger, gormDB.CloneConnection()),
		registrationDAO: NewRegistrationDAO(logger, gormDB.CloneConnection()),
		orgDAO:          NewOrganizationDAO(logger, gormDB.CloneConnection(), idGenerator),
		farmDAO:         NewFarmDAO(logger, gormDB.CloneConnection(), idGenerator),
		deviceDAO:       NewDeviceDAO(logger, gormDB.CloneConnection()),
		deviceConfigDAO: NewDeviceSettingDAO(logger, gormDB.CloneConnection()),
		metricDAO:       NewMetricDAO(logger, gormDB.CloneConnection()),
		channelDAO:      NewChannelDAO(logger, gormDB.CloneConnection()),
		scheduleDAO:     NewScheduleDAO(logger, gormDB.CloneConnection()),
		conditionDAO:    NewConditionDAO(logger, gormDB.CloneConnection()),
		algorithmDAO:    NewGenericGormDAO[*config.AlgorithmStruct](logger, gormDB.CloneConnection()),
		eventLogDAO:     NewEventLogDAO(logger, gormDB.CloneConnection(), 0),
		userDAO:         NewUserDAO(logger, gormDB.CloneConnection()),
		roleDAO:         NewRoleDAO(logger, gormDB.CloneConnection()),
		customerDAO:     NewCustomerDAO(logger, gormDB.CloneConnection()),
		workflowDAO:     NewWorkflowDAO(logger, gormDB.CloneConnection()),
		workflowStepDAO: NewWorkflowStepDAO(logger, gormDB.CloneConnection())}
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

// Retuns a FarmDAO with a new connection to the database
func (registry *GormDaoRegistry) NewFarmDAO() dao.FarmDAO {
	db := registry.gormDB.CloneConnection()
	return NewFarmDAO(registry.logger, db, registry.idGenerator)
}

// Gets the global DeviceDAO from the registry
func (registry *GormDaoRegistry) GetDeviceDAO() dao.DeviceDAO {
	return registry.deviceDAO
}

// Retuns a DeviceDAO with a new connection to the database
func (registry *GormDaoRegistry) NewDeviceDAO() dao.DeviceDAO {
	return NewDeviceDAO(registry.logger, registry.gormDB.CloneConnection())
}

func (registry *GormDaoRegistry) SetDeviceDAO(dao dao.DeviceDAO) {
	registry.deviceDAO = dao
}

func (registry *GormDaoRegistry) GetDeviceConfigDAO() dao.DeviceSettingDAO {
	return registry.deviceConfigDAO
}

func (registry *GormDaoRegistry) SetDeviceConfigDAO(dao dao.DeviceSettingDAO) {
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

func (registry *GormDaoRegistry) GetEventLogDAO() dao.EventLogDAO {
	return registry.eventLogDAO
}

func (registry *GormDaoRegistry) SetEventLogDAO(dao dao.EventLogDAO) {
	registry.eventLogDAO = dao
}

func (registry *GormDaoRegistry) GetUserDAO() dao.UserDAO {
	return registry.userDAO
}

func (registry *GormDaoRegistry) SetUserDAO(dao dao.UserDAO) {
	registry.userDAO = dao
}

func (registry *GormDaoRegistry) GetRegistrationDAO() dao.RegistrationDAO {
	return registry.registrationDAO
}

func (registry *GormDaoRegistry) SetRegistrationDAO(regDAO dao.RegistrationDAO) {
	registry.registrationDAO = regDAO
}

func (registry *GormDaoRegistry) GetPermissionDAO() dao.PermissionDAO {
	return registry.permissionDAO
}

func (registry *GormDaoRegistry) SetPermissionDAO(permissionDAO dao.PermissionDAO) {
	registry.permissionDAO = permissionDAO
}

func (registry *GormDaoRegistry) GetRoleDAO() dao.RoleDAO {
	return registry.roleDAO
}

func (registry *GormDaoRegistry) SetRoleDAO(dao dao.RoleDAO) {
	registry.roleDAO = dao
}

func (registry *GormDaoRegistry) GetCustomerDAO() dao.CustomerDAO {
	return registry.customerDAO
}

func (registry *GormDaoRegistry) SetCustomerDAO(dao dao.CustomerDAO) {
	registry.customerDAO = dao
}

func (registry *GormDaoRegistry) GetWorkflowDAO() dao.WorkflowDAO {
	return registry.workflowDAO
}

func (registry *GormDaoRegistry) SetWorkflowDAO(dao dao.WorkflowDAO) {
	registry.workflowDAO = dao
}

func (registry *GormDaoRegistry) GetWorkflowStepDAO() dao.WorkflowStepDAO {
	return registry.workflowStepDAO
}

func (registry *GormDaoRegistry) SetWorkflowStepDAO(dao dao.WorkflowStepDAO) {
	registry.workflowStepDAO = dao
}
