package rest

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type ProvisionerRestServicer interface {
	Provision(w http.ResponseWriter, r *http.Request)
	DeProvision(w http.ResponseWriter, r *http.Request)
	RestService
}

type ProvisionerRestService struct {
	app             *app.App
	userService     service.UserServicer
	farmProvisioner provisioner.FarmProvisioner
	middleware      middleware.JsonWebTokenMiddleware
	httpWriter      response.HttpWriter
	ProvisionerRestServicer
}

func NewProvisionerRestService(
	app *app.App,
	userService service.UserServicer,
	farmProvisioner provisioner.FarmProvisioner,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) ProvisionerRestServicer {

	return &ProvisionerRestService{
		app:             app,
		userService:     userService,
		farmProvisioner: farmProvisioner,
		middleware:      middleware,
		httpWriter:      httpWriter}
}

func (restService *ProvisionerRestService) Provision(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	orgID, err := strconv.ParseUint(params["orgId"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	userID := session.GetUser().Identifier()
	user, err := restService.userService.Get(userID)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	roleID := user.GetRoles()[0].Identifier()
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
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, farmConfig)
}

func (restService *ProvisionerRestService) DeProvision(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	err = restService.farmProvisioner.Deprovision(session.GetUser(), farmID)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, true)
}
