package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type WorkflowRestService interface {
	GetWorkflow(w http.ResponseWriter, r *http.Request)
	GetWorkflows(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	RunWorkflow(w http.ResponseWriter, r *http.Request)
	View(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultWorkflowRestService struct {
	workflowService service.WorkflowService
	middleware      service.Middleware
	jsonWriter      common.HttpWriter
	WorkflowRestService
}

func NewWorkflowRestService(workflowService service.WorkflowService, middleware service.Middleware,
	jsonWriter common.HttpWriter) WorkflowRestService {

	return &DefaultWorkflowRestService{
		workflowService: workflowService,
		middleware:      middleware,
		jsonWriter:      jsonWriter}
}

func (restService *DefaultWorkflowRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	endpoint := fmt.Sprintf("%s/workflows", baseFarmURI)
	workflowEndpoint := fmt.Sprintf("%s/{id}", endpoint)
	workflowViewEndpoint := fmt.Sprintf("%s/view", endpoint)
	workflowRunEndpoint := fmt.Sprintf("%s/run", workflowEndpoint)
	router.Handle(workflowViewEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.View)),
	)).Methods("GET")
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetWorkflows)),
	)).Methods("GET")
	router.Handle(workflowEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetWorkflow)),
	)).Methods("GET")
	router.Handle(workflowRunEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.RunWorkflow)),
	)).Methods("GET")
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Create)),
	)).Methods("POST")
	router.Handle(workflowEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Update)),
	)).Methods("PUT")
	router.Handle(workflowEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Delete)),
	)).Methods("DELETE")
	return []string{endpoint, workflowEndpoint}
}

func (restService *DefaultWorkflowRestService) View(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()
	logger.Debugf("REST service /view request email=%s", session.GetUser().GetEmail())

	params := mux.Vars(r)
	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
	if err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	workflows, err := restService.workflowService.GetListView(session, farmID)
	if err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, workflows)
}

func (restService *DefaultWorkflowRestService) GetWorkflow(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	params := mux.Vars(r)
	workflowID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	logger.Debugf("workflowID=%d", workflowID)

	workflow, err := restService.workflowService.GetWorkflow(session, workflowID)
	if err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		restService.jsonWriter.Error500(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, workflow)
}

func (restService *DefaultWorkflowRestService) GetWorkflows(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		session.GetLogger().Errorf("session: %s, error: %s", session, err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	restService.jsonWriter.Write(w, http.StatusOK, restService.workflowService.GetWorkflows(session))
}

func (restService *DefaultWorkflowRestService) Create(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()
	logger.Debug("Decoding JSON request")

	var workflow config.Workflow
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&workflow); err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	logger.Debugf("workflow=%+v", workflow)

	persisted, err := restService.workflowService.Create(session, &workflow)
	if err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, persisted)
}

func (restService *DefaultWorkflowRestService) Update(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	workflow := config.NewWorkflow()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&workflow); err != nil {
		logger.Errorf("session: %s, error: %s", session, err.Error())
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	logger.Debugf("Decoding JSON request: %s", buf)
	logger.Debugf("workflow=%+v", workflow)

	if err = restService.workflowService.Update(session, workflow); err != nil {
		logger.Errorf("session: %s, error: %s", session, err.Error())
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}

func (restService *DefaultWorkflowRestService) Delete(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	logger.Debugf("workflow.id=%d", id)

	if err = restService.workflowService.Delete(session, &config.Workflow{ID: id}); err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}

func (restService *DefaultWorkflowRestService) RunWorkflow(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	logger.Debugf("workflow.id=%d", id)

	if err = restService.workflowService.Run(session, id); err != nil {
		logger.Errorf("session: %s, error: %s", session, err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
