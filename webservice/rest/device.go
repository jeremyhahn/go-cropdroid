package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
)

var (
	ErrControllerNotFound = errors.New("Controller not found")
)

type ControllerRestService interface {
	View(w http.ResponseWriter, r *http.Request)
	State(w http.ResponseWriter, r *http.Request)
	Switch(w http.ResponseWriter, r *http.Request)
	TimerSwitch(w http.ResponseWriter, r *http.Request)
	History(w http.ResponseWriter, r *http.Request)
	RestService
}

type MicroControllerRestService struct {
	serviceRegistry service.ServiceRegistry
	middleware      service.Middleware
	jsonWriter      common.HttpWriter
	ControllerRestService
}

func NewControllerRestService(serviceRegistry service.ServiceRegistry,
	middleware service.Middleware, jsonWriter common.HttpWriter) ControllerRestService {
	return &MicroControllerRestService{
		serviceRegistry: serviceRegistry,
		middleware:      middleware,
		jsonWriter:      jsonWriter}
}

func (restService *MicroControllerRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	// /farms/{farmID}/devices/{controller_type}
	endpoint := fmt.Sprintf("%s/devices/{controllerType}", baseFarmURI)
	// /farms/{farmID}/devices/{controller_type}/view
	viewEndpoint := fmt.Sprintf("%s/view", endpoint)
	// /farm/{farmID}/{devices/controller_type}/metrics/{key}/{value}
	metricEndpoint := fmt.Sprintf("%s/metrics/{key}/{value}", endpoint)
	// /farms/{farmID}/devices/{controller_type}/history/metric
	historyEndpoint := fmt.Sprintf("%s/history/{metric}", endpoint)
	// /farms/{farmID}/devices/{controller_type}/switch/{channel}/{postion}
	switchEndpoint := fmt.Sprintf("%s/switch/{channel}/{position}", endpoint)
	// /farms/{farmID}/devices/{controller_type}/timerSwitch/{channel}/{duration}
	timerSwitchEndpoint := fmt.Sprintf("%s/timerSwitch/{channel}/{duration}", endpoint)
	router.Handle(viewEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.View)),
	))
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.State)),
	))
	router.Handle(metricEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Metric)),
	))
	router.Handle(historyEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.History)),
	))
	router.Handle(switchEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Switch)),
	))
	router.Handle(timerSwitchEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.TimerSwitch)),
	))
	return []string{endpoint, viewEndpoint, metricEndpoint, historyEndpoint, switchEndpoint, timerSwitchEndpoint}
}

func (restService *MicroControllerRestService) getControllerService(r *http.Request) (common.ControllerService, error) {
	params := mux.Vars(r)
	controllerType := params["controllerType"]
	farmID, err := strconv.Atoi(params["farmID"])
	if err != nil {
		return nil, err
	}
	services, err := restService.serviceRegistry.GetControllerServices(farmID)
	if err != nil {
		return nil, err
	}
	for _, service := range services {
		if service.GetControllerType() == controllerType {
			return service, nil
		}
	}
	return nil, ErrControllerNotFound
}

func (restService *MicroControllerRestService) View(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer ctx.Close()

	ctx.GetLogger().Debugf("[MicroControllerRestService.Status] REST service /view request email=%s", ctx.GetUser().GetEmail())

	controllerService, err := restService.getControllerService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	view, err := controllerService.GetView()
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, view)
}

func (restService *MicroControllerRestService) Metric(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer ctx.Close()

	ctx.GetLogger().Debugf("[MicroControllerRestService.Metric] REST service /metric request from %s", ctx.GetUser().GetEmail())

	params := mux.Vars(r)
	key := params["key"]
	value := params["value"]

	ctx.GetLogger().Debugf("[MicroControllerRestService.Metric] /%s/%s", key, value)

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		ctx.GetLogger().Errorf("[MicroControllerRestService.Metric] Error: %s", err)
	}

	controllerService, err := restService.getControllerService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	if err = controllerService.SetMetricValue(key, floatValue); err != nil {
		ctx.GetLogger().Errorf("[MicroControllerRestService.Metric] Error: Unable to set metric %s for %s controller",
			key, controllerService.GetControllerType())
		BadRequestError(w, r, err, restService.jsonWriter)
	}

	restService.jsonWriter.Write(w, http.StatusOK, floatValue)
}

func (restService *MicroControllerRestService) State(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer ctx.Close()

	ctx.GetLogger().Debugf("[MicroControllerRestService.Status] REST service /state request email=%s", ctx.GetUser().GetEmail())

	controllerService, err := restService.getControllerService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	state, err := controllerService.GetState()
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	restService.jsonWriter.Write(w, http.StatusOK, state)
}

func (restService *MicroControllerRestService) Switch(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
	}
	defer ctx.Close()

	params := mux.Vars(r)
	channel := params["channel"]
	position := params["position"]

	ctx.GetLogger().Debugf("[MicroControllerRestService.Switch] REST service /switch request channel=%s, position=%s, user=%s", channel, position, ctx.GetUser().GetEmail())

	_channel, err := strconv.Atoi(channel)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	_position, err := strconv.Atoi(position)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	controllerService, err := restService.getControllerService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	controllerType := controllerService.GetControllerType()
	message := fmt.Sprintf("User %s switching on %s channel %s", ctx.GetUser().GetEmail(), controllerType, channel)
	eventEntity, err := controllerService.Switch(_channel, _position, message)
	if err != nil {
		ctx.GetLogger().Error("[MicroControllerRestService.Switch] Error: %s", err.Error())
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, eventEntity)
}

func (restService *MicroControllerRestService) TimerSwitch(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
	}
	defer ctx.Close()

	params := mux.Vars(r)
	channel := params["channel"]
	duration := params["duration"]

	ctx.GetLogger().Debugf("[MicroControllerRestService.Switch] REST service /timerSwitch request channel=%s, duration=%s, user=%s", channel, duration, ctx.GetUser().GetEmail())

	_channel, err := strconv.Atoi(channel)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	_duration, err := strconv.Atoi(duration)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	controllerService, err := restService.getControllerService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	controllerType := controllerService.GetControllerType()
	message := fmt.Sprintf("User %s switching on %s channel %s for %s seconds",
		ctx.GetUser().GetEmail(), controllerType, channel, duration)
	eventEntity, err := controllerService.TimerSwitch(_channel, _duration, message)
	if err != nil {
		ctx.GetLogger().Error("[MicroControllerRestService.TimerSwitch] Error: %s", err.Error())
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, eventEntity)
}

func (restService *MicroControllerRestService) History(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
	}
	defer ctx.Close()

	params := mux.Vars(r)
	metric := params["metric"]

	ctx.GetLogger().Debugf("[MicroControllerRestService.History] REST service /history request. metric=%s", metric)

	controllerService, err := restService.getControllerService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	history, err := controllerService.GetHistory(metric)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
	}

	restService.jsonWriter.Write(w, http.StatusOK, history)
}

/*
func (restService *MicroControllerRestService) SetMetric(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
	}
	defer ctx.Close()

	params := mux.Vars(r)
	metric := params["metric"]
	value := params["value"]

	fvalue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		ctx.GetLogger().Errorf("[MicroControllerRestService.SetMetric] Error: %s", err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	ctx.GetLogger().Debugf("[MicroControllerRestService.SetMetric] REST service / request. metric=%s, value=%s", metric, value)

	controllerService, err := restService.getControllerService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	err = controllerService.SetMetricValue(metric, fvalue)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
*/
