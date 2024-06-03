package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type ConditionRestServicer interface {
	SetService(conditionService service.ConditionServicer)
	SetMiddleware(middleware middleware.JsonWebTokenMiddleware)
	ListView(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	RestService
}

type ConditionRestService struct {
	conditionMapper  mapper.ConditionMapper
	conditionService service.ConditionServicer
	middleware       middleware.JsonWebTokenMiddleware
	httpWriter       response.HttpWriter
	ConditionRestServicer
}

func NewConditionRestService(
	conditionMapper mapper.ConditionMapper,
	conditionService service.ConditionServicer,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) ConditionRestServicer {

	return &ConditionRestService{
		conditionMapper:  conditionMapper,
		conditionService: conditionService,
		middleware:       middleware,
		httpWriter:       httpWriter}
}

// Dependency injection to set mocked condition service
func (restService *ConditionRestService) SetService(conditionService service.ConditionServicer) {
	restService.conditionService = conditionService
}

// Dependency injection to set mocked JWT middleware
func (restService *ConditionRestService) SetMiddleware(middleware middleware.JsonWebTokenMiddleware) {
	restService.middleware = middleware
}

// Returns a user friendly list of device channel conditions
func (restService *ConditionRestService) ListView(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	channelID, err := strconv.ParseUint(params["channelID"], 10, 64)
	if err != nil {
		session.GetLogger().Errorf("Error parsing channelID. params=%s, error=%s", params, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}
	condition, err := restService.conditionService.ListView(session, channelID)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, condition)
}

// Create a new device channel condition
func (restService *ConditionRestService) Create(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var condition config.ConditionStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&condition); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	persisted, err := restService.conditionService.Create(session, &condition)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, persisted)
}

// Update a device channel condition
func (restService *ConditionRestService) Update(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var condition config.ConditionStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&condition); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err = restService.conditionService.Update(session, &condition); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

// Delete a device channel condition
func (restService *ConditionRestService) Delete(w http.ResponseWriter, r *http.Request) {
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
	if err = restService.conditionService.Delete(session, &config.ConditionStruct{ID: id}); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

// func (restService *ConditionRestService) GetConditions(w http.ResponseWriter, r *http.Request) {

// 	session, err := restService.middleware.CreateSession(w, r)
// 	if err != nil {
// 		BadRequestError(w, r, err, restService.jsonWriter)
// 		return
// 	}
// 	defer session.Close()

// 	params := mux.Vars(r)
// 	deviceID, err := strconv.Atoi(params["deviceID"])
// 	if err != nil {
// 		BadRequestError(w, r, err, restService.jsonWriter)
// 		return
// 	}

// 	session.GetLogger().Debugf("deviceID=%d", deviceID)

// 	conditions, err := restService.conditionService.GetConditions(session, deviceID)
// 	if err != nil {
// 		session.GetLogger().Errorf("Error: ", err)
// 		restService.jsonWriter.Error500(w, err)
// 		return
// 	}

// 	restService.jsonWriter.Write(w, http.StatusOK, conditions)
// }
