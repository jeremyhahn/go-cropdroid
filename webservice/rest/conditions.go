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
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type ConditionRestService interface {
	GetListView(w http.ResponseWriter, r *http.Request)
	//GetConditions(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultConditionRestService struct {
	conditionService service.ConditionService
	conditionMapper  mapper.ConditionMapper
	middleware       service.Middleware
	jsonWriter       common.HttpWriter
	ConditionRestService
}

func NewConditionRestService(conditionService service.ConditionService, conditionMapper mapper.ConditionMapper,
	middleware service.Middleware, jsonWriter common.HttpWriter) ConditionRestService {

	return &DefaultConditionRestService{
		conditionService: conditionService,
		conditionMapper:  conditionMapper,
		middleware:       middleware,
		jsonWriter:       jsonWriter}
}

func (restService *DefaultConditionRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	conditionsEndpoint := fmt.Sprintf("%s/conditions", baseFarmURI)
	getConditionsEndpoint := fmt.Sprintf("%s/conditions/{id}", baseFarmURI)
	getChannelConditionsEndpoint := fmt.Sprintf("%s/channel/{channelID}", conditionsEndpoint)
	router.Handle(getChannelConditionsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetListView)),
	)).Methods("GET")
	/*
		router.Handle("/api/v1/conditions/channelID/{channelID}", negroni.New(
			negroni.HandlerFunc(restService.middleware.Validate),
			negroni.Wrap(http.HandlerFunc(conditionRestService.GetConditions)),
		)).Methods("GET")*/
	router.Handle(conditionsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Create)),
	)).Methods("POST")
	router.Handle(conditionsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Update)),
	)).Methods("PUT")
	router.Handle(getConditionsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Delete)),
	)).Methods("DELETE")
	return []string{conditionsEndpoint, getConditionsEndpoint, getChannelConditionsEndpoint}
}

func (restService *DefaultConditionRestService) GetListView(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	channelID, err := strconv.Atoi(params["channelID"])
	if err != nil {
		session.GetLogger().Errorf("Error parsing channelID. params=%s, error=%s", params, err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("channelID=%d", channelID)

	condition, err := restService.conditionService.GetListView(session, channelID)
	if err != nil {
		session.GetLogger().Errorf("Error: %s", err)
		restService.jsonWriter.Error500(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, condition)
}

// func (restService *DefaultConditionRestService) GetConditions(w http.ResponseWriter, r *http.Request) {

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

func (restService *DefaultConditionRestService) Create(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debug("Decoding JSON request")

	var condition config.Condition
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&condition); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("condition=%+v", condition)

	persisted, err := restService.conditionService.Create(session, &condition)
	if err != nil {
		session.GetLogger().Errorf("Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, persisted)
}

func (restService *DefaultConditionRestService) Update(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debug("Decoding JSON request")

	var condition viewmodel.Condition
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&condition); err != nil {
		session.GetLogger().Errorf("Error: %s", err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("condition=%+v", condition)

	conditionConfig := restService.conditionMapper.MapViewToConfig(condition)
	if err = restService.conditionService.Update(session, conditionConfig); err != nil {
		session.GetLogger().Errorf("Error: %s", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}

func (restService *DefaultConditionRestService) Delete(w http.ResponseWriter, r *http.Request) {

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

	session.GetLogger().Debugf("condition.id=%d", id)

	if err = restService.conditionService.Delete(session, &config.Condition{ID: id}); err != nil {
		session.GetLogger().Errorf("Error: %s", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
