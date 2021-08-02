// +build cluster

package rest

import (
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type ClusterRestServiceRegistry struct {
	jwtService      service.JsonWebTokenService
	serviceRegistry service.ServiceRegistry
	services        []RestService
	RestServiceRegistry
}

func NewClusterRestServiceRegistry(publicKey string, mapperRegistry mapper.MapperRegistry, serviceRegistry service.ServiceRegistry) RestServiceRegistry {

	jsonWriter := NewJsonWriter()
	jwtService := serviceRegistry.GetJsonWebTokenService()

	restServices := make([]RestService, 0)

	//configRestService := NewConfigRestService(serviceRegistry.GetConfigService(), jwtService, jsonWriter)
	channelRestService := NewChannelRestService(serviceRegistry.GetChannelService(), mapperRegistry.GetChannelMapper(), jwtService, jsonWriter)
	metricRestService := NewMetricRestService(serviceRegistry.GetMetricService(), mapperRegistry.GetMetricMapper(), jwtService, jsonWriter)
	conditionRestService := NewConditionRestService(serviceRegistry.GetConditionService(), mapperRegistry.GetConditionMapper(), jwtService, jsonWriter)
	scheduleRestService := NewScheduleRestService(serviceRegistry.GetScheduleService(), jwtService, jsonWriter)
	algorithmRestService := NewAlgorithmRestService(serviceRegistry.GetAlgorithmService(), jwtService, jsonWriter)
	deviceFactoryRestService := NewDeviceFactoryRestService(serviceRegistry.GetDeviceFactory(), jwtService, jsonWriter)
	workflowRestService := NewWorkflowRestService(serviceRegistry.GetWorkflowService(), jwtService, jsonWriter)
	workflowStepRestService := NewWorkflowStepRestService(serviceRegistry.GetWorkflowStepService(), jwtService, jsonWriter)
	googleRestService := NewGoogleRestService(serviceRegistry.GetGoogleAuthService(), jwtService, jsonWriter)

	//restServices = append(restServices, configRestService)
	restServices = append(restServices, channelRestService)
	restServices = append(restServices, metricRestService)
	restServices = append(restServices, conditionRestService)
	restServices = append(restServices, scheduleRestService)
	restServices = append(restServices, algorithmRestService)
	restServices = append(restServices, deviceFactoryRestService)
	restServices = append(restServices, googleRestService)
	restServices = append(restServices, NewFarmRestService(publicKey, jwtService, jsonWriter))
	restServices = append(restServices, NewDeviceRestService(serviceRegistry, jwtService, jsonWriter))
	restServices = append(restServices, workflowRestService)
	restServices = append(restServices, workflowStepRestService)

	// Create unique list of device types
	/*
		deviceServices := make(map[string]common.DeviceService, 0)
		for _, farmService := range serviceRegistry.GetFarmServices() {
			devices, err := serviceRegistry.GetDeviceServices(farmService.GetFarmID())
			if err != nil {
				log.Fatal(err)
			}
			for _, cservice := range devices {
				if _, ok := deviceServices[cservice.GetDeviceType()]; !ok {
					deviceServices[cservice.GetDeviceType()] = cservice
					restServices = append(restServices, NewDeviceRestService(cservice, jwtService, jsonWriter))
				}
			}
		}*/

	return &ClusterRestServiceRegistry{
		jwtService: jwtService,
		services:   restServices}
}

func (registry *ClusterRestServiceRegistry) GetRestServices() []RestService {
	return registry.services
}
