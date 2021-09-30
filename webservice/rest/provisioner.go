package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type ProvisionerRestService interface {
	Provision(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultProvisionerRestService struct {
	farmProvisioner   provisioner.FarmProvisioner
	middlewareService service.Middleware
	jsonWriter        common.HttpWriter
	ProvisionerRestService
}

func NewProvisionerRestService(farmProvisioner provisioner.FarmProvisioner,
	middlewareService service.Middleware, jsonWriter common.HttpWriter) ProvisionerRestService {

	return &DefaultProvisionerRestService{
		farmProvisioner:   farmProvisioner,
		middlewareService: middlewareService,
		jsonWriter:        jsonWriter}
}

func (restService *DefaultProvisionerRestService) RegisterEndpoints(
	router *mux.Router, baseURI, baseFarmURI string) []string {

	provisionerEndpoint := fmt.Sprintf("%s/provisioner", baseURI)
	provisionEndpoint := fmt.Sprintf("%s/provision", provisionerEndpoint)
	deprovisionEndpoint := fmt.Sprintf("%s/deprovision/{farmID}", provisionerEndpoint)
	router.Handle(provisionEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Provision)),
	)).Methods("POST")
	router.Handle(deprovisionEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Deprovision)),
	)).Methods("DELETE")
	return []string{provisionerEndpoint, provisionEndpoint, deprovisionEndpoint}
}

func (restService *DefaultProvisionerRestService) Provision(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := &provisioner.ProvisionerParams{}
	user := session.GetUser()
	farmConfig, err := restService.farmProvisioner.Provision(user, params)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("farmConfig=%+v", farmConfig)

	restService.jsonWriter.Write(w, http.StatusOK, farmConfig)
}

func (restService *DefaultProvisionerRestService) Deprovision(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("/deprovision/%d", farmID)

	err = restService.farmProvisioner.Deprovision(session.GetUser(), farmID)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, true)
}
