package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/service"
)

type FarmRestService interface {
	Config(w http.ResponseWriter, r *http.Request)
	State(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultFarmRestService struct {
	farmService service.FarmService
	middleware  service.Middleware
	jsonWriter  common.HttpWriter
	FarmRestService
}

func NewFarmRestService(middleware service.Middleware, jsonWriter common.HttpWriter) FarmRestService {
	return &DefaultFarmRestService{
		middleware: middleware,
		jsonWriter: jsonWriter}
}

func (restService *DefaultFarmRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	// /farms/{farmID}
	//endpoint := fmt.Sprintf("%s/%s", baseURI, restService.farmService.GetConfig().GetID())
	endpoint := baseFarmURI
	// /farms/{farmID}/config
	configEndpoint := fmt.Sprintf("%s/config", endpoint)
	// /farms/{farmID}/config/{controllerID}/{key}?value=foo
	setConfigEndpoint := fmt.Sprintf("%s/{controllerID}/{key}", configEndpoint)
	// /farms/{farmID}/state
	stateEndpoint := fmt.Sprintf("%s/state", endpoint)
	router.Handle(configEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Config)),
	))
	router.Handle(stateEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.State)),
	))
	router.Handle(setConfigEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Set)),
	))
	return []string{endpoint, configEndpoint, stateEndpoint, setConfigEndpoint}
}

func (restService *DefaultFarmRestService) Config(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debugf("[DefaultFarmRestService.Config] REST service /config request email=%s", session.GetUser().GetEmail())

	restService.jsonWriter.Write(w, http.StatusOK, session.GetFarmService().GetConfig())
}

func (restService *DefaultFarmRestService) State(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debugf("[DefaultFarmRestService.State] REST service /state request email=%s", session.GetUser().GetEmail())

	restService.jsonWriter.Write(w, http.StatusOK, session.GetFarmService().GetState())
}

func (restService *DefaultFarmRestService) Set(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	farmID := params["farmID"]
	controllerID := params["controllerID"]
	key := params["key"]
	value := r.FormValue("value")

	//restService.session.GetLogger().Debugf("[ConfDefaultFarmRestServiceigRestService.SetServer] controllerID=%s, key=%s, value=%s, params=%+v", controllerID, key, value, params)

	intFarmID, err := strconv.ParseInt(farmID, 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	intControllerID, err := strconv.ParseInt(controllerID, 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	/*
		if err := restService.configService.SetValue(session, int(intFarmID), int(intControllerID), key, value); err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}*/
	if err := session.GetFarmService().SetConfigValue(session, int(intFarmID), int(intControllerID), key, value); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
