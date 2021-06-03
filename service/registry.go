package service

import (
	"errors"
	"sync"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/datastore"
	"github.com/jeremyhahn/cropdroid/mapper"
)

type DefaultServiceRegistry struct {
	datastoreRegistry   datastore.DatastoreRegistry
	mapperRegistry      mapper.MapperRegistry
	algorithmService    AlgorithmService
	authService         AuthService
	changefeedService   ChangefeedService
	channelService      ChannelService
	configService       ConfigService
	conditionService    ConditionService
	controllerFactory   ControllerFactory
	controllerServices  map[int][]common.ControllerService
	eventLogService     EventLogService
	farmFactory         *FarmFactory
	farmServices        map[int]FarmService
	farmServicesMutex   *sync.RWMutex
	googleAuthService   AuthService
	jwtService          JsonWebTokenService
	mailer              common.Mailer
	metricService       MetricService
	notificationService NotificationService
	scheduleService     ScheduleService
	userService         UserService
}

var (
	ErrFarmAlreadyExists = errors.New("Farm already exists")
	ErrFarmNotFound      = errors.New("Farm not found")
)

func NewServiceRegistry() ServiceRegistry {
	return &DefaultServiceRegistry{}
}

func CreateServiceRegistry(_app *app.App, daos datastore.DatastoreRegistry, mappers mapper.MapperRegistry) ServiceRegistry {

	algorithmService := NewAlgorithmService(daos.GetAlgorithmDAO())
	authService := NewLocalAuthService(_app, daos.GetUserDAO(), daos.GetOrganizationDAO(), daos.GetFarmDAO(), mappers.GetUserMapper())
	channelService := NewChannelService(daos.GetChannelDAO(), mappers.GetChannelMapper(), nil) // ConfigService
	//configService
	//eventLogService := NewEventLogService(app, daos.GetEventLogDAO(), common.CONTROLLER_TYPE_SERVER)
	metricService := NewMetricService(daos.GetMetricDAO(), mappers.GetMetricMapper(), nil)               // ConfigService
	scheduleService := NewScheduleService(_app, daos.GetScheduleDAO(), mappers.GetScheduleMapper(), nil) // ConfigService

	conditionService := NewConditionService(_app.Logger, daos.GetConditionDAO(), mappers.GetConditionMapper(), nil) // ConfigService

	//serviceRegistry.SetMailer(NewMailer(farm.logger, farm.buildSmtp()))
	notificationService := NewNotificationService(_app.Logger, nil) // Mailer

	controllerFactory := NewControllerFactory(daos.GetControllerDAO(), mappers.GetControllerMapper())

	authServices := make(map[int]AuthService, 1)
	authServices[common.AUTH_TYPE_LOCAL] = authService

	registry := &DefaultServiceRegistry{
		algorithmService:    algorithmService,
		authService:         authService,
		channelService:      channelService,
		conditionService:    conditionService,
		controllerFactory:   controllerFactory,
		controllerServices:  make(map[int][]common.ControllerService, 0),
		farmServicesMutex:   &sync.RWMutex{},
		farmServices:        make(map[int]FarmService, 0),
		metricService:       metricService,
		notificationService: notificationService,
		scheduleService:     scheduleService}

	registry.SetUserService(NewUserService(_app, daos.GetUserDAO(), daos.GetOrganizationDAO(), daos.GetRoleDAO(),
		daos.GetFarmDAO(), mappers.GetUserMapper(), authServices, registry))

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

func (registry *DefaultServiceRegistry) SetControllerFactory(controllerFactory ControllerFactory) {
	registry.controllerFactory = controllerFactory
}

func (registry *DefaultServiceRegistry) GetControllerFactory() ControllerFactory {
	return registry.controllerFactory
}

func (registry *DefaultServiceRegistry) SetControllerServices(farmID int, controllerServices []common.ControllerService) {
	registry.controllerServices[farmID] = controllerServices
}

func (registry *DefaultServiceRegistry) GetControllerServices(farmID int) ([]common.ControllerService, error) {
	if services, ok := registry.controllerServices[farmID]; ok {
		return services, nil
	}
	return nil, ErrFarmNotFound
}

func (registry *DefaultServiceRegistry) SetEventLogService(eventLogService EventLogService) {
	registry.eventLogService = eventLogService
}

func (registry *DefaultServiceRegistry) GetEventLogService() EventLogService {
	return registry.eventLogService
}

func (registry *DefaultServiceRegistry) SetFarmFactory(farmFactory *FarmFactory) {
	registry.farmFactory = farmFactory
}

func (registry *DefaultServiceRegistry) GetFarmFactory() *FarmFactory {
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

func (registry *DefaultServiceRegistry) SetFarmServices(farmServices map[int]FarmService) {
	registry.farmServicesMutex.Lock()
	defer registry.farmServicesMutex.Unlock()
	registry.farmServices = farmServices
}

func (registry *DefaultServiceRegistry) GetFarmServices() map[int]FarmService {
	registry.farmServicesMutex.RLock()
	defer registry.farmServicesMutex.RUnlock()
	return registry.farmServices
}

func (registry *DefaultServiceRegistry) GetFarmService(farmID int) (FarmService, bool) {
	registry.farmServicesMutex.RLock()
	defer registry.farmServicesMutex.RUnlock()
	service, ok := registry.farmServices[farmID]
	return service, ok
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

func (registry *DefaultServiceRegistry) SetMailer(mailer common.Mailer) {
	registry.mailer = mailer
}

func (registry *DefaultServiceRegistry) GetMailer() common.Mailer {
	return registry.mailer
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
