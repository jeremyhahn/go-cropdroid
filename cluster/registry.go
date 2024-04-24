//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"
)

type RaftDaoRegistry struct {
	logger           *logging.Logger
	idGenerator      util.IdGenerator
	raftNode         RaftNode
	serverDAO        ServerDAO
	permissionDAO    dao.PermissionDAO
	registrationDAO  dao.RegistrationDAO
	orgDAO           dao.OrganizationDAO
	farmDAO          dao.FarmDAO
	deviceDAO        dao.DeviceDAO
	deviceSettingDAO dao.DeviceSettingDAO
	metricDAO        dao.MetricDAO
	channelDAO       dao.ChannelDAO
	scheduleDAO      dao.ScheduleDAO
	conditionDAO     dao.ConditionDAO
	algorithmDAO     dao.AlgorithmDAO
	eventLogDAO      dao.EventLogDAO
	userDAO          dao.UserDAO
	roleDAO          dao.RoleDAO
	customerDAO      dao.CustomerDAO
	workflowDAO      dao.WorkflowDAO
	workflowStepDAO  dao.WorkflowStepDAO
	dao.Registry
}

func NewRaftRegistry(logger *logging.Logger,
	idGenerator util.IdGenerator,
	raftNode RaftNode) dao.Registry {

	raftOptions := raftNode.GetParams().RaftOptions

	serverDAO := NewRaftServerDAO(logger,
		raftNode, raftOptions.SystemClusterID)
	// This Raft is automatically started in the NewRaftNode constructor
	// as soon as there are enough nodes to form the quorum, as Gossip.Join()
	// adds nodes to the Raft cluster.
	//serverDAO.(*RaftServerDAO).StartCluster()

	eventLogDAO := NewRaftEventLogDAO(logger,
		raftNode, raftOptions.SystemClusterID)
	eventLogDAO.(*RaftEventLogDAO).StartCluster()

	orgDAO := NewRaftOrganizationDAO(logger,
		raftNode, raftOptions.OrganizationClusterID, serverDAO)
	orgDAO.(*RaftOrganizationDAO).StartCluster()

	userDAO := NewRaftUserDAO(logger,
		raftNode, raftOptions.UserClusterID)
	userDAO.(*RaftUserDAO).StartCluster()

	roleDAO := NewRaftRoleDAO(logger,
		raftNode, raftOptions.RoleClusterID)
	roleDAO.(*RaftRoleDAO).StartCluster()

	customerDAO := NewRaftCustomerDAO(logger,
		raftNode, raftOptions.CustomerClusterID)
	customerDAO.(*RaftCustomerDAO).StartCluster()

	farmDAO := NewRaftFarmConfigDAO(logger,
		raftNode, serverDAO, userDAO)

	deviceDAO := NewRaftDeviceConfigDAO(logger,
		raftNode, farmDAO)

	deviceSettingsDAO := NewRaftDeviceSettingDAO(logger,
		raftNode, deviceDAO)

	metricDAO := NewRaftMetricDAO(logger,
		raftNode, farmDAO)

	channelDAO := NewRaftChannelDAO(logger,
		raftNode, farmDAO)

	conditionDAO := NewRaftConditionDAO(logger,
		raftNode, farmDAO)

	scheduleDAO := NewRaftScheduleDAO(logger,
		raftNode, farmDAO)

	algorithmDAO := NewRaftAlgorithmDAO(logger,
		raftNode, raftOptions.AlgorithmClusterID)
	algorithmDAO.(*RaftAlgorithmDAO).StartCluster()

	workflowDAO := NewRaftWorkflowDAO(logger,
		raftNode, farmDAO)

	workflowStepDAO := NewRaftWorkflowStepDAO(logger,
		raftNode, farmDAO)

	registrationDAO := NewRaftRegistrationDAO(logger,
		raftNode, raftOptions.RegistrationClusterID)
	registrationDAO.(*RaftRegistrationDAO).StartCluster()

	permissionDAO := NewRaftPermissionDAO(logger,
		orgDAO, farmDAO, userDAO)

	// Wait for clusters to become ready
	eventLogClusterID := raftNode.GetParams().
		IdGenerator.CreateEventLogClusterID(raftOptions.SystemClusterID)
	raftNode.WaitForClusterReady(eventLogClusterID)

	raftNode.WaitForClusterReady(raftOptions.OrganizationClusterID)
	raftNode.WaitForClusterReady(raftOptions.RoleClusterID)
	raftNode.WaitForClusterReady(raftOptions.UserClusterID)

	registry := &RaftDaoRegistry{
		logger:           logger,
		idGenerator:      idGenerator,
		raftNode:         raftNode,
		serverDAO:        serverDAO,
		registrationDAO:  registrationDAO,
		orgDAO:           orgDAO,
		farmDAO:          farmDAO,
		deviceDAO:        deviceDAO,
		deviceSettingDAO: deviceSettingsDAO,
		metricDAO:        metricDAO,
		channelDAO:       channelDAO,
		scheduleDAO:      scheduleDAO,
		conditionDAO:     conditionDAO,
		algorithmDAO:     algorithmDAO,
		eventLogDAO:      eventLogDAO,
		userDAO:          userDAO,
		roleDAO:          roleDAO,
		customerDAO:      customerDAO,
		workflowDAO:      workflowDAO,
		workflowStepDAO:  workflowStepDAO,
		permissionDAO:    permissionDAO}

	return registry
}

// func (registry *RaftDaoRegistry) GetOrganizationDAO() dao.OrganizationDAO {
// 	return registry.orgDAO
// }

// func (registry *RaftDaoRegistry) SetOrganizationDAO(dao dao.OrganizationDAO) {
// 	registry.orgDAO = dao
// }

func (registry *RaftDaoRegistry) GetFarmDAO() dao.FarmDAO {
	return registry.farmDAO
}

func (registry *RaftDaoRegistry) SetFarmDAO(dao dao.FarmDAO) {
	registry.farmDAO = dao
}

// Retuns a FarmDAO with a new connection to the database
func (registry *RaftDaoRegistry) NewFarmDAO() dao.FarmDAO {
	farmDAO := NewRaftFarmConfigDAO(registry.logger,
		registry.raftNode, registry.serverDAO, registry.userDAO)
	return farmDAO
}

// Gets the global DeviceDAO from the registry
func (registry *RaftDaoRegistry) GetDeviceDAO() dao.DeviceDAO {
	return registry.deviceDAO
}

// Retuns a DeviceDAO with a new connection to the database
func (registry *RaftDaoRegistry) NewDeviceDAO() dao.DeviceDAO {
	farmDAO := registry.NewFarmDAO()
	return NewRaftDeviceConfigDAO(registry.logger,
		registry.raftNode, farmDAO)
}

func (registry *RaftDaoRegistry) SetDeviceDAO(dao dao.DeviceDAO) {
	registry.deviceDAO = dao
}

func (registry *RaftDaoRegistry) GetDeviceSettingDAO() dao.DeviceSettingDAO {
	return registry.deviceSettingDAO
}

func (registry *RaftDaoRegistry) SetDeviceSettingDAO(dao dao.DeviceSettingDAO) {
	registry.deviceSettingDAO = dao
}

func (registry *RaftDaoRegistry) GetMetricDAO() dao.MetricDAO {
	return registry.metricDAO
}

func (registry *RaftDaoRegistry) SetMetricDAO(dao dao.MetricDAO) {
	registry.metricDAO = dao
}

func (registry *RaftDaoRegistry) GetChannelDAO() dao.ChannelDAO {
	return registry.channelDAO
}

func (registry *RaftDaoRegistry) SetChannelDAO(dao dao.ChannelDAO) {
	registry.channelDAO = dao
}

func (registry *RaftDaoRegistry) GetScheduleDAO() dao.ScheduleDAO {
	return registry.scheduleDAO
}

func (registry *RaftDaoRegistry) SetScheduleDAO(dao dao.ScheduleDAO) {
	registry.scheduleDAO = dao
}

func (registry *RaftDaoRegistry) GetConditionDAO() dao.ConditionDAO {
	return registry.conditionDAO
}

func (registry *RaftDaoRegistry) SetConditionDAO(dao dao.ConditionDAO) {
	registry.conditionDAO = dao
}

func (registry *RaftDaoRegistry) GetAlgorithmDAO() dao.AlgorithmDAO {
	return registry.algorithmDAO
}

func (registry *RaftDaoRegistry) SetAlgorithmDAO(dao dao.AlgorithmDAO) {
	registry.algorithmDAO = dao
}

func (registry *RaftDaoRegistry) GetEventLogDAO() dao.EventLogDAO {
	return registry.eventLogDAO
}

func (registry *RaftDaoRegistry) SetEventLogDAO(dao dao.EventLogDAO) {
	registry.eventLogDAO = dao
}

func (registry *RaftDaoRegistry) GetUserDAO() dao.UserDAO {
	return registry.userDAO
}

func (registry *RaftDaoRegistry) SetUserDAO(dao dao.UserDAO) {
	registry.userDAO = dao
}

func (registry *RaftDaoRegistry) GetRegistrationDAO() dao.RegistrationDAO {
	return registry.registrationDAO
}

func (registry *RaftDaoRegistry) SetRegistrationDAO(regDAO dao.RegistrationDAO) {
	registry.registrationDAO = regDAO
}

func (registry *RaftDaoRegistry) GetPermissionDAO() dao.PermissionDAO {
	return registry.permissionDAO
}

func (registry *RaftDaoRegistry) SetPermissionDAO(permissionDAO dao.PermissionDAO) {
	registry.permissionDAO = permissionDAO
}

func (registry *RaftDaoRegistry) GetRoleDAO() dao.RoleDAO {
	return registry.roleDAO
}

func (registry *RaftDaoRegistry) SetRoleDAO(dao dao.RoleDAO) {
	registry.roleDAO = dao
}

func (registry *RaftDaoRegistry) GetCustomerDAO() dao.CustomerDAO {
	return registry.customerDAO
}

func (registry *RaftDaoRegistry) SetCustomerDAO(dao dao.CustomerDAO) {
	registry.customerDAO = dao
}

func (registry *RaftDaoRegistry) GetWorkflowDAO() dao.WorkflowDAO {
	return registry.workflowDAO
}

func (registry *RaftDaoRegistry) SetWorkflowDAO(dao dao.WorkflowDAO) {
	registry.workflowDAO = dao
}

func (registry *RaftDaoRegistry) GetWorkflowStepDAO() dao.WorkflowStepDAO {
	return registry.workflowStepDAO
}

func (registry *RaftDaoRegistry) SetWorkflowStepDAO(dao dao.WorkflowStepDAO) {
	registry.workflowStepDAO = dao
}

func (registry *RaftDaoRegistry) GetOrganizationDAO() dao.OrganizationDAO {
	return registry.orgDAO
}

func (registry *RaftDaoRegistry) GetServerDAO() ServerDAO {
	return registry.serverDAO
}
