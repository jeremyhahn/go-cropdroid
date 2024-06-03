package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type WorkflowRestServicer interface {
	GetWorkflow(w http.ResponseWriter, r *http.Request)
	GetWorkflows(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	RunWorkflow(w http.ResponseWriter, r *http.Request)
	View(w http.ResponseWriter, r *http.Request)
	RestService
}

type WorkflowRestService struct {
	workflowService service.WorkflowService
	middleware      middleware.JsonWebTokenMiddleware
	httpWriter      response.HttpWriter
	WorkflowRestServicer
}

func NewWorkflowRestService(
	workflowService service.WorkflowService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WorkflowRestServicer {

	return &WorkflowRestService{
		workflowService: workflowService,
		middleware:      middleware,
		httpWriter:      httpWriter}
}

func (restService *WorkflowRestService) View(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	workflows, err := restService.workflowService.GetListView(session, farmID)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, workflows)
}

func (restService *WorkflowRestService) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	workflowID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	workflow, err := restService.workflowService.GetWorkflow(session, workflowID)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, workflow)
}

func (restService *WorkflowRestService) GetWorkflows(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	restService.httpWriter.Success200(w, r, restService.workflowService.GetWorkflows(session))
}

func (restService *WorkflowRestService) Create(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var workflow *config.WorkflowStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(workflow); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	persisted, err := restService.workflowService.Create(session, workflow)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, persisted)
}

func (restService *WorkflowRestService) Update(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	workflow := config.NewWorkflow()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&workflow); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err = restService.workflowService.Update(session, workflow); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, nil)
}

func (restService *WorkflowRestService) Delete(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err = restService.workflowService.Delete(session, &config.WorkflowStruct{ID: id}); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

func (restService *WorkflowRestService) RunWorkflow(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err = restService.workflowService.Run(session, id); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}
