package rest

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type DeviceFactoryRestService interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultMicroFactoryRestService struct {
	serviceRegistry service.ServiceRegistry
	middleware      service.Middleware
	jsonWriter      common.HttpWriter
	DeviceFactoryRestService
}

func NewDeviceFactoryRestService(serviceRegistry service.ServiceRegistry,
	middleware service.Middleware, jsonWriter common.HttpWriter) DeviceFactoryRestService {

	return &DefaultMicroFactoryRestService{
		serviceRegistry: serviceRegistry,
		middleware:      middleware,
		jsonWriter:      jsonWriter}
}

func (restService *DefaultMicroFactoryRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	getDevicesEndpoint := fmt.Sprintf("%s/devices", baseFarmURI)
	router.Handle(getDevicesEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetAll)),
	)).Methods("GET")
	return []string{getDevicesEndpoint}
}

func (restService *DefaultMicroFactoryRestService) GetAll(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debug("getting devices")

	deviceFactory := restService.serviceRegistry.GetDeviceFactory()
	devices, err := deviceFactory.GetDevices(session)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("devices=%v+", devices)

	restService.jsonWriter.Write(w, http.StatusOK, devices)
}
