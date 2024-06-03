package rest

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type RoleRestServicer interface {
	GetListView(w http.ResponseWriter, r *http.Request)
	RestService
}

type RoleRestService struct {
	roleService service.RoleServicer
	middleware  middleware.JsonWebTokenMiddleware
	httpWriter  response.HttpWriter
	RoleRestServicer
}

func NewRoleRestService(
	roleService service.RoleServicer,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) RoleRestServicer {

	return &RoleRestService{
		roleService: roleService,
		middleware:  middleware,
		httpWriter:  httpWriter}
}

func (restService *RoleRestService) GetListView(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	page := params["page"]

	p, err := strconv.Atoi(page)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}

	pageQuery := query.NewPageQuery()
	pageQuery.Page = p
	pageQuery.SortOrder = query.SORT_DESCENDING
	pageResult, err := restService.roleService.GetPage(pageQuery)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, pageResult)
}

func (restService *RoleRestService) SetRole(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	// var role config.Schedule
	// decoder := json.NewDecoder(r.Body)
	// if err := decoder.Decode(&schedule); err != nil {
	// 	BadRequestError(w, r, err, restService.jsonWriter)
	// 	return
	// }
	// persisted, err := restService.scheduleService.Create(session, &schedule)
	// if err != nil {
	// 	session.GetLogger().Errorf("Error: ", err)
	// 	restService.jsonWriter.Error200(w, err)
	// 	return
	// }
	restService.httpWriter.Success200(w, r, nil)
}
