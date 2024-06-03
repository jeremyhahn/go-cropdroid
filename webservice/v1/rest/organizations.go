package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type OrganizationRestServicer interface {
	GetListView(w http.ResponseWriter, r *http.Request)
	//GetOrganizations(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	// Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Page(w http.ResponseWriter, r *http.Request)
	GetUsers(w http.ResponseWriter, r *http.Request)
	RestService
}

type OrganizationRestService struct {
	orgService service.OrganizationService
	middleware middleware.JsonWebTokenMiddleware
	httpWriter response.HttpWriter
	OrganizationRestServicer
}

func NewOrganizationRestService(
	orgService service.OrganizationService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) OrganizationRestServicer {

	return &OrganizationRestService{
		orgService: orgService,
		middleware: middleware,
		httpWriter: httpWriter}
}

func (restService *OrganizationRestService) Page(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	params := mux.Vars(r)
	page := params["page"]
	p, err := strconv.Atoi(page)
	if err != nil {
		logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}
	pageQuery := query.NewPageQuery()
	pageQuery.Page = p
	orgs, err := restService.orgService.Page(session, pageQuery)
	if err != nil {
		restService.httpWriter.Error500(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, orgs)
}

func (restService *OrganizationRestService) Create(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var org *config.OrganizationStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(org); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	err = restService.orgService.Create(org)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, org)
}

func (restService *OrganizationRestService) GetUsers(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	users, err := restService.orgService.GetUsers(session)
	if err != nil {
		restService.httpWriter.Error500(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, users)
}

func (restService *OrganizationRestService) Delete(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	if err = restService.orgService.Delete(session); err != nil {
		restService.httpWriter.Error500(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

// func (restService *OrganizationRestService) GetOrganizations(w http.ResponseWriter, r *http.Request) {

// 	session, err := restService.JsonWebTokenMiddleware.CreateSession(w, r)
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

// func (restService *OrganizationRestService) Create(w http.ResponseWriter, r *http.Request) {

// 	session, err := restService.JsonWebTokenMiddleware.CreateSession(w, r)
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

// func (restService *OrganizationRestService) Update(w http.ResponseWriter, r *http.Request) {

// 	session, err := restService.JsonWebTokenMiddleware.CreateSession(w, r)
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
