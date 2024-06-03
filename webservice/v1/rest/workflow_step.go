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

type WorkflowStepRestServicer interface {
	GetStep(w http.ResponseWriter, r *http.Request)
	GetSteps(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultWorkflowStepRestService struct {
	workflowStepService service.WorkflowStepService
	middleware          middleware.JsonWebTokenMiddleware
	httpWriter          response.HttpWriter
	WorkflowStepRestServicer
}

func NewWorkflowStepRestService(
	workflowStepService service.WorkflowStepService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WorkflowStepRestServicer {

	return &DefaultWorkflowStepRestService{
		workflowStepService: workflowStepService,
		middleware:          middleware,
		httpWriter:          httpWriter}
}

func (restService *DefaultWorkflowStepRestService) GetStep(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	workflowID, err := strconv.ParseUint(params["workflowID"], 10, 64)
	stepID, err := strconv.ParseUint(params["stepID"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	step, err := restService.workflowStepService.GetStep(session, workflowID, stepID)
	if err != nil {
		restService.httpWriter.Error500(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, step)
}

func (restService *DefaultWorkflowStepRestService) GetSteps(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	workflowID, err := strconv.ParseUint(params["workflowID"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	steps, err := restService.workflowStepService.GetSteps(session, workflowID)
	if err != nil {
		session.GetLogger().Errorf("Error: ", err)
		restService.httpWriter.Error200(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, steps)
}

func (restService *DefaultWorkflowStepRestService) Create(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var workflowStep *config.WorkflowStepStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(workflowStep); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	persisted, err := restService.workflowStepService.Create(session, workflowStep)
	if err != nil {
		restService.httpWriter.Error200(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, persisted)
}

func (restService *DefaultWorkflowStepRestService) Update(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var workflowStep *config.WorkflowStepStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(workflowStep); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err = restService.workflowStepService.Update(session, workflowStep); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

func (restService *DefaultWorkflowStepRestService) Delete(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	workflowID, err := strconv.ParseUint(params["workflowID"], 10, 64)
	stepID, err := strconv.ParseUint(params["stepID"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	workflowStep := config.NewWorkflowStep()
	workflowStep.SetID(stepID)
	workflowStep.SetWorkflowID(workflowID)
	if err = restService.workflowStepService.Delete(session, workflowStep); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}
