package router

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

type OrganizationRouter struct {
	middleware              middleware.JsonWebTokenMiddleware
	organizationRestService rest.OrganizationRestServicer
	WebServiceRouter
}

// Creates a new web service organization router
func NewOrganizationRouter(
	organizationService service.OrganizationService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &OrganizationRouter{
		middleware: middleware,
		organizationRestService: rest.NewOrganizationRestService(
			organizationService,
			middleware,
			httpWriter)}
}

// Registers all of the organization endpoints at the root of the farm (/api/v1/farms/{farmID})
func (organizationRouter *OrganizationRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	return []string{
		organizationRouter.getPage(router, baseFarmURI),
		organizationRouter.getUsers(router, baseFarmURI),
		organizationRouter.create(router, baseFarmURI),
		organizationRouter.delete(router, baseFarmURI)}
}

// @Summary List organizations
// @Description Returns a page of organization entries
// @Tags Organization
// @Produce  json
// @Param   page	path	string	false	"string valid"	minlength(1)	maxlength(20)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /organizations/{page} [get]
// @Security JWT
func (organizationRouter *OrganizationRouter) getPage(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/organizations/{page}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(organizationRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(organizationRouter.organizationRestService.Page)),
	))
	return endpoint
}

// @Summary List all users in the organization
// @Description Returns all users in the organization
// @Tags Organization
// @Produce  json
// @Param   organizationID	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /organizations/{organizationID}/users [get]
// @Security JWT
func (organizationRouter *OrganizationRouter) getUsers(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/organizations/{organizationID}/users", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(organizationRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(organizationRouter.organizationRestService.GetUsers)),
	))
	return endpoint
}

// @Summary Create an organization
// @Description Creates a new organization
// @Tags Organization
// @Produce  json
// @Param   Organization	body	config.Organization	true	"config.Organization struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /organizations [post]
// @Security JWT
func (organizationRouter *OrganizationRouter) create(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/organizations/{organizationID}/users", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(organizationRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(organizationRouter.organizationRestService.Create)),
	))
	return endpoint
}

// @Summary Delete an organization
// @Description Deletes the organization referenced in the user JWT (current logged in organization)
// @Tags Organization
// @Produce  json
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /organizations [delete]
// @Security JWT
func (organizationRouter *OrganizationRouter) delete(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/organizations/{organizationID}/users", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(organizationRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(organizationRouter.organizationRestService.Delete)),
	))
	return endpoint
}
