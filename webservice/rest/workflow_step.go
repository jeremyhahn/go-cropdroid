package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type WorkflowStepRestService interface {
	GetStep(w http.ResponseWriter, r *http.Request)
	GetSteps(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultWorkflowStepRestService struct {
	workflowStepService service.WorkflowStepService
	middleware          service.Middleware
	jsonWriter          common.HttpWriter
	WorkflowStepRestService
}

func NewWorkflowStepRestService(workflowStepService service.WorkflowStepService, middleware service.Middleware,
	jsonWriter common.HttpWriter) WorkflowStepRestService {

	return &DefaultWorkflowStepRestService{
		workflowStepService: workflowStepService,
		middleware:          middleware,
		jsonWriter:          jsonWriter}
}

func (restService *DefaultWorkflowStepRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	endpoint := fmt.Sprintf("%s/workflows/{workflowID}/steps", baseFarmURI)
	stepEndpoint := fmt.Sprintf("%s/{stepID}", endpoint)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetSteps)),
	)).Methods("GET")
	router.Handle(stepEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetStep)),
	)).Methods("GET")
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Create)),
	)).Methods("POST")
	router.Handle(stepEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Update)),
	)).Methods("PUT")
	router.Handle(stepEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Delete)),
	)).Methods("DELETE")
	return []string{endpoint, stepEndpoint}
}

func (restService *DefaultWorkflowStepRestService) GetStep(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	workflowID, err := strconv.ParseUint(params["workflowID"], 10, 64)
	stepID, err := strconv.ParseUint(params["stepID"], 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("workflowID=%d,stepID=%d", workflowID, stepID)

	step, err := restService.workflowStepService.GetStep(session, workflowID, stepID)
	if err != nil {
		session.GetLogger().Errorf("Error: ", err)
		restService.jsonWriter.Error500(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, step)
}

func (restService *DefaultWorkflowStepRestService) GetSteps(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	workflowID, err := strconv.ParseUint(params["workflowID"], 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("workflowID=%d", workflowID)

	steps, err := restService.workflowStepService.GetSteps(session, workflowID)
	if err != nil {
		session.GetLogger().Errorf("Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, steps)
}

func (restService *DefaultWorkflowStepRestService) Create(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debug("Decoding JSON request")

	var workflowStep config.WorkflowStep
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&workflowStep); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("workflowStep=%+v", workflowStep)

	persisted, err := restService.workflowStepService.Create(session, &workflowStep)
	if err != nil {
		session.GetLogger().Errorf("Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, persisted)
}

func (restService *DefaultWorkflowStepRestService) Update(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debug("Decoding JSON request")

	var workflowStep config.WorkflowStep
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&workflowStep); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("workflowStep=%+v", workflowStep)

	if err = restService.workflowStepService.Update(session, &workflowStep); err != nil {
		session.GetLogger().Errorf("Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}

func (restService *DefaultWorkflowStepRestService) Delete(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	workflowID, err := strconv.ParseUint(params["workflowID"], 10, 64)
	stepID, err := strconv.ParseUint(params["stepID"], 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("workflowID=%d,stepID=%d", workflowID, stepID)

	workflowStep := config.NewWorkflowStep()
	workflowStep.SetID(stepID)
	workflowStep.SetWorkflowID(workflowID)

	if err = restService.workflowStepService.Delete(session, workflowStep); err != nil {
		session.GetLogger().Errorf("Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
