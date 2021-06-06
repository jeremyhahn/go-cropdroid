package rest

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type ControllerFactoryRestService interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	RestService
}

type MicroControllerFactoryRestService struct {
	controllerFactory service.ControllerFactory
	middleware        service.Middleware
	jsonWriter        common.HttpWriter
	ControllerFactoryRestService
}

func NewControllerFactoryRestService(controllerFactory service.ControllerFactory,
	middleware service.Middleware, jsonWriter common.HttpWriter) ControllerFactoryRestService {

	return &MicroControllerFactoryRestService{
		controllerFactory: controllerFactory,
		middleware:        middleware,
		jsonWriter:        jsonWriter}
}

func (restService *MicroControllerFactoryRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	getControllersEndpoint := fmt.Sprintf("%s/devices", baseFarmURI)
	router.Handle(getControllersEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetAll)),
	)).Methods("GET")
	return []string{getControllersEndpoint}
}

func (restService *MicroControllerFactoryRestService) GetAll(w http.ResponseWriter, r *http.Request) {

	scope, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer scope.Close()

	scope.GetLogger().Debug("[ControllerFactoryRestService.GetAll] getting controllers")

	controllers, err := restService.controllerFactory.GetControllers(scope)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	scope.GetLogger().Debugf("[ControllerFactoryRestService.Get] controllers=%v+", controllers)

	restService.jsonWriter.Write(w, http.StatusOK, controllers)
}
