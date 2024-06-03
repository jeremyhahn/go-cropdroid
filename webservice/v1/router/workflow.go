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

type WorkflowRouter struct {
	middleware          middleware.JsonWebTokenMiddleware
	workflowRestService rest.WorkflowRestServicer
	WebServiceRouter
}

// Creates a new web service workflow router
func NewWorkflowRouter(
	workflowService service.WorkflowService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &WorkflowRouter{
		middleware: middleware,
		workflowRestService: rest.NewWorkflowRestService(
			workflowService,
			middleware,
			httpWriter)}
}

// Registers all of the workflow endpoints at the root of the farm (/api/v1/farm/{farmID})
func (workflowRouter *WorkflowRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	workflowsBaseURI := fmt.Sprintf("%s/workflows", baseFarmURI)
	return []string{
		workflowRouter.view(router, workflowsBaseURI),
		workflowRouter.workflows(router, workflowsBaseURI),
		workflowRouter.workflow(router, workflowsBaseURI),
		workflowRouter.run(router, workflowsBaseURI),
		workflowRouter.create(router, workflowsBaseURI),
		workflowRouter.update(router, workflowsBaseURI),
		workflowRouter.delete(router, workflowsBaseURI)}
}

// @Summary List workflows for UI view
// @Description Returns a list of workflows to be consumed by a UI
// @Tags Workflows
// @Produce  json
// @Param	farmID	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/view [get]
// @Security JWT
func (workflowRouter *WorkflowRouter) view(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/view", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowRouter.workflowRestService.View)),
	))
	return endpoint
}

// @Summary List workflows
// @Description Returns a list of workflows
// @Tags Workflows
// @Produce  json
// @Param	farmID	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows [get]
// @Security JWT
func (workflowRouter *WorkflowRouter) workflows(router *mux.Router, workflowsBaseURI string) string {
	router.Handle(workflowsBaseURI, negroni.New(
		negroni.HandlerFunc(workflowRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowRouter.workflowRestService.GetWorkflows)),
	))
	return workflowsBaseURI
}

// @Summary Get workflow
// @Description Returns the requested workflow
// @Tags Workflows
// @Produce  json
// @Param	farmID	path	integer	true	"string valid"
// @Param	id		path	integer	true	"string valid"	"Workflow ID"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/{id} [get]
// @Security JWT
func (workflowRouter *WorkflowRouter) workflow(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/{id}", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowRouter.workflowRestService.GetWorkflow)),
	))
	return endpoint
}

// @Summary Run workflow
// @Description Runs a workflow
// @Tags Workflows
// @Produce  json
// @Param	farmID	path	integer	true	"string valid"
// @Param	id		path	integer	true	"string valid"	"Workflow ID"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/{id}/run [get]
// @Security JWT
func (workflowRouter *WorkflowRouter) run(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/{id}/run", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowRouter.workflowRestService.RunWorkflow)),
	))
	return endpoint
}

// @Summary Create workflow
// @Description Creates a new workflow
// @Tags Workflows
// @Produce  json
// @Param	farmID	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows [post]
// @Security JWT
func (workflowRouter *WorkflowRouter) create(router *mux.Router, workflowsBaseURI string) string {
	router.Handle(workflowsBaseURI, negroni.New(
		negroni.HandlerFunc(workflowRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowRouter.workflowRestService.Create)),
	)).Methods("POST")
	return workflowsBaseURI
}

// @Summary Update workflow
// @Description Updates an existing workflow
// @Tags Workflows
// @Produce  json
// @Param	farmID	path	integer	true	"string valid"
// @Param	id		path	integer	true	"string valid"	"Workflow ID"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/{id} [put]
// @Security JWT
func (workflowRouter *WorkflowRouter) update(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/{id}", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowRouter.workflowRestService.Update)),
	)).Methods("PUT")
	return endpoint
}

// @Summary Delete workflow
// @Description Deletes an existing workflow
// @Tags Workflows
// @Produce  json
// @Param	farmID	path	integer	true	"string valid"
// @Param	id		path	integer	true	"string valid"	"Workflow ID"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/workflows/{id} [delete]
// @Security JWT
func (workflowRouter *WorkflowRouter) delete(router *mux.Router, workflowsBaseURI string) string {
	endpoint := fmt.Sprintf("%s/{id}", workflowsBaseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(workflowRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(workflowRouter.workflowRestService.Delete)),
	)).Methods("DELETE")
	return endpoint
}
