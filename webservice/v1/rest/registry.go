package rest

import (
	"sync"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/websocket"
)

type RestRegistry struct {
	app                      *app.App
	mapperRegistry           mapper.MapperRegistry
	serviceRegistry          service.ServiceRegistry
	algorithmRestService     AlgorithmRestServicer
	channelRestService       ChannelRestServicer
	conditionRestService     ConditionRestServicer
	deviceRestService        DeviceRestServicer
	eventLogRestService      EventLogRestServicer
	farmWebSocketRestService FarmWebSocketRestServicer
	googleRestService        GoogleRestServicer
	jsonWebTokenService      JsonWebTokenServicer
	metricRestService        MetricRestServicer
	organizationRestService  OrganizationRestServicer
	provisionerRestService   ProvisionerRestServicer
	registrationRestService  RegistrationRestServicer
	roleRestService          RoleRestServicer
	scheduleRestService      ScheduleRestServicer
	shoppingCartRestService  ShoppingCartRestServicer
	systemRestService        SystemRestServicer
	workflowRestService      WorkflowRestServicer
	workflowStepRestService  WorkflowStepRestServicer
	farmHubs                 map[uint64]*websocket.FarmHub
	farmsMutex               *sync.RWMutex
	notificationHubs         map[uint64]*websocket.NotificationHub
	notificationsMutex       *sync.RWMutex
	endpointList             *[]string
	RestServiceRegistry
}

func NewRestServiceRegistry(
	app *app.App,
	roleDAO dao.RoleDAO,
	mapperRegistry mapper.MapperRegistry,
	serviceRegistry service.ServiceRegistry) RestServiceRegistry {

	endpointList := make([]string, 0)
	registry := &RestRegistry{
		app:                app,
		mapperRegistry:     mapperRegistry,
		serviceRegistry:    serviceRegistry,
		farmHubs:           make(map[uint64]*websocket.FarmHub, 0),
		farmsMutex:         &sync.RWMutex{},
		notificationHubs:   make(map[uint64]*websocket.NotificationHub, 0),
		notificationsMutex: &sync.RWMutex{},
		endpointList:       &endpointList}

	registry.createJsonWebTokenService(roleDAO)

	httpWriter := response.NewResponseWriter(app.Logger, nil)
	farmProvisioner := serviceRegistry.GetFarmProvisioner()

	stripeWebhookKey := ""
	if app.Stripe != nil && app.Stripe.Key != nil && app.Stripe.Key.Webook != "" {
		stripeWebhookKey = app.Stripe.Key.Webook
	}

	registry.algorithmRestService = NewAlgorithmRestService(
		serviceRegistry.GetAlgorithmService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.channelRestService = NewChannelRestService(
		serviceRegistry.GetChannelService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.deviceRestService = NewDeviceRestService(
		serviceRegistry,
		registry.jsonWebTokenService,
		httpWriter)

	registry.metricRestService = NewMetricRestService(
		app.Logger,
		serviceRegistry.GetMetricService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.conditionRestService = NewConditionRestService(
		mapperRegistry.GetConditionMapper(),
		serviceRegistry.GetConditionService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.scheduleRestService = NewScheduleRestService(
		serviceRegistry.GetScheduleService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.eventLogRestService = NewEventLogRestService(
		app,
		app.Logger,
		serviceRegistry,
		registry.jsonWebTokenService,
		httpWriter)

	registry.farmWebSocketRestService = NewFarmWebSocketRestService(
		app.Logger,
		registry.farmHubs,
		registry.farmsMutex,
		registry.notificationHubs,
		registry.notificationsMutex,
		serviceRegistry,
		registry.jsonWebTokenService,
		httpWriter)

	registry.googleRestService = NewGoogleRestService(
		serviceRegistry.GetGoogleAuthService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.organizationRestService = NewOrganizationRestService(
		serviceRegistry.GetOrganizationService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.provisionerRestService = NewProvisionerRestService(
		app,
		serviceRegistry.GetUserService(),
		farmProvisioner,
		registry.jsonWebTokenService,
		httpWriter)

	registry.roleRestService = NewRoleRestService(
		serviceRegistry.GetRoleService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.shoppingCartRestService = NewShoppingCartRestService(
		app.Logger,
		stripeWebhookKey,
		serviceRegistry.GetShoppingCartService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.systemRestService = NewSystemRestService(
		app,
		serviceRegistry,
		httpWriter,
		registry.endpointList)

	registry.workflowRestService = NewWorkflowRestService(
		serviceRegistry.GetWorkflowService(),
		registry.jsonWebTokenService,
		httpWriter)

	registry.workflowStepRestService = NewWorkflowStepRestService(
		serviceRegistry.GetWorkflowStepService(),
		registry.jsonWebTokenService,
		httpWriter)

	return registry
}

func (registry *RestRegistry) AlgorithmRestService() AlgorithmRestServicer {
	return registry.algorithmRestService
}

func (registry *RestRegistry) ChannelRestService() ChannelRestServicer {
	return registry.channelRestService
}

func (registry *RestRegistry) ConditionRestService() ConditionRestServicer {
	return registry.conditionRestService
}

func (registry *RestRegistry) DeviceRestService() DeviceRestServicer {
	return registry.deviceRestService
}

func (registry *RestRegistry) EventLogRestService() EventLogRestServicer {
	return registry.eventLogRestService
}

func (registry *RestRegistry) FarmWebSocketRestService() FarmWebSocketRestServicer {
	return registry.farmWebSocketRestService
}

func (registry *RestRegistry) GoogleRestService() GoogleRestServicer {
	return registry.googleRestService
}

func (registry *RestRegistry) JsonWebTokenService() JsonWebTokenServicer {
	return registry.jsonWebTokenService
}

func (registry *RestRegistry) MetricRestService() MetricRestServicer {
	return registry.metricRestService
}

func (registry *RestRegistry) OrganizationRestService() OrganizationRestServicer {
	return registry.organizationRestService
}

func (registry *RestRegistry) ProvisionerRestService() ProvisionerRestServicer {
	return registry.provisionerRestService
}

func (registry *RestRegistry) RegistrationRequest() RegistrationRestServicer {
	return registry.registrationRestService
}

func (registry *RestRegistry) RoleRestService() RoleRestServicer {
	return registry.roleRestService
}

func (registry *RestRegistry) ScheduleRestService() ScheduleRestServicer {
	return registry.scheduleRestService
}

func (registry *RestRegistry) ShoppingCartRestService() ShoppingCartRestServicer {
	return registry.shoppingCartRestService
}

func (registry *RestRegistry) SystemRestService() SystemRestServicer {
	return registry.systemRestService
}

func (registry *RestRegistry) WorkflowRestService() WorkflowRestServicer {
	return registry.workflowRestService
}

func (registry *RestRegistry) WorkflowStepRestService() WorkflowStepRestServicer {
	return registry.workflowStepRestService
}

func (registry *RestRegistry) createJsonWebTokenService(roleDAO dao.RoleDAO) {
	httpWriter := response.NewResponseWriter(registry.app.Logger, nil)
	defaultRole, err := roleDAO.GetByName(registry.app.DefaultRole, common.CONSISTENCY_LOCAL)
	if err != nil {
		registry.app.Logger.Fatal(err)
	}
	jsonWebTokenService, err := CreateJsonWebTokenService(registry.app,
		registry.app.IdGenerator, defaultRole, registry.mapperRegistry.GetDeviceMapper(),
		registry.serviceRegistry, httpWriter, registry.app.JwtExpiration) // 1 year jwt expiration
	if err != nil {
		registry.app.Logger.Fatal(err)
	}
	registry.jsonWebTokenService = jsonWebTokenService
}
