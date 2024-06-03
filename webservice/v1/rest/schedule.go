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

type ScheduleRestServicer interface {
	GetSchedule(w http.ResponseWriter, r *http.Request)
	//GetSchedules(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	RestService
}

type ScheduleRestService struct {
	scheduleService service.ScheduleService
	middleware      middleware.JsonWebTokenMiddleware
	httpWriter      response.HttpWriter
	ScheduleRestServicer
}

func NewScheduleRestService(
	scheduleService service.ScheduleService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) ScheduleRestServicer {

	return &ScheduleRestService{
		scheduleService: scheduleService,
		middleware:      middleware,
		httpWriter:      httpWriter}
}

func (restService *ScheduleRestService) GetSchedule(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	channelID, err := strconv.ParseUint(params["channelID"], 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	schedule, err := restService.scheduleService.GetSchedule(session, channelID)
	if err != nil {
		restService.httpWriter.Error500(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, schedule)
}

func (restService *ScheduleRestService) Create(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var schedule config.Schedule
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&schedule); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	persisted, err := restService.scheduleService.Create(session, schedule)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, persisted)
}

func (restService *ScheduleRestService) Update(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var schedule config.Schedule
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&schedule); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err = restService.scheduleService.Update(session, schedule); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

func (restService *ScheduleRestService) Delete(w http.ResponseWriter, r *http.Request) {
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
	if err = restService.scheduleService.Delete(session, &config.ScheduleStruct{ID: id}); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

// func (restService *ScheduleRestService) GetSchedules(w http.ResponseWriter, r *http.Request) {

// 	session, err := restService.JsonWebTokenMiddleware.CreateSession(w, r)
// 	if err != nil {
// 		BadRequestError(w, r, err, restService.httpWriter)
// 		return
// 	}
// 	defer session.Close()

// 	params := mux.Vars(r)
// 	deviceID, err := strconv.Atoi(params["deviceID"])
// 	if err != nil {
// 		BadRequestError(w, r, err, restService.httpWriter)
// 		return
// 	}

// 	session.GetLogger().Debugf("deviceID=%d", deviceID)

// 	schedules, err := restService.scheduleService.GetSchedules(session.GetUser(), deviceID)
// 	if err != nil {
// 		session.GetLogger().Errorf("Error: ", err)
// 		restService.httpWriter.Error500(w, err)
// 		return
// 	}

// 	restService.httpWriter.Write(w, http.StatusOK, schedules)
// }
