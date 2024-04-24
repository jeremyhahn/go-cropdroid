package service

import (
	"errors"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/shoppingcart"
)

type DefaultServiceRegistry struct {
	app                   *app.App
	datastoreRegistry     dao.Registry
	mapperRegistry        mapper.MapperRegistry
	algorithmService      AlgorithmService
	authService           AuthService
	changefeedService     ChangefeedService
	channelService        ChannelService
	configService         ConfigService
	conditionService      ConditionService
	deviceFactory         DeviceFactory
	deviceServices        map[uint64][]DeviceService
	deviceServicesMutex   *sync.RWMutex
	eventLogServices      map[uint64]EventLogService
	eventLogServicesMutex *sync.RWMutex
	farmFactory           FarmFactory
	farmServices          map[uint64]FarmService
	farmServicesMutex     *sync.RWMutex
	farmProvisioner       provisioner.FarmProvisioner
	googleAuthService     AuthService
	jwtService            JsonWebTokenService
	metricService         MetricService
	notificationService   NotificationService
	organizationService   OrganizationService
	roleService           RoleService
	scheduleService       ScheduleService
	userService           UserService
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
		farmServices:          make(map[uint64]FarmService, 0),
		deviceServicesMutex:   &sync.RWMutex{},
		deviceServices:        make(map[uint64][]DeviceService, 0),
		eventLogServicesMutex: &sync.RWMutex{},
		eventLogServices:      make(map[uint64]EventLogService, 0)}
}

func CreateServiceRegistry(_app *app.App, daos dao.Registry,
	mappers mapper.MapperRegistry) ServiceRegistry {

	algorithmService := NewAlgorithmService(daos.GetAlgorithmDAO())
	channelService := NewChannelService(daos.GetChannelDAO(), mappers.GetChannelMapper())

	metricService := NewMetricService(daos.GetMetricDAO(), mappers.GetMetricMapper())
	scheduleService := NewScheduleService(_app, daos.GetScheduleDAO(), mappers.GetScheduleMapper(), nil)

	conditionService := NewConditionService(_app.Logger, daos.GetConditionDAO(), mappers.GetConditionMapper())
	workflowService := NewWorkflowService(_app, daos.GetWorkflowDAO(), mappers.GetWorkflowMapper())
	workflowStepService := NewWorkflowStepService(_app, daos.GetWorkflowStepDAO())

	notificationService := NewNotificationService(_app.Logger, nil) // Mailer

	roleService := NewRoleService(_app.Logger, daos.GetRoleDAO())

	authServices := make(map[int]AuthService, 2)
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
		farmServices:          make(map[uint64]FarmService, 0),
		deviceServicesMutex:   &sync.RWMutex{},
		deviceServices:        make(map[uint64][]DeviceService, 0),
		eventLogServicesMutex: &sync.RWMutex{},
		eventLogServices:      make(map[uint64]EventLogService, 0),
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

func (registry *DefaultServiceRegistry) SetAlgorithmService(algoService AlgorithmService) {
	registry.algorithmService = algoService
}

func (registry *DefaultServiceRegistry) GetAlgorithmService() AlgorithmService {
	return registry.algorithmService
}

func (registry *DefaultServiceRegistry) SetAuthService(authService AuthService) {
	registry.authService = authService
}

func (registry *DefaultServiceRegistry) GetAuthService() AuthService {
	return registry.authService
}

func (registry *DefaultServiceRegistry) SetChangefeedService(changefeedService ChangefeedService) {
	registry.changefeedService = changefeedService
}

func (registry *DefaultServiceRegistry) GetChangefeedService() ChangefeedService {
	return registry.changefeedService
}

func (registry *DefaultServiceRegistry) SetChannelService(channelService ChannelService) {
	registry.channelService = channelService
}

func (registry *DefaultServiceRegistry) GetChannelService() ChannelService {
	return registry.channelService
}

func (registry *DefaultServiceRegistry) SetConditionService(conditionService ConditionService) {
	registry.conditionService = conditionService
}

func (registry *DefaultServiceRegistry) GetConditionService() ConditionService {
	return registry.conditionService
}

func (registry *DefaultServiceRegistry) SetConfigService(configService ConfigService) {
	registry.configService = configService
}

func (registry *DefaultServiceRegistry) GetConfigService() ConfigService {
	return registry.configService
}

func (registry *DefaultServiceRegistry) SetDeviceFactory(deviceFactory DeviceFactory) {
	registry.deviceFactory = deviceFactory
}

func (registry *DefaultServiceRegistry) GetDeviceFactory() DeviceFactory {
	return registry.deviceFactory
}

func (registry *DefaultServiceRegistry) SetDeviceServices(farmID uint64, deviceServices []DeviceService) {
	registry.deviceServicesMutex.Lock()
	defer registry.deviceServicesMutex.Unlock()
	registry.deviceServices[farmID] = deviceServices
}

func (registry *DefaultServiceRegistry) GetDeviceServices(farmID uint64) ([]DeviceService, error) {
	registry.deviceServicesMutex.Lock()
	defer registry.deviceServicesMutex.Unlock()
	if services, ok := registry.deviceServices[farmID]; ok {
		return services, nil
	}
	return nil, ErrFarmNotFound
}

func (registry *DefaultServiceRegistry) GetDeviceService(farmID uint64,
	deviceType string) (DeviceService, error) {

	registry.deviceServicesMutex.Lock()
	defer registry.deviceServicesMutex.Unlock()
	if services, ok := registry.deviceServices[farmID]; ok {
		for _, service := range services {
			if service.GetDeviceType() == deviceType {
				return service, nil
			}
		}
		return nil, ErrDeviceNotFound
	}
	return nil, ErrFarmNotFound
}

func (registry *DefaultServiceRegistry) GetDeviceServiceByID(farmID uint64, deviceID uint64) (DeviceService, error) {
	registry.deviceServicesMutex.Lock()
	defer registry.deviceServicesMutex.Unlock()
	if services, ok := registry.deviceServices[farmID]; ok {
		for _, service := range services {
			if service.GetID() == deviceID {
				return service, nil
			}
		}
		return nil, ErrDeviceNotFound
	}
	return nil, ErrFarmNotFound
}

func (registry *DefaultServiceRegistry) AddEventLogService(eventLogService EventLogService) error {
	registry.eventLogServicesMutex.Lock()
	defer registry.eventLogServicesMutex.Unlock()
	if _, ok := registry.eventLogServices[eventLogService.GetFarmID()]; ok {
		return ErrEventLogAlreadyExists
	}
	registry.eventLogServices[eventLogService.GetFarmID()] = eventLogService
	return nil
}

func (registry *DefaultServiceRegistry) SetEventLogService(eventLogServices map[uint64]EventLogService) {
	registry.eventLogServicesMutex.Lock()
	defer registry.eventLogServicesMutex.Unlock()
	registry.eventLogServices = eventLogServices
}

func (registry *DefaultServiceRegistry) GetEventLogServices() map[uint64]EventLogService {
	registry.eventLogServicesMutex.RLock()
	defer registry.eventLogServicesMutex.RUnlock()
	return registry.eventLogServices
}

func (registry *DefaultServiceRegistry) GetEventLogService(farmID uint64) EventLogService {
	registry.eventLogServicesMutex.RLock()
	defer registry.eventLogServicesMutex.RUnlock()
	service, _ := registry.eventLogServices[farmID]
	return service
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

func (registry *DefaultServiceRegistry) AddFarmService(farmService FarmService) error {
	registry.farmServicesMutex.Lock()
	defer registry.farmServicesMutex.Unlock()
	if _, ok := registry.farmServices[farmService.GetFarmID()]; ok {
		return ErrFarmAlreadyExists
	}
	registry.farmServices[farmService.GetFarmID()] = farmService
	return nil
}

func (registry *DefaultServiceRegistry) SetFarmServices(farmServices map[uint64]FarmService) {
	registry.farmServicesMutex.Lock()
	defer registry.farmServicesMutex.Unlock()
	registry.farmServices = farmServices
}

func (registry *DefaultServiceRegistry) GetFarmServices() map[uint64]FarmService {
	registry.farmServicesMutex.RLock()
	defer registry.farmServicesMutex.RUnlock()
	return registry.farmServices
}

func (registry *DefaultServiceRegistry) GetFarmService(farmID uint64) FarmService {
	registry.farmServicesMutex.RLock()
	defer registry.farmServicesMutex.RUnlock()
	service, _ := registry.farmServices[farmID]
	return service
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

func (registry *DefaultServiceRegistry) SetGoogleAuthService(googleAuthService AuthService) {
	registry.googleAuthService = googleAuthService
}

func (registry *DefaultServiceRegistry) GetGoogleAuthService() AuthService {
	return registry.googleAuthService
}

func (registry *DefaultServiceRegistry) SetJsonWebTokenService(jwtService JsonWebTokenService) {
	registry.jwtService = jwtService
}

func (registry *DefaultServiceRegistry) GetJsonWebTokenService() JsonWebTokenService {
	return registry.jwtService
}

func (registry *DefaultServiceRegistry) SetMetricService(metricService MetricService) {
	registry.metricService = metricService
}

func (registry *DefaultServiceRegistry) GetMetricService() MetricService {
	return registry.metricService
}

func (registry *DefaultServiceRegistry) SetNotificationService(notificationService NotificationService) {
	registry.notificationService = notificationService
}

func (registry *DefaultServiceRegistry) GetNotificationService() NotificationService {
	return registry.notificationService
}

func (registry *DefaultServiceRegistry) SetScheduleService(scheduleService ScheduleService) {
	registry.scheduleService = scheduleService
}

func (registry *DefaultServiceRegistry) GetScheduleService() ScheduleService {
	return registry.scheduleService
}

func (registry *DefaultServiceRegistry) SetUserService(userService UserService) {
	registry.userService = userService
}

func (registry *DefaultServiceRegistry) GetUserService() UserService {
	return registry.userService
}

func (registry *DefaultServiceRegistry) SetOrganizationService(organizationService OrganizationService) {
	registry.organizationService = organizationService
}

func (registry *DefaultServiceRegistry) GetOrganizationService() OrganizationService {
	return registry.organizationService
}

func (registry *DefaultServiceRegistry) SetRoleService(roleService RoleService) {
	registry.roleService = roleService
}

func (registry *DefaultServiceRegistry) GetRoleService() RoleService {
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
