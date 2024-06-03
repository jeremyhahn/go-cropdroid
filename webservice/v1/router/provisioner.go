package router

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

type ProvisionerRouter struct {
	middleware             middleware.JsonWebTokenMiddleware
	provisionerRestService rest.ProvisionerRestServicer
	WebServiceRouter
}

// Creates a new web service provisioner router
func NewProvisionerRouter(
	app *app.App,
	userService service.UserServicer,
	farmProvisioner provisioner.FarmProvisioner,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &ProvisionerRouter{
		middleware: middleware,
		provisionerRestService: rest.NewProvisionerRestService(
			app,
			userService,
			farmProvisioner,
			middleware,
			httpWriter)}
}

// Registers all of the provisioner endpoints at the root of the webservice (/api/v1)
func (provisionerRouter *ProvisionerRouter) RegisterRoutes(router *mux.Router, baseURI string) []string {
	return []string{
		provisionerRouter.provision(router, baseURI),
		provisionerRouter.deprovision(router, baseURI)}
}

// @Summary Provision new farm
// @Description Creates a new farm using system configured defaults
// @Tags Provisioner
// @Produce  json
// @Param   orgId		path	string	true	"string default"     default(0)
// @Param   farmName	path	string	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /provisioner/provision/{orgId}/{farmName} [post]
// @Security JWT
func (provisionerRouter *ProvisionerRouter) provision(router *mux.Router, baseURI string) string {
	endpoint := fmt.Sprintf("%s/provisioner/provision/{orgId}/{farmName}", baseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(provisionerRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(provisionerRouter.provisionerRestService.Provision)),
	))
	return endpoint
}

// @Summary Derovision existing farm
// @Description Deletes the farm and all associated data, running services, and API endpoints
// @Tags Provisioner
// @Produce  json
// @Param   farmID		path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /provisioner/deprovision/{farmID} [post]
// @Security JWT
func (provisionerRouter *ProvisionerRouter) deprovision(router *mux.Router, baseURI string) string {
	endpoint := fmt.Sprintf("%s/provisioner/deprovision/{farmID}", baseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(provisionerRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(provisionerRouter.provisionerRestService.DeProvision)),
	))
	return endpoint
}
