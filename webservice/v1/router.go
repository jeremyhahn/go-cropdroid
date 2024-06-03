package v1

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/router"
)

type RouterV1 struct {
	app                      *app.App
	baseURI                  string
	baseFarmURI              string
	mapperRegistry           mapper.MapperRegistry
	serviceRegistry          service.ServiceRegistry
	restServiceRegistry      rest.RestServiceRegistry
	jsonWebTokenMiddleware   middleware.JsonWebTokenMiddleware
	farmWebSocketRestService rest.FarmWebSocketRestServicer
	router                   *mux.Router
	responseWriter           response.HttpWriter
	endpointList             []string
	router.WebServiceRouter
}

func NewRouterV1(
	app *app.App,
	mapperRegistry mapper.MapperRegistry,
	serviceRegistry service.ServiceRegistry,
	restServiceRegistry rest.RestServiceRegistry,
	farmWebSocketRestService rest.FarmWebSocketRestServicer,
	router *mux.Router,
	responseWriter response.HttpWriter) router.WebServiceRouter {

	return &RouterV1{
		app:                      app,
		mapperRegistry:           mapperRegistry,
		serviceRegistry:          serviceRegistry,
		restServiceRegistry:      restServiceRegistry,
		jsonWebTokenMiddleware:   restServiceRegistry.JsonWebTokenService(),
		farmWebSocketRestService: farmWebSocketRestService,
		router:                   router,
		responseWriter:           responseWriter,
		endpointList:             make([]string, 0)}
}

// Registers all routes for standalone mode
func (v1Router *RouterV1) RegisterRoutes(router *mux.Router, baseURI string) []string {
	v1Router.baseURI = baseURI
	v1Router.baseFarmURI = fmt.Sprintf("%s/farms/{farmID}", baseURI)
	endpointList := make([]string, 0)
	endpointList = append(endpointList, v1Router.systemRoutes()...)
	endpointList = append(endpointList, v1Router.registrationRoutes()...)
	endpointList = append(endpointList, v1Router.authenticationRoutes()...)
	endpointList = append(endpointList, v1Router.farmRoutes()...)
	endpointList = append(endpointList, v1Router.algorithmRoutes()...)
	endpointList = append(endpointList, v1Router.channelRoutes()...)
	endpointList = append(endpointList, v1Router.conditionRoutes()...)
	endpointList = append(endpointList, v1Router.deviceRoutes()...)
	endpointList = append(endpointList, v1Router.googleRoutes()...)
	endpointList = append(endpointList, v1Router.metricRoutes()...)
	endpointList = append(endpointList, v1Router.organizationRoutes()...)
	endpointList = append(endpointList, v1Router.provisionerRoutes()...)
	endpointList = append(endpointList, v1Router.roleRoutes()...)
	endpointList = append(endpointList, v1Router.scheduleRoutes()...)
	endpointList = append(endpointList, v1Router.shoppingCartRoutes()...)
	endpointList = append(endpointList, v1Router.workflowStepRoutes()...)
	endpointList = append(endpointList, v1Router.workflowRoutes()...)
	endpointList = append(endpointList, v1Router.eventLogRoutes()...)
	endpoints := v1Router.sortAndDeDupe(endpointList)
	v1Router.app.Logger.Debug(strings.Join(endpoints[:], "\n"))
	v1Router.app.Logger.Debugf("Loaded %d REST endpoints", len(endpoints))
	v1Router.endpointList = endpoints
	return endpoints
}

// Registers all routes except those which will be served by the cluster router
func (v1Router *RouterV1) registerNonClusterRoutes(router *mux.Router, baseURI string) []string {
	v1Router.baseURI = baseURI
	v1Router.baseFarmURI = fmt.Sprintf("%s/farms/{farmID}", baseURI)
	endpointList := make([]string, 0)
	endpointList = append(endpointList, v1Router.registrationRoutes()...)
	endpointList = append(endpointList, v1Router.authenticationRoutes()...)
	endpointList = append(endpointList, v1Router.farmRoutes()...)
	endpointList = append(endpointList, v1Router.algorithmRoutes()...)
	endpointList = append(endpointList, v1Router.channelRoutes()...)
	endpointList = append(endpointList, v1Router.conditionRoutes()...)
	endpointList = append(endpointList, v1Router.deviceRoutes()...)
	endpointList = append(endpointList, v1Router.googleRoutes()...)
	endpointList = append(endpointList, v1Router.metricRoutes()...)
	endpointList = append(endpointList, v1Router.organizationRoutes()...)
	endpointList = append(endpointList, v1Router.provisionerRoutes()...)
	endpointList = append(endpointList, v1Router.roleRoutes()...)
	endpointList = append(endpointList, v1Router.scheduleRoutes()...)
	endpointList = append(endpointList, v1Router.shoppingCartRoutes()...)
	endpointList = append(endpointList, v1Router.workflowStepRoutes()...)
	endpointList = append(endpointList, v1Router.workflowRoutes()...)
	endpointList = append(endpointList, v1Router.eventLogRoutes()...)
	return endpointList
}

func (v1Router *RouterV1) sortAndDeDupe(endpointList []string) []string {
	// Create unique list
	uniqueList := make(map[string]bool, len(endpointList))
	for _, endpoint := range endpointList {
		uniqueList[endpoint] = true
	}
	// Create a new array from the unique list
	endpoints := make([]string, len(uniqueList))
	i := 0
	for k, _ := range uniqueList {
		endpoints[i] = k
		i++
	}
	// Sort the endpoints
	sort.Strings(endpoints)
	return endpoints
}

func (v1Router *RouterV1) systemRoutes() []string {
	systemRouter := router.NewSystemRouter(
		v1Router.app,
		v1Router.serviceRegistry,
		v1Router.restServiceRegistry.JsonWebTokenService(),
		v1Router.router,
		v1Router.responseWriter,
		&v1Router.endpointList)
	return systemRouter.RegisterRoutes(v1Router.router, v1Router.baseURI)
}

func (v1Router *RouterV1) registrationRoutes() []string {
	registrationRouter := router.NewRegistrationRouter(
		v1Router.app,
		v1Router.serviceRegistry.GetUserService(),
		v1Router.responseWriter)
	return registrationRouter.RegisterRoutes(v1Router.router, v1Router.baseURI)
}

func (v1Router *RouterV1) authenticationRoutes() []string {
	registrationRouter := router.NewAuthenticationRouter(
		v1Router.app,
		v1Router.jsonWebTokenMiddleware.(middleware.AuthMiddleware))
	return registrationRouter.RegisterRoutes(v1Router.router, v1Router.baseURI)
}

func (v1Router *RouterV1) farmRoutes() []string {
	farmRouter := router.NewFarmRouter(
		v1Router.app.Domain,
		v1Router.app.CA,
		v1Router.baseFarmURI,
		v1Router.serviceRegistry,
		v1Router.jsonWebTokenMiddleware,
		v1Router.farmWebSocketRestService,
		v1Router.responseWriter)
	return farmRouter.RegisterRoutes(v1Router.router, v1Router.baseURI)
}

func (v1Router *RouterV1) algorithmRoutes() []string {
	algorithmRouter := router.NewAlgorithmRouter(
		v1Router.restServiceRegistry.AlgorithmRestService())
	return algorithmRouter.RegisterRoutes(v1Router.router, v1Router.baseURI)
}

func (v1Router *RouterV1) channelRoutes() []string {
	channelRouter := router.NewChannelRouter(
		v1Router.serviceRegistry.GetChannelService(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return channelRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) conditionRoutes() []string {
	conditionRouter := router.NewConditionRouter(
		v1Router.mapperRegistry.GetConditionMapper(),
		v1Router.serviceRegistry.GetConditionService(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return conditionRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) deviceRoutes() []string {
	deviceRouter := router.NewDeviceRouter(
		v1Router.serviceRegistry,
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return deviceRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) googleRoutes() []string {
	deviceRouter := router.NewGoogleRouter(
		v1Router.serviceRegistry.GetGoogleAuthService(),
		v1Router.jsonWebTokenMiddleware.(middleware.AuthMiddleware),
		v1Router.responseWriter)
	return deviceRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) metricRoutes() []string {
	metricRouter := router.NewMetricRouter(
		v1Router.app.Logger,
		v1Router.serviceRegistry.GetMetricService(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return metricRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) organizationRoutes() []string {
	orgRouter := router.NewOrganizationRouter(
		v1Router.serviceRegistry.GetOrganizationService(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return orgRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) provisionerRoutes() []string {
	orgRouter := router.NewProvisionerRouter(
		v1Router.app,
		v1Router.serviceRegistry.GetUserService(),
		v1Router.serviceRegistry.GetFarmProvisioner(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return orgRouter.RegisterRoutes(v1Router.router, v1Router.baseURI)
}

func (v1Router *RouterV1) roleRoutes() []string {
	roleRouter := router.NewRoleRouter(
		v1Router.serviceRegistry.GetRoleService(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return roleRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) scheduleRoutes() []string {
	scheduleRouter := router.NewScheduleRouter(
		v1Router.serviceRegistry.GetScheduleService(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return scheduleRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) shoppingCartRoutes() []string {
	if v1Router.app.Stripe == nil {
		return []string{}
	}
	shoppingCartRouter := router.NewShoppingCartRouter(
		v1Router.app.Logger,
		v1Router.app.Stripe.Key.Webook,
		v1Router.serviceRegistry.GetShoppingCartService(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return shoppingCartRouter.RegisterRoutes(v1Router.router, v1Router.baseURI)
}

func (v1Router *RouterV1) workflowStepRoutes() []string {
	workflowStepRouter := router.NewWorkflowStepRouter(
		v1Router.serviceRegistry.GetWorkflowStepService(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return workflowStepRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) workflowRoutes() []string {
	workflowRouter := router.NewWorkflowRouter(
		v1Router.serviceRegistry.GetWorkflowService(),
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return workflowRouter.RegisterRoutes(v1Router.router, v1Router.baseFarmURI)
}

func (v1Router *RouterV1) eventLogRoutes() []string {
	eventLogRouter := router.NewEventLogRouter(
		v1Router.app,
		v1Router.app.Logger,
		v1Router.serviceRegistry,
		v1Router.jsonWebTokenMiddleware,
		v1Router.responseWriter)
	return eventLogRouter.RegisterRoutes(v1Router.router, v1Router.baseURI)
}
