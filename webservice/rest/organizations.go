package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type OrganizationRestService interface {
	GetListView(w http.ResponseWriter, r *http.Request)
	//GetOrganizations(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	// Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultOrganizationRestService struct {
	orgService service.OrganizationService
	middleware service.Middleware
	jsonWriter common.HttpWriter
	OrganizationRestService
}

func NewOrganizationRestService(orgService service.OrganizationService,
	middleware service.Middleware, jsonWriter common.HttpWriter) OrganizationRestService {

	return &DefaultOrganizationRestService{
		orgService: orgService,
		middleware: middleware,
		jsonWriter: jsonWriter}
}

func (restService *DefaultOrganizationRestService) RegisterEndpoints(
	router *mux.Router, baseURI, baseFarmURI string) []string {

	orgsEndpoint := fmt.Sprintf("%s/organizations", baseURI)
	orgUserEndpoint := fmt.Sprintf("%s/{organizationID}/users", orgsEndpoint)
	router.Handle(orgsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetAll)),
	)).Methods("GET")
	router.Handle(orgUserEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetUsers)),
	)).Methods("GET")
	router.Handle(orgsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Create)),
	)).Methods("POST")
	// router.Handle(orgsEndpoint, negroni.New(
	// 	negroni.HandlerFunc(restService.middleware.Validate),
	// 	negroni.Wrap(http.HandlerFunc(restService.Update)),
	// )).Methods("PUT")
	router.Handle(orgsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Delete)),
	)).Methods("DELETE")
	return []string{orgsEndpoint, orgUserEndpoint}
}

func (restService *DefaultOrganizationRestService) GetAll(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	orgs, err := restService.orgService.GetAll(session)
	if err != nil {
		session.GetLogger().Errorf("Error: %s", err)
		restService.jsonWriter.Error500(w, err)
		return
	}

	restService.jsonWriter.Success200(w, orgs)
}

func (restService *DefaultOrganizationRestService) Create(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	var org config.Organization
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&org); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("organization=%+v", org)

	err = restService.orgService.Create(&org)
	if err != nil {
		session.GetLogger().Errorf("Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Success200(w, org)
}

func (restService *DefaultOrganizationRestService) GetUsers(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debugf("organizationID=%d",
		session.GetRequestedOrganizationID())

	users, err := restService.orgService.GetUsers(session)
	if err != nil {
		session.GetLogger().Errorf("Error: %s", err)
		restService.jsonWriter.Error500(w, err)
		return
	}

	restService.jsonWriter.Success200(w, users)
}

func (restService *DefaultOrganizationRestService) Delete(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debugf("organizationID=%d",
		session.GetRequestedOrganizationID())

	if err = restService.orgService.Delete(session); err != nil {
		session.GetLogger().Errorf("Error: %s", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}

// func (restService *DefaultOrganizationRestService) GetOrganizations(w http.ResponseWriter, r *http.Request) {

// 	session, err := restService.middleware.CreateSession(w, r)
// 	if err != nil {
// 		BadRequestError(w, r, err, restService.jsonWriter)
// 		return
// 	}
// 	defer session.Close()

// 	params := mux.Vars(r)
// 	deviceID, err := strconv.Atoi(params["deviceID"])
// 	if err != nil {
// 		BadRequestError(w, r, err, restService.jsonWriter)
// 		return
// 	}

// 	session.GetLogger().Debugf("deviceID=%d", deviceID)

// 	orgs, err := restService.orgService.GetOrganizations(session, deviceID)
// 	if err != nil {
// 		session.GetLogger().Errorf("Error: ", err)
// 		restService.jsonWriter.Error500(w, err)
// 		return
// 	}

// 	restService.jsonWriter.Write(w, http.StatusOK, orgs)
// }

// func (restService *DefaultOrganizationRestService) Create(w http.ResponseWriter, r *http.Request) {

// 	session, err := restService.middleware.CreateSession(w, r)
// 	if err != nil {
// 		BadRequestError(w, r, err, restService.jsonWriter)
// 		return
// 	}
// 	defer session.Close()

// 	session.GetLogger().Debug("Decoding JSON request")

// 	var org config.Organization
// 	decoder := json.NewDecoder(r.Body)
// 	if err := decoder.Decode(&org); err != nil {
// 		BadRequestError(w, r, err, restService.jsonWriter)
// 		return
// 	}

// 	session.GetLogger().Debugf("org=%+v", org)

// 	persisted, err := restService.orgService.Create(session, &org)
// 	if err != nil {
// 		session.GetLogger().Errorf("Error: ", err)
// 		restService.jsonWriter.Error200(w, err)
// 		return
// 	}

// 	restService.jsonWriter.Write(w, http.StatusOK, persisted)
// }

// func (restService *DefaultOrganizationRestService) Update(w http.ResponseWriter, r *http.Request) {

// 	session, err := restService.middleware.CreateSession(w, r)
// 	if err != nil {
// 		BadRequestError(w, r, err, restService.jsonWriter)
// 		return
// 	}
// 	defer session.Close()

// 	session.GetLogger().Debug("Decoding JSON request")

// 	var org viewmodel.Organization
// 	decoder := json.NewDecoder(r.Body)
// 	if err := decoder.Decode(&org); err != nil {
// 		session.GetLogger().Errorf("Error: %s", err)
// 		BadRequestError(w, r, err, restService.jsonWriter)
// 		return
// 	}

// 	session.GetLogger().Debugf("org=%+v", org)

// 	orgConfig := restService.orgMapper.MapViewToConfig(org)
// 	if err = restService.orgService.Update(session, orgConfig); err != nil {
// 		session.GetLogger().Errorf("Error: %s", err)
// 		restService.jsonWriter.Error200(w, err)
// 		return
// 	}

// 	restService.jsonWriter.Write(w, http.StatusOK, nil)
// }
