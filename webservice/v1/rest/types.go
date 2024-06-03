package rest

import (
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/op/go-logging"
)

type RestService interface {
	RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string
}

type RestServiceRegistry interface {
	AlgorithmRestService() AlgorithmRestServicer
	ChannelRestService() ChannelRestServicer
	ConditionRestService() ConditionRestServicer
	DeviceRestService() DeviceRestServicer
	EventLogRestService() EventLogRestServicer
	FarmWebSocketRestService() FarmWebSocketRestServicer
	GoogleRestService() GoogleRestServicer
	JsonWebTokenService() JsonWebTokenServicer
	MetricRestService() MetricRestServicer
	OrganizationRestService() OrganizationRestServicer
	ProvisionerRestService() ProvisionerRestServicer
	RegistrationRestService() RegistrationRestServicer
	RoleRestService() RoleRestServicer
	ScheduleRestService() ScheduleRestServicer
	ShoppingCartRestService() ShoppingCartRestServicer
	SystemRestService() SystemRestServicer
	WorkflowRestService() WorkflowRestServicer
	WorkflowStepRestService() WorkflowStepRestServicer
}

type GenericRestService[E any] struct {
	logger     *logging.Logger
	jsonWriter response.HttpWriter
	SystemRestService
}
