package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type FarmRestService interface {
	Config(w http.ResponseWriter, r *http.Request)
	State(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultFarmRestService struct {
	publicKey  string
	middleware service.Middleware
	jsonWriter common.HttpWriter
	FarmRestService
}

func NewFarmRestService(publicKey string, middleware service.Middleware,
	jsonWriter common.HttpWriter) FarmRestService {
	return &DefaultFarmRestService{
		publicKey:  publicKey,
		middleware: middleware,
		jsonWriter: jsonWriter}
}

func (restService *DefaultFarmRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	// /farms/{farmID}
	//endpoint := fmt.Sprintf("%s/%s", baseURI, restService.farmService.GetConfig().GetID())
	endpoint := baseFarmURI
	// /farms/{farmID}/config
	configEndpoint := fmt.Sprintf("%s/config", endpoint)
	// /farms/{farmID}/config/{deviceID}/{key}?value=foo
	setConfigEndpoint := fmt.Sprintf("%s/{deviceID}/{key}", configEndpoint)
	// /farms/{farmID}/state
	stateEndpoint := fmt.Sprintf("%s/state", endpoint)
	// /farms/{farmID}/pubkey
	pubKeyEndpoint := fmt.Sprintf("%s/pubkey", endpoint)
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
	router.Handle(pubKeyEndpoint, http.HandlerFunc(restService.PublicKey))
	return []string{endpoint, configEndpoint, stateEndpoint, setConfigEndpoint}
}

func (restService *DefaultFarmRestService) Config(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debugf("REST service /config request email=%s", session.GetUser().GetEmail())

	restService.jsonWriter.Write(w, http.StatusOK, session.GetFarmService().GetConfig())
}

func (restService *DefaultFarmRestService) State(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debugf("REST service /farms/{farmID}/state request email=%s", session.GetUser().GetEmail())

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
	deviceID := params["deviceID"]
	key := params["key"]
	value := r.FormValue("value")

	//restService.session.GetLogger().Debugf("deviceID=%s, key=%s, value=%s, params=%+v", deviceID, key, value, params)

	uint64FarmID, err := strconv.ParseUint(farmID, 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	uint64DeviceID, err := strconv.ParseUint(deviceID, 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	/*
		if err := restService.configService.SetValue(session, int(intFarmID), int(intDeviceID), key, value); err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}*/
	if err := session.GetFarmService().SetConfigValue(session, uint64FarmID, uint64DeviceID, key, value); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}

func (restService *DefaultFarmRestService) PublicKey(w http.ResponseWriter, r *http.Request) {

	/*
		session, err := restService.middleware.CreateSession(w, r)
		if err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}
		defer session.Close()

		session.GetLogger().Debugf("REST service /farms/{farmID}/pubkey request email=%s", session.GetUser().GetEmail())

		restService.jsonWriter.Write(w, http.StatusOK, session.GetFarmService().GetPublicKey())
	*/

	restService.jsonWriter.Write(w, http.StatusOK, restService.publicKey)
}
