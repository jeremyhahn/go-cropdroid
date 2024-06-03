package service

import (
	"errors"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/shoppingcart"
)

type ServiceRegistry interface {
	SetAlgorithmService(AlgorithmServicer)
	GetAlgorithmService() AlgorithmServicer
	SetAuthService(AuthServicer)
	GetAuthService() AuthServicer
	SetChannelService(ChannelServicer)
	GetChannelService() ChannelServicer
	SetConditionService(ConditionServicer)
	GetConditionService() ConditionServicer
	SetDeviceFactory(DeviceFactory)
	GetDeviceFactory() DeviceFactory
	SetDeviceServices(farmID uint64, deviceServices []DeviceServicer)
	GetDeviceServices(farmID uint64) ([]DeviceServicer, error)
	GetDeviceService(farmID uint64, deviceType string) (DeviceServicer, error)
	GetDeviceServiceByID(farmID uint64, deviceID uint64) (DeviceServicer, error)
	SetDeviceService(farmID uint64, deviceService DeviceServicer)
	AddEventLogService(eventLogService EventLogServicer) error
	SetEventLogService(eventLogServices map[uint64]EventLogServicer)
	GetEventLogServices() map[uint64]EventLogServicer
	GetEventLogService(farmID uint64) EventLogServicer
	RemoveEventLogService(farmID uint64)
	SetFarmFactory(FarmFactory FarmFactory)
	GetFarmFactory() FarmFactory
	AddFarmService(farmService FarmServicer) error
	SetFarmServices(map[uint64]FarmServicer)
	GetFarmServices() map[uint64]FarmServicer
	GetFarmService(uint64) FarmServicer
	RemoveFarmService(farmID uint64)
	SetFarmProvisioner(farmProvisioner provisioner.FarmProvisioner)
	GetFarmProvisioner() provisioner.FarmProvisioner
	SetGoogleAuthService(googleAuthService AuthServicer)
	GetGoogleAuthService() AuthServicer
	SetMetricService(MetricService)
	GetMetricService() MetricService
	SetNotificationService(NotificationServicer)
	GetNotificationService() NotificationServicer
	SetScheduleService(ScheduleService)
	GetScheduleService() ScheduleService
	SetShoppingCartService(shoppingcart.ShoppingCartService)
	GetShoppingCartService() shoppingcart.ShoppingCartService
	SetOrganizationService(organizationService OrganizationService)
	GetOrganizationService() OrganizationService
	SetRoleService(roleService RoleServicer)
	GetRoleService() RoleServicer
	SetUserService(UserServicer)
	GetUserService() UserServicer
	SetWorkflowService(WorkflowService)
	GetWorkflowService() WorkflowService
	SetWorkflowStepService(WorkflowStepService)
	GetWorkflowStepService() WorkflowStepService
}

type DefaultServiceRegistry struct {
	app                   *app.App
	algorithmService      AlgorithmServicer
	authService           AuthServicer
	channelService        ChannelServicer
	conditionService      ConditionServicer
	deviceFactory         DeviceFactory
	deviceServices        map[uint64][]DeviceServicer
	deviceServicesMutex   *sync.RWMutex
	eventLogServices      map[uint64]EventLogServicer
	eventLogServicesMutex *sync.RWMutex
	farmFactory           FarmFactory
	farmServices          map[uint64]FarmServicer
	farmServicesMutex     *sync.RWMutex
	farmProvisioner       provisioner.FarmProvisioner
	googleAuthService     AuthServicer
	metricService         MetricService
	notificationService   NotificationServicer
	organizationService   OrganizationService
	roleService           RoleServicer
	scheduleService       ScheduleService
	userService           UserServicer
	workflowService       WorkflowService
	workflowStepService   WorkflowStepService
	shoppingCartService   shoppingcart.ShoppingCartService
	ServiceRegistry
}

var (
	ErrFarmAlreadyExists     = errors.New("farm already exists")
	ErrFarmNotFound          = errors.New("farm not found")
	ErrDeviceAlreadyExists   = errors.New("device already exists")
	ErrEventLogAlreadyExists = errors.New("event log already exists")
)

func NewServiceRegistry(app *app.App) ServiceRegistry {
	return &DefaultServiceRegistry{
		app:                   app,
		farmServicesMutex:     &sync.RWMutex{},
		farmServices:          make(map[uint64]FarmServicer, 0),
		deviceServicesMutex:   &sync.RWMutex{},
		deviceServices:        make(map[uint64][]DeviceServicer, 0),
		eventLogServicesMutex: &sync.RWMutex{},
		eventLogServices:      make(map[uint64]EventLogServicer, 0)}
}

func CreateServiceRegistry(_app *app.App, daos dao.Registry,
	mappers mapper.MapperRegistry) ServiceRegistry {

	algorithmService := NewAlgorithmService(daos.GetAlgorithmDAO())
	channelService := NewChannelService(daos.GetChannelDAO(), mappers.GetChannelMapper())

	metricService := NewMetricService(daos.GetMetricDAO(), mappers.GetMetricMapper())
	scheduleService := NewScheduleService(_app, daos.GetScheduleDAO())

	conditionService := NewConditionService(_app.Logger, daos.GetConditionDAO(), mappers.GetConditionMapper())
	workflowService := NewWorkflowService(_app, daos.GetWorkflowDAO(), mappers.GetWorkflowMapper())
	workflowStepService := NewWorkflowStepService(_app, daos.GetWorkflowStepDAO())

	notificationService := NewNotificationService(_app.Logger, nil) // Mailer

	roleService := NewRoleService(_app.Logger, daos.GetRoleDAO())

	authServices := make(map[int]AuthServicer, 2)
	authService := NewLocalAuthService(_app, daos.GetPermissionDAO(),
		daos.GetRegistrationDAO(), daos.GetOrganizationDAO(),
		daos.GetFarmDAO(), daos.GetUserDAO(), daos.GetRoleDAO(),
		mappers.GetUserMapper())
	gas := NewGoogleAuthService(_app, daos.GetPermissionDAO(),
		daos.GetUserDAO(), daos.GetRoleDAO(), daos.GetFarmDAO(),
		mappers.GetUserMapper())
	authServices[common.AUTH_TYPE_LOCAL] = authService
	authServices[common.AUTH_TYPE_GOOGLE] = gas

	shoppingCartService := shoppingcart.NewStripeService(_app, daos.GetCustomerDAO())

	registry := &DefaultServiceRegistry{
		app:                   _app,
		algorithmService:      algorithmService,
		authService:           authService,
		googleAuthService:     gas,
		channelService:        channelService,
		conditionService:      conditionService,
		farmServicesMutex:     &sync.RWMutex{},
		farmServices:          make(map[uint64]FarmServicer, 0),
		deviceServicesMutex:   &sync.RWMutex{},
		deviceServices:        make(map[uint64][]DeviceServicer, 0),
		eventLogServicesMutex: &sync.RWMutex{},
		eventLogServices:      make(map[uint64]EventLogServicer, 0),
		metricService:         metricService,
		notificationService:   notificationService,
		scheduleService:       scheduleService,
		shoppingCartService:   shoppingCartService,
		roleService:           roleService,
		workflowService:       workflowService,
		workflowStepService:   workflowStepService}

	registry.SetUserService(NewUserService(_app, daos.GetUserDAO(), daos.GetOrganizationDAO(),
		daos.GetRoleDAO(), daos.GetPermissionDAO(), daos.GetFarmDAO(),
		mappers.GetUserMapper(), authServices, registry))

	return registry
}

func (registry *DefaultServiceRegistry) SetAlgorithmService(algoService AlgorithmServicer) {
	registry.algorithmService = algoService
}

func (registry *DefaultServiceRegistry) GetAlgorithmService() AlgorithmServicer {
	return registry.algorithmService
}

func (registry *DefaultServiceRegistry) SetAuthService(authService AuthServicer) {
	registry.authService = authService
}

func (registry *DefaultServiceRegistry) GetAuthService() AuthServicer {
	return registry.authService
}

func (registry *DefaultServiceRegistry) SetChannelService(channelService ChannelServicer) {
	registry.channelService = channelService
}

func (registry *DefaultServiceRegistry) GetChannelService() ChannelServicer {
	return registry.channelService
}

func (registry *DefaultServiceRegistry) SetConditionService(conditionService ConditionServicer) {
	registry.conditionService = conditionService
}

func (registry *DefaultServiceRegistry) GetConditionService() ConditionServicer {
	return registry.conditionService
}

func (registry *DefaultServiceRegistry) SetDeviceFactory(deviceFactory DeviceFactory) {
	registry.deviceFactory = deviceFactory
}

func (registry *DefaultServiceRegistry) GetDeviceFactory() DeviceFactory {
	return registry.deviceFactory
}

func (registry *DefaultServiceRegistry) SetDeviceServices(farmID uint64, deviceServices []DeviceServicer) {
	registry.deviceServicesMutex.Lock()
	defer registry.deviceServicesMutex.Unlock()
	registry.deviceServices[farmID] = deviceServices
}

func (registry *DefaultServiceRegistry) GetDeviceServices(farmID uint64) ([]DeviceServicer, error) {
	registry.deviceServicesMutex.Lock()
	defer registry.deviceServicesMutex.Unlock()
	if services, ok := registry.deviceServices[farmID]; ok {
		return services, nil
	}
	return nil, ErrFarmNotFound
}

func (registry *DefaultServiceRegistry) GetDeviceService(farmID uint64,
	deviceType string) (DeviceServicer, error) {

	registry.deviceServicesMutex.Lock()
	defer registry.deviceServicesMutex.Unlock()
	if services, ok := registry.deviceServices[farmID]; ok {
		for _, deviceService := range services {
			if deviceService.DeviceType() == deviceType {
				return deviceService, nil
			}
		}
		return nil, ErrDeviceNotFound
	}
	return nil, ErrFarmNotFound
}

func (registry *DefaultServiceRegistry) GetDeviceServiceByID(farmID uint64, deviceID uint64) (DeviceServicer, error) {
	registry.deviceServicesMutex.Lock()
	defer registry.deviceServicesMutex.Unlock()
	if services, ok := registry.deviceServices[farmID]; ok {
		for _, service := range services {
			if service.ID() == deviceID {
				return service, nil
			}
		}
		return nil, ErrDeviceNotFound
	}
	return nil, ErrFarmNotFound
}

func (registry *DefaultServiceRegistry) SetDeviceService(farmID uint64, deviceService DeviceServicer) {
	registry.deviceServicesMutex.Lock()
	defer registry.deviceServicesMutex.Unlock()
	registry.deviceServices[farmID] = append(registry.deviceServices[farmID], deviceService)
}

func (registry *DefaultServiceRegistry) AddEventLogService(eventLogService EventLogServicer) error {
	registry.eventLogServicesMutex.Lock()
	defer registry.eventLogServicesMutex.Unlock()
	if _, ok := registry.eventLogServices[eventLogService.GetFarmID()]; ok {
		return ErrEventLogAlreadyExists
	}
	registry.eventLogServices[eventLogService.GetFarmID()] = eventLogService
	return nil
}

func (registry *DefaultServiceRegistry) SetEventLogService(eventLogServices map[uint64]EventLogServicer) {
	registry.eventLogServicesMutex.Lock()
	defer registry.eventLogServicesMutex.Unlock()
	registry.eventLogServices = eventLogServices
}

func (registry *DefaultServiceRegistry) GetEventLogServices() map[uint64]EventLogServicer {
	registry.eventLogServicesMutex.RLock()
	defer registry.eventLogServicesMutex.RUnlock()
	return registry.eventLogServices
}

func (registry *DefaultServiceRegistry) GetEventLogService(farmID uint64) EventLogServicer {
	registry.eventLogServicesMutex.RLock()
	defer registry.eventLogServicesMutex.RUnlock()
	return registry.eventLogServices[farmID]
}

func (registry *DefaultServiceRegistry) RemoveEventLogService(farmID uint64) {
	registry.eventLogServicesMutex.Lock()
	defer registry.eventLogServicesMutex.Unlock()
	delete(registry.eventLogServices, farmID)
}

func (registry *DefaultServiceRegistry) SetFarmFactory(farmFactory FarmFactory) {
	registry.farmFactory = farmFactory
}

func (registry *DefaultServiceRegistry) GetFarmFactory() FarmFactory {
	return registry.farmFactory
}

func (registry *DefaultServiceRegistry) AddFarmService(farmService FarmServicer) error {
	registry.farmServicesMutex.Lock()
	defer registry.farmServicesMutex.Unlock()
	if _, ok := registry.farmServices[farmService.GetFarmID()]; ok {
		return ErrFarmAlreadyExists
	}
	registry.farmServices[farmService.GetFarmID()] = farmService
	return nil
}

func (registry *DefaultServiceRegistry) SetFarmServices(farmServices map[uint64]FarmServicer) {
	registry.farmServicesMutex.Lock()
	defer registry.farmServicesMutex.Unlock()
	registry.farmServices = farmServices
}

func (registry *DefaultServiceRegistry) GetFarmServices() map[uint64]FarmServicer {
	registry.farmServicesMutex.RLock()
	defer registry.farmServicesMutex.RUnlock()
	return registry.farmServices
}

func (registry *DefaultServiceRegistry) GetFarmService(farmID uint64) FarmServicer {
	registry.farmServicesMutex.RLock()
	defer registry.farmServicesMutex.RUnlock()
	return registry.farmServices[farmID]
}

func (registry *DefaultServiceRegistry) RemoveFarmService(farmID uint64) {
	registry.farmServicesMutex.Lock()
	defer registry.farmServicesMutex.Unlock()
	delete(registry.farmServices, farmID)
}

func (registry *DefaultServiceRegistry) SetFarmProvisioner(farmProvisioner provisioner.FarmProvisioner) {
	registry.farmProvisioner = farmProvisioner
}

func (registry *DefaultServiceRegistry) GetFarmProvisioner() provisioner.FarmProvisioner {
	return registry.farmProvisioner
}

func (registry *DefaultServiceRegistry) SetGoogleAuthService(googleAuthService AuthServicer) {
	registry.googleAuthService = googleAuthService
}

func (registry *DefaultServiceRegistry) GetGoogleAuthService() AuthServicer {
	return registry.googleAuthService
}

func (registry *DefaultServiceRegistry) SetMetricService(metricService MetricService) {
	registry.metricService = metricService
}

func (registry *DefaultServiceRegistry) GetMetricService() MetricService {
	return registry.metricService
}

func (registry *DefaultServiceRegistry) SetNotificationService(notificationService NotificationServicer) {
	registry.notificationService = notificationService
}

func (registry *DefaultServiceRegistry) GetNotificationService() NotificationServicer {
	return registry.notificationService
}

func (registry *DefaultServiceRegistry) SetScheduleService(scheduleService ScheduleService) {
	registry.scheduleService = scheduleService
}

func (registry *DefaultServiceRegistry) GetScheduleService() ScheduleService {
	return registry.scheduleService
}

func (registry *DefaultServiceRegistry) SetUserService(userService UserServicer) {
	registry.userService = userService
}

func (registry *DefaultServiceRegistry) GetUserService() UserServicer {
	return registry.userService
}

func (registry *DefaultServiceRegistry) SetOrganizationService(organizationService OrganizationService) {
	registry.organizationService = organizationService
}

func (registry *DefaultServiceRegistry) GetOrganizationService() OrganizationService {
	return registry.organizationService
}

func (registry *DefaultServiceRegistry) SetRoleService(roleService RoleServicer) {
	registry.roleService = roleService
}

func (registry *DefaultServiceRegistry) GetRoleService() RoleServicer {
	return registry.roleService
}

func (registry *DefaultServiceRegistry) SetWorkflowService(workflowService WorkflowService) {
	registry.workflowService = workflowService
}

func (registry *DefaultServiceRegistry) GetWorkflowService() WorkflowService {
	return registry.workflowService
}

func (registry *DefaultServiceRegistry) SetWorkflowStepService(workflowStepService WorkflowStepService) {
	registry.workflowStepService = workflowStepService
}

func (registry *DefaultServiceRegistry) GetWorkflowStepService() WorkflowStepService {
	return registry.workflowStepService
}

func (registry *DefaultServiceRegistry) SetShoppingCartService(shoppingCartService shoppingcart.ShoppingCartService) {
	registry.shoppingCartService = shoppingCartService
}

func (registry *DefaultServiceRegistry) GetShoppingCartService() shoppingcart.ShoppingCartService {
	return registry.shoppingCartService
}
