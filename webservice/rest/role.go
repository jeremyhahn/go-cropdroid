package rest

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type RoleRestService interface {
	GetListView(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultRoleRestService struct {
	roleService service.RoleService
	middleware  service.Middleware
	jsonWriter  common.HttpWriter
	RoleRestService
}

func NewRoleRestService(roleService service.RoleService,
	middleware service.Middleware, jsonWriter common.HttpWriter) RoleRestService {

	return &DefaultRoleRestService{
		roleService: roleService,
		middleware:  middleware,
		jsonWriter:  jsonWriter}
}

func (restService *DefaultRoleRestService) RegisterEndpoints(
	router *mux.Router, baseURI, baseFarmURI string) []string {

	rolesEndpoint := fmt.Sprintf("%s/roles", baseURI)
	router.Handle(rolesEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetListView)),
	)).Methods("GET")
	return []string{rolesEndpoint}
}

func (restService *DefaultRoleRestService) GetListView(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()
	roles, err := restService.roleService.GetAll()
	if err != nil {
		session.GetLogger().Errorf("DefaultRoleRestService error: %s", err)
		restService.jsonWriter.Error500(w, err)
		return
	}
	restService.jsonWriter.Success200(w, roles)
}

func (restService *DefaultRoleRestService) SetRole(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debug("Decoding JSON request")

	// var role config.Schedule
	// decoder := json.NewDecoder(r.Body)
	// if err := decoder.Decode(&schedule); err != nil {
	// 	BadRequestError(w, r, err, restService.jsonWriter)
	// 	return
	// }

	// session.GetLogger().Debugf("schedule=%+v", schedule)

	// persisted, err := restService.scheduleService.Create(session, &schedule)
	// if err != nil {
	// 	session.GetLogger().Errorf("Error: ", err)
	// 	restService.jsonWriter.Error200(w, err)
	// 	return
	// }

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
