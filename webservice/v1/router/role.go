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

type RoleRouter struct {
	middleware      middleware.JsonWebTokenMiddleware
	roleRestService rest.RoleRestServicer
	WebServiceRouter
}

// Creates a new web service role router
func NewRoleRouter(
	roleService service.RoleServicer,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &RoleRouter{
		middleware: middleware,
		roleRestService: rest.NewRoleRestService(
			roleService,
			middleware,
			httpWriter)}
}

// Registers all of the role endpoints at the root of the webservice (/api/v1)
func (roleRouter *RoleRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	return []string{
		roleRouter.listView(router, baseFarmURI)}
}

// @Summary List roles
// @Description Returns a page of role entries
// @Tags Role
// @Produce  json
// //@Param   page	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /roles [get]
// @Security JWT
func (roleRouter *RoleRouter) listView(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/roles", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(roleRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(roleRouter.roleRestService.GetListView)),
	))
	return endpoint
}
