package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type ProvisionerRestService interface {
	Provision(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultProvisionerRestService struct {
	app               *app.App
	userService       service.UserService
	farmProvisioner   provisioner.FarmProvisioner
	middlewareService service.Middleware
	jsonWriter        common.HttpWriter
	ProvisionerRestService
}

func NewProvisionerRestService(app *app.App,
	userService service.UserService,
	farmProvisioner provisioner.FarmProvisioner,
	middlewareService service.Middleware,
	jsonWriter common.HttpWriter) ProvisionerRestService {

	return &DefaultProvisionerRestService{
		app:               app,
		userService:       userService,
		farmProvisioner:   farmProvisioner,
		middlewareService: middlewareService,
		jsonWriter:        jsonWriter}
}

func (restService *DefaultProvisionerRestService) RegisterEndpoints(
	router *mux.Router, baseURI, baseFarmURI string) []string {

	provisionerEndpoint := fmt.Sprintf("%s/provisioner", baseURI)
	provisionEndpoint := fmt.Sprintf("%s/provision/{orgId}/{farmName}", provisionerEndpoint)
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

	params := mux.Vars(r)
	orgID, err := strconv.ParseUint(params["orgId"], 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	userID := session.GetUser().GetID()
	user, err := restService.userService.Get(userID)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	roleID := user.GetRoles()[0].GetID()
	provisionerParams := &common.ProvisionerParams{
		UserID:           userID,
		RoleID:           roleID,
		OrganizationID:   orgID,
		FarmName:         params["farmName"],
		ConfigStoreType:  restService.app.DefaultConfigStoreType,
		StateStoreType:   restService.app.DefaultStateStoreType,
		DataStoreType:    restService.app.DefaultDataStoreType,
		ConsistencyLevel: restService.app.DefaultConsistencyLevel}
	farmConfig, err := restService.farmProvisioner.Provision(user, provisionerParams)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("farmConfig=%+v", farmConfig)

	restService.jsonWriter.Success200(w, farmConfig)
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

	restService.jsonWriter.Success200(w, true)
}
