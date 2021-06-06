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

type ScheduleRestService interface {
	GetSchedule(w http.ResponseWriter, r *http.Request)
	//GetSchedules(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultScheduleRestService struct {
	scheduleService service.ScheduleService
	middleware      service.Middleware
	jsonWriter      common.HttpWriter
	ScheduleRestService
}

func NewScheduleRestService(scheduleService service.ScheduleService, middleware service.Middleware,
	jsonWriter common.HttpWriter) ScheduleRestService {

	return &DefaultScheduleRestService{
		scheduleService: scheduleService,
		middleware:      middleware,
		jsonWriter:      jsonWriter}
}

func (restService *DefaultScheduleRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	scheduleEndpoint := fmt.Sprintf("%s/schedule", baseFarmURI)
	scheduleItemEndpoint := fmt.Sprintf("%s/{id}", scheduleEndpoint)
	getScheduleEndpoint := fmt.Sprintf("%s/channel/{channelID}", scheduleEndpoint)
	router.Handle(getScheduleEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetSchedule)),
	)).Methods("GET")
	/*
		router.Handle("/api/v1/schedule/controller/{controllerID}", negroni.New(
			negroni.HandlerFunc(server.middleware.Validate),
			negroni.Wrap(http.HandlerFunc(scheduleRestService.GetSchedules)),
		)).Methods("GET")*/
	router.Handle(scheduleEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Create)),
	)).Methods("POST")
	router.Handle(scheduleEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Update)),
	)).Methods("PUT")
	router.Handle(scheduleItemEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Delete)),
	)).Methods("DELETE")
	return []string{scheduleEndpoint, scheduleItemEndpoint, getScheduleEndpoint}
}

func (restService *DefaultScheduleRestService) GetSchedule(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	channelID, err := strconv.Atoi(params["channelID"])
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("[ScheduleRestService.GetSchedule] channelID=%d", channelID)

	schedule, err := restService.scheduleService.GetSchedule(session, channelID)
	if err != nil {
		session.GetLogger().Errorf("[ScheduleRestService.GetSchedule] Error: ", err)
		restService.jsonWriter.Error500(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, schedule)
}

/*
func (restService *DefaultScheduleRestService) GetSchedules(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	controllerID, err := strconv.Atoi(params["controllerID"])
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("[ScheduleRestService.GetSchedules] controllerID=%d", controllerID)

	schedules, err := restService.scheduleService.GetSchedules(session.GetUser(), controllerID)
	if err != nil {
		session.GetLogger().Errorf("[ScheduleRestService.GetSchedules] Error: ", err)
		restService.jsonWriter.Error500(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, schedules)
}
*/

func (restService *DefaultScheduleRestService) Create(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debug("[ScheduleRestService.Create] Decoding JSON request")

	var schedule config.Schedule
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&schedule); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("[ScheduleRestService.Create] schedule=%+v", schedule)

	persisted, err := restService.scheduleService.Create(session, &schedule)
	if err != nil {
		session.GetLogger().Errorf("[ScheduleRestService.Create] Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, persisted)
}

func (restService *DefaultScheduleRestService) Update(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debug("[ScheduleRestService.Update] Decoding JSON request")

	var schedule config.Schedule
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&schedule); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("[ScheduleRestService.Update] schedule=%+v", schedule)

	if err = restService.scheduleService.Update(session, &schedule); err != nil {
		session.GetLogger().Errorf("[ScheduleRestService.Update] Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}

func (restService *DefaultScheduleRestService) Delete(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("[ScheduleRestService.Delete] schedule.id=%d", id)

	if err = restService.scheduleService.Delete(session, &config.Schedule{ID: id}); err != nil {
		session.GetLogger().Errorf("[ScheduleRestService.Delete] Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
