//go:build cluster
// +build cluster

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
)

type ClusterSystemRouter struct {
	systemRestService rest.SystemRestServicer
	SystemRouter
	WebServiceRouter
}

// Creates a new web service system router that overrides the
// status and eventlog endpoints
func NewClusterSystemRouter(
	app *app.App,
	clusterServiceRegistry service.ClusterServiceRegistry,
	middleware middleware.JsonWebTokenMiddleware,
	router *mux.Router,
	jsonWriter response.HttpWriter,
	endpointList *[]string) WebServiceRouter {

	return &ClusterSystemRouter{
		systemRestService: rest.NewClusterSystemRestService(
			app,
			clusterServiceRegistry,
			jsonWriter,
			endpointList),
		SystemRouter: SystemRouter{
			middleware: middleware,
			systemRestService: rest.NewSystemRestService(
				app,
				clusterServiceRegistry,
				jsonWriter,
				endpointList)}}
}

// Registers all of the system endpoints at the root of the webservice (/api/v1)
func (systemRouter *ClusterSystemRouter) RegisterRoutes(router *mux.Router, baseURI string) []string {
	return []string{
		systemRouter.endpoints(router, baseURI),
		systemRouter.status(router, baseURI),
		systemRouter.pubkey(router, baseURI),
		systemRouter.config(router, baseURI),
		systemRouter.eventlog(router, baseURI)}
}

func (systemRouter *ClusterSystemRouter) status(router *mux.Router, baseURI string) string {
	endpoint := fmt.Sprintf("%s/status", baseURI)
	router.Handle(endpoint, negroni.New(
		negroni.NewLogger(),
		negroni.HandlerFunc(systemRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(systemRouter.systemRestService.Status)),
	))
	return endpoint
}

func (systemRouter *ClusterSystemRouter) eventlog(router *mux.Router, baseURI string) string {
	endpoint := fmt.Sprintf("%s/events/{page}", baseURI)
	router.Handle(endpoint, negroni.New(
		negroni.NewLogger(),
		negroni.HandlerFunc(systemRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(systemRouter.systemRestService.EventsPage)),
	))
	return endpoint
}
