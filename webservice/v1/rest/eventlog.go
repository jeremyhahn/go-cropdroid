package rest

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	logging "github.com/op/go-logging"
)

type EventLogRestServicer interface {
	SystemPage(w http.ResponseWriter, r *http.Request)
	FarmPage(w http.ResponseWriter, r *http.Request)
}

type EventLogRestService struct {
	app               *app.App
	logger            *logging.Logger
	serviceRegistry   service.ServiceRegistry
	middlewareService middleware.JsonWebTokenMiddleware
	httpWriter        response.HttpWriter
	EventLogRestServicer
}

func NewEventLogRestService(
	app *app.App,
	logger *logging.Logger,
	serviceRegistry service.ServiceRegistry,
	middlewareService middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) EventLogRestServicer {

	return &EventLogRestService{
		app:               app,
		logger:            logger,
		serviceRegistry:   serviceRegistry,
		middlewareService: middlewareService,
		httpWriter:        httpWriter}
}

// Writes a page of farm event log entries
func (restService *EventLogRestService) FarmPage(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	page := params["page"]
	sFarmID := params["farmID"]
	farmID, err := strconv.ParseUint(sFarmID, 10, 64)
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
	}

	p, err := strconv.Atoi(page)
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	pageQuery := query.NewPageQuery()
	pageQuery.Page = p
	pageQuery.SortOrder = query.SORT_DESCENDING
	pageResult, err := restService.serviceRegistry.GetEventLogService(farmID).GetPage(pageQuery, common.CONSISTENCY_LOCAL)
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, pageResult)
}

// Writes a page of system event log entries
func (restService *EventLogRestService) SystemPage(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	page := params["page"]

	p, err := strconv.Atoi(page)
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	pageQuery := query.NewPageQuery()
	pageQuery.Page = p
	pageQuery.SortOrder = query.SORT_DESCENDING
	pageResult, err := restService.serviceRegistry.GetEventLogService(restService.app.ClusterID).GetPage(pageQuery, common.CONSISTENCY_LOCAL)
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, pageResult)
}
