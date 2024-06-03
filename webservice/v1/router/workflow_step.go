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

type WorkflowStepRouter struct {
	middleware              middleware.JsonWebTokenMiddleware
	workflowStepRestService rest.WorkflowStepRestServicer
	WebServiceRouter
}

// Creates a new web service workflowStep router
func NewWorkflowStepRouter(
	workflowStepService service.WorkflowStepService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &WorkflowStepRouter{
		middleware: middleware,
		workflowStepRestService: rest.NewWorkflowStepRestService(
			workflowStepService,
			middleware,
			httpWriter)}
}

// Registers all of the workflowStep endpoints at the root of the farm (/api/v1/farm/{farmID})
func (workflowStepRouter *WorkflowStepRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	workflowsBaseURI := fmt.Sprintf("%s/workflows", baseFarmURI)
	return []string{
		workflowStepRouter.steps(router, workflowsBaseURI),
		workflowStepRouter.step(router, workflowsBaseURI),
		workflowStepRouter.create(router, workflowsBaseURI),
		workflowStepRouter.update(router, workflowsBaseURI),
		workflowStepRouter.delete(router, workflowsBaseURI)}
}

// @Summary List workflow steps
// @Description Returns all of the steps (with current state) associated with a workflow.
// @Tags Workflows
// @Produce  json
// @Param	farmID		path	integer				true	"string valid"
// @Param   workflowID	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/{workflowID}/steps [get]
// @Security JWT
func (workflowStepRouter *WorkflowStepRouter) steps(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/{workflowID}/steps", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowStepRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowStepRouter.workflowStepRestService.GetSteps)),
	))
	return endpoint
}

// @Summary Get workflow step
// @Description Returns the specified workflow step
// @Tags Workflows
// @Produce  json
// @Param	farmID		path	integer				true	"string valid"
// @Param   workflowID	path	integer	true	"string valid"
// @Param   stepID		path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/{workflowID}/steps/{stepID} [get]
// @Security JWT
func (workflowStepRouter *WorkflowStepRouter) step(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/{workflowID}/steps/{stepID}", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowStepRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowStepRouter.workflowStepRestService.GetStep)),
	))
	return endpoint
}

// @Summary Create workflow step
// @Description Creates a new workflow step
// @Tags Workflows
// @Produce  json
// @Param	farmID		path	integer				true	"string valid"
// @Param	workflowID	path	integer	true	"string valid"
// @Param	WorkflowStep body config.WorkflowStep true "config.WorkflowStep struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/{workflowID}/steps [post]
// @Security JWT
func (workflowStepRouter *WorkflowStepRouter) create(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/{workflowID}/steps", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowStepRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowStepRouter.workflowStepRestService.Create)),
	)).Methods("POST")
	return endpoint
}

// @Summary Update workflow step
// @Description Updates an existing workflow step
// @Tags Workflows
// @Produce  json
// @Param	farmID			path	integer				true	"string valid"
// @Param	workflowID		path	integer				true	"string valid"
// @Param   stepID			path	integer				true	"string valid"
// @Param	WorkflowStep	body	config.WorkflowStep true	"config.WorkflowStep struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/{workflowID}/steps/{stepID} [put]
// @Security JWT
func (workflowStepRouter *WorkflowStepRouter) update(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/{workflowID}/steps/{stepID}", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowStepRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowStepRouter.workflowStepRestService.Update)),
	)).Methods("PUT")
	return endpoint
}

// @Summary Delete workflow step
// @Description Delete an existing workflow step
// @Tags Workflows
// @Produce  json
// @Param	farmID			path	integer				true	"string valid"
// @Param	workflowID		path	integer				true	"string valid"
// @Param   stepID			path	integer				true	"string valid"
// @Param	WorkflowStep	body 	config.WorkflowStep true	"config.WorkflowStep struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/{workflowID}/steps [delete]
// @Security JWT
func (workflowStepRouter *WorkflowStepRouter) delete(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/{workflowID}/steps/{stepID}", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowStepRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowStepRouter.workflowStepRestService.Delete)),
	)).Methods("DELETE")
	return endpoint
}
