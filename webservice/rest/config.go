// +build notnow

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

type ConfigRestService interface {
	//Get(w http.ResponseWriter, r *http.Request)
	Set(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultConfigRestService struct {
	session       service.Session
	configService service.ConfigService
	middleware    service.Middleware
	jsonWriter    common.HttpWriter
}

//func NewConfigRestService(session service.Session, configService service.ConfigService, middleware service.Middleware,
func NewConfigRestService(configService service.ConfigService, middleware service.Middleware,
	jsonWriter common.HttpWriter) ConfigRestService {

	return &DefaultConfigRestService{
		//session:       session,
		configService: configService,
		middleware:    middleware,
		jsonWriter:    jsonWriter}
}

func (restService *DefaultConfigRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	configEndpoint := fmt.Sprintf("%s/config", baseFarmURI)
	setConfigEndpoint := fmt.Sprintf("%s/{controllerID}/{key}", configEndpoint)
	/*
		router.Handle(configEndpoint, negroni.New(
			negroni.HandlerFunc(restService.middleware.Validate),
			negroni.Wrap(http.HandlerFunc(restService.Get)),
		))*/
	router.Handle(setConfigEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Set)),
	))
	return []string{configEndpoint, setConfigEndpoint}
}

/*
func (restService *DefaultConfigRestService) Get(w http.ResponseWriter, r *http.Request) {
	//restService.session.GetLogger().Debugf("[ConfigRestService.Index]")
	//restService.jsonWriter.Write(w, http.StatusOK, restService.session.GetFarmService().GetConfig())
	restService.jsonWriter.Write(w, http.StatusOK, restService.configService.GetServerConfig())
}*/

func (restService *DefaultConfigRestService) Set(w http.ResponseWriter, r *http.Request) {

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

	//restService.session.GetLogger().Debugf("[ConfigRestService.SetServer] controllerID=%s, key=%s, value=%s, params=%+v", controllerID, key, value, params)

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

	if err := restService.configService.SetValue(session, int(intFarmID), int(intControllerID), key, value); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
