package rest

import (
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type FreewareRestServiceRegistry struct {
	jwtService      service.JsonWebTokenService
	serviceRegistry service.ServiceRegistry
	services        []RestService
	RestServiceRegistry
}

func NewFreewareRestServiceRegistry(mapperRegistry mapper.MapperRegistry, serviceRegistry service.ServiceRegistry) RestServiceRegistry {

	jsonWriter := NewJsonWriter()
	jwtService := serviceRegistry.GetJsonWebTokenService()

	restServices := make([]RestService, 0)

	//configRestService := NewConfigRestService(serviceRegistry.GetConfigService(), jwtService, jsonWriter)
	channelRestService := NewChannelRestService(serviceRegistry.GetChannelService(), mapperRegistry.GetChannelMapper(), jwtService, jsonWriter)
	metricRestService := NewMetricRestService(serviceRegistry.GetMetricService(), mapperRegistry.GetMetricMapper(), jwtService, jsonWriter)
	conditionRestService := NewConditionRestService(serviceRegistry.GetConditionService(), mapperRegistry.GetConditionMapper(), jwtService, jsonWriter)
	scheduleRestService := NewScheduleRestService(serviceRegistry.GetScheduleService(), jwtService, jsonWriter)
	algorithmRestService := NewAlgorithmRestService(serviceRegistry.GetAlgorithmService(), jwtService, jsonWriter)
	controllerFactoryRestService := NewControllerFactoryRestService(serviceRegistry.GetControllerFactory(), jwtService, jsonWriter)

	//restServices = append(restServices, configRestService)
	restServices = append(restServices, channelRestService)
	restServices = append(restServices, metricRestService)
	restServices = append(restServices, conditionRestService)
	restServices = append(restServices, scheduleRestService)
	restServices = append(restServices, algorithmRestService)
	restServices = append(restServices, controllerFactoryRestService)
	restServices = append(restServices, NewFarmRestService(jwtService, jsonWriter))
	restServices = append(restServices, NewControllerRestService(serviceRegistry, jwtService, jsonWriter))

	/*
		for _, farmService := range serviceRegistry.GetFarmServices() {
			controllerServices, _ := farmService.BuildControllerServices()
			for _, controllerService := range controllerServices {
				restServices = append(restServices, NewControllerRestService(
					controllerService, controllerService.GetControllerType(), jwtService, jsonWriter))
			}
		}
	*/

	/*
		// Create unique list of controller types
		controllerServices := make(map[string]common.ControllerService, 0)
		for _, farmService := range serviceRegistry.GetFarmServices() {
			controllers, err := serviceRegistry.GetControllerServices(farmService.GetFarmID())
			if err != nil {
				log.Fatal(err)
			}
			for _, cservice := range controllers {
				if _, ok := controllerServices[cservice.GetControllerType()]; !ok {
					controllerServices[cservice.GetControllerType()] = cservice
					restServices = append(restServices, NewControllerRestService(cservice, jwtService, jsonWriter))
				}
			}
		}*/

	return &FreewareRestServiceRegistry{
		jwtService: jwtService,
		services:   restServices}
}

func (registry *FreewareRestServiceRegistry) GetRestServices() []RestService {
	return registry.services
}
