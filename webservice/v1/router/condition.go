package router

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

type ConditionRouter struct {
	middleware           middleware.JsonWebTokenMiddleware
	conditionRestService rest.ConditionRestServicer
	WebServiceRouter
}

// Creates a new web service condition router
func NewConditionRouter(
	conditionMapper mapper.ConditionMapper,
	conditionService service.ConditionServicer,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &ConditionRouter{
		middleware: middleware,
		conditionRestService: rest.NewConditionRestService(
			conditionMapper,
			conditionService,
			middleware,
			httpWriter)}
}

// Registers all of the condition endpoints for a farm
func (conditionRouter *ConditionRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	return []string{
		conditionRouter.listView(router, baseFarmURI),
		conditionRouter.create(router, baseFarmURI),
		conditionRouter.update(router, baseFarmURI),
		conditionRouter.delete(router, baseFarmURI)}
}

// @Summary List channel conditions
// @Description Returns all conditions for the requested channel
// @Tags Conditions
// @Accept json
// @Produce  json
// @Param	farmID	path	integer	true	"string valid"
// @Param   page	path	integer	false	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/conditions/{channelID} [get]
// @Security JWT
func (conditionRouter *ConditionRouter) listView(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/conditions/{channelID}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(conditionRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(conditionRouter.conditionRestService.ListView)),
	))
	return endpoint
}

// @Summary Create channel condition
// @Description Creates a new channel condition
// @Tags Conditions
// @Accept json
// @Produce  json
// @Param	farmID		path	integer			true	"string valid"
// @Param	condition	body	config.Condition true	"config.Condition struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/conditions [post]
// @Security JWT
func (conditionRouter *ConditionRouter) create(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/conditions", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(conditionRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(conditionRouter.conditionRestService.Create)),
	)).Methods(http.MethodPost)
	return endpoint
}

// @Summary Update channel condition
// @Description Updates an existing channel condition
// @Tags Conditions
// @Accept json
// @Produce  json
// @Param	farmID		path	integer			true	"string valid"
// @Param	condition	body	config.Condition true	"config.Condition struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/conditions [put]
// @Security JWT
func (conditionRouter *ConditionRouter) update(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/conditions", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(conditionRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(conditionRouter.conditionRestService.Update)),
	)).Methods(http.MethodPut)
	return endpoint
}

// @Summary Delete channel condition
// @Description Deletes an existing channel condition
// @Tags Conditions
// @Accept json
// @Produce  json
// @Param	farmID	path	integer	true	"string valid"
// @Param	id		path	integer true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/conditions [delete]
// @Security JWT
func (conditionRouter *ConditionRouter) delete(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/conditions/{id}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(conditionRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(conditionRouter.conditionRestService.Delete)),
	)).Methods(http.MethodDelete)
	return endpoint
}
