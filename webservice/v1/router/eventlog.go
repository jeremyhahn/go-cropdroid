package router

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
	"github.com/op/go-logging"
)

type EventLogRouter struct {
	middleware          middleware.JsonWebTokenMiddleware
	eventLogRestService rest.EventLogRestServicer
	WebServiceRouter
}

// Creates a new web service event log router
func NewEventLogRouter(
	app *app.App,
	logger *logging.Logger,
	serviceRegistry service.ServiceRegistry,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &EventLogRouter{
		middleware: middleware,
		eventLogRestService: rest.NewEventLogRestService(
			app,
			logger,
			serviceRegistry,
			middleware,
			httpWriter)}
}

// Registers all of the eventLog endpoints at the root of the webservice (/api/v1)
func (eventLogRouter *EventLogRouter) RegisterRoutes(router *mux.Router, baseURI string) []string {
	return []string{
		eventLogRouter.system(router, baseURI),
		eventLogRouter.farm(router, baseURI)}
}

// @Summary List events
// @Description Returns a page of event log entries
// @Tags Event Log
// @Accept json
// @Produce  json
// @Param   page	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /eventlog/{page} [get]
// @Security JWT
func (eventLogRouter *EventLogRouter) system(router *mux.Router, baseURI string) string {
	endpoint := fmt.Sprintf("%s/eventlog/{page}", baseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(eventLogRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(eventLogRouter.eventLogRestService.SystemPage)),
	))
	return endpoint
}

// @Summary List event log
// @Description Returns a page of event log entries for the requested farm
// @Tags Farms
// @Accept json
// @Produce  json
// @Param   farmID	path	integer	true	"string valid"
// @Param   page	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/events/{page} [get]
// @Security JWT
func (eventLogRouter *EventLogRouter) farm(router *mux.Router, baseURI string) string {
	endpoint := fmt.Sprintf("%s/farms/{farmID}/events/{page}", baseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(eventLogRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(eventLogRouter.eventLogRestService.FarmPage)),
	))
	return endpoint
}
