package rest

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type DefaultRestServiceRegistry struct {
	jwtService      service.JsonWebTokenService
	serviceRegistry service.ServiceRegistry
	services        []RestService
	RestServiceRegistry
}

func NewRestServiceRegistry(app *app.App, publicKey string, mapperRegistry mapper.MapperRegistry, serviceRegistry service.ServiceRegistry) RestServiceRegistry {

	jsonWriter := NewJsonWriter()
	jwtService := serviceRegistry.GetJsonWebTokenService()
	farmProvisioner := serviceRegistry.GetFarmProvisioner()

	restServices := make([]RestService, 0)

	channelRestService := NewChannelRestService(serviceRegistry.GetChannelService(), mapperRegistry.GetChannelMapper(), jwtService, jsonWriter)
	metricRestService := NewMetricRestService(serviceRegistry.GetMetricService(), mapperRegistry.GetMetricMapper(), jwtService, jsonWriter)
	conditionRestService := NewConditionRestService(serviceRegistry.GetConditionService(), mapperRegistry.GetConditionMapper(), jwtService, jsonWriter)
	scheduleRestService := NewScheduleRestService(serviceRegistry.GetScheduleService(), jwtService, jsonWriter)
	algorithmRestService := NewAlgorithmRestService(serviceRegistry.GetAlgorithmService(), jwtService, jsonWriter)
	deviceFactoryRestService := NewDeviceFactoryRestService(serviceRegistry, jwtService, jsonWriter)
	workflowRestService := NewWorkflowRestService(serviceRegistry.GetWorkflowService(), jwtService, jsonWriter)
	workflowStepRestService := NewWorkflowStepRestService(serviceRegistry.GetWorkflowStepService(), jwtService, jsonWriter)
	googleRestService := NewGoogleRestService(serviceRegistry.GetGoogleAuthService(), jwtService, jsonWriter)
	provisionerRestService := NewProvisionerRestService(app, serviceRegistry.GetUserService(), farmProvisioner, jwtService, jsonWriter)
	roleRestService := NewRoleRestService(serviceRegistry.GetRoleService(), jwtService, jsonWriter)
	organizationRestService := NewOrganizationRestService(serviceRegistry.GetOrganizationService(), jwtService, jsonWriter)

	restServices = append(restServices, channelRestService)
	restServices = append(restServices, metricRestService)
	restServices = append(restServices, conditionRestService)
	restServices = append(restServices, scheduleRestService)
	restServices = append(restServices, algorithmRestService)
	restServices = append(restServices, deviceFactoryRestService)
	restServices = append(restServices, googleRestService)
	restServices = append(restServices, NewFarmRestService(publicKey,
		serviceRegistry.GetFarmFactory(), serviceRegistry.GetUserService(),
		jwtService, jsonWriter))
	restServices = append(restServices, NewDeviceRestService(
		serviceRegistry, jwtService, jsonWriter))
	restServices = append(restServices, workflowRestService)
	restServices = append(restServices, workflowStepRestService)
	restServices = append(restServices, provisionerRestService)
	restServices = append(restServices, organizationRestService)
	restServices = append(restServices, roleRestService)

	return &DefaultRestServiceRegistry{
		jwtService: jwtService,
		services:   restServices}
}

func (registry *DefaultRestServiceRegistry) GetRestServices() []RestService {
	return registry.services
}
