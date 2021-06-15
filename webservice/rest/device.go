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
	ErrDeviceNotFound = errors.New("device not found")
)

type DeviceRestService interface {
	View(w http.ResponseWriter, r *http.Request)
	State(w http.ResponseWriter, r *http.Request)
	Switch(w http.ResponseWriter, r *http.Request)
	TimerSwitch(w http.ResponseWriter, r *http.Request)
	History(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultDeviceRestService struct {
	serviceRegistry service.ServiceRegistry
	middleware      service.Middleware
	jsonWriter      common.HttpWriter
	DeviceRestService
}

func NewDeviceRestService(serviceRegistry service.ServiceRegistry,
	middleware service.Middleware, jsonWriter common.HttpWriter) DeviceRestService {
	return &DefaultDeviceRestService{
		serviceRegistry: serviceRegistry,
		middleware:      middleware,
		jsonWriter:      jsonWriter}
}

func (restService *DefaultDeviceRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	// /farms/{farmID}/devices/{device_type}
	endpoint := fmt.Sprintf("%s/devices/{deviceType}", baseFarmURI)
	// /farms/{farmID}/devices/{device_type}/view
	viewEndpoint := fmt.Sprintf("%s/view", endpoint)
	// /farm/{farmID}/{devices/device_type}/metrics/{key}/{value}
	metricEndpoint := fmt.Sprintf("%s/metrics/{key}/{value}", endpoint)
	// /farms/{farmID}/devices/{device_type}/history/metric
	historyEndpoint := fmt.Sprintf("%s/history/{metric}", endpoint)
	// /farms/{farmID}/devices/{device_type}/switch/{channel}/{postion}
	switchEndpoint := fmt.Sprintf("%s/switch/{channel}/{position}", endpoint)
	// /farms/{farmID}/devices/{device_type}/timerSwitch/{channel}/{duration}
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

func (restService *DefaultDeviceRestService) getDeviceService(r *http.Request) (service.DeviceService, error) {
	params := mux.Vars(r)
	deviceType := params["deviceType"]
	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
	if err != nil {
		return nil, err
	}
	services, err := restService.serviceRegistry.GetDeviceServices(farmID)
	if err != nil {
		return nil, err
	}
	for _, service := range services {
		if service.GetDeviceType() == deviceType {
			return service, nil
		}
	}
	return nil, ErrDeviceNotFound
}

func (restService *DefaultDeviceRestService) View(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer ctx.Close()

	ctx.GetLogger().Debugf("REST service /view request email=%s", ctx.GetUser().GetEmail())

	deviceService, err := restService.getDeviceService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	view, err := deviceService.GetView()
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, view)
}

func (restService *DefaultDeviceRestService) Metric(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer ctx.Close()

	ctx.GetLogger().Debugf("REST service /metric request from %s", ctx.GetUser().GetEmail())

	params := mux.Vars(r)
	key := params["key"]
	value := params["value"]

	ctx.GetLogger().Debugf("/%s/%s", key, value)

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		ctx.GetLogger().Errorf("Error: %s", err)
	}

	deviceService, err := restService.getDeviceService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	if err = deviceService.SetMetricValue(key, floatValue); err != nil {
		ctx.GetLogger().Errorf("Error: Unable to set metric %s for %s device",
			key, deviceService.GetDeviceType())
		BadRequestError(w, r, err, restService.jsonWriter)
	}

	restService.jsonWriter.Write(w, http.StatusOK, floatValue)
}

func (restService *DefaultDeviceRestService) State(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer ctx.Close()

	ctx.GetLogger().Debugf("REST service /state request email=%s", ctx.GetUser().GetEmail())

	deviceService, err := restService.getDeviceService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	state, err := deviceService.GetState()
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	restService.jsonWriter.Write(w, http.StatusOK, state)
}

func (restService *DefaultDeviceRestService) Switch(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
	}
	defer ctx.Close()

	params := mux.Vars(r)
	channel := params["channel"]
	position := params["position"]

	ctx.GetLogger().Debugf("REST service /switch request channel=%s, position=%s, user=%s", channel, position, ctx.GetUser().GetEmail())

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

	deviceService, err := restService.getDeviceService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	deviceType := deviceService.GetDeviceType()
	message := fmt.Sprintf("User %s switching on %s channel %s", ctx.GetUser().GetEmail(), deviceType, channel)
	eventEntity, err := deviceService.Switch(_channel, _position, message)
	if err != nil {
		ctx.GetLogger().Error("Error: %s", err.Error())
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, eventEntity)
}

func (restService *DefaultDeviceRestService) TimerSwitch(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
	}
	defer ctx.Close()

	params := mux.Vars(r)
	channel := params["channel"]
	duration := params["duration"]

	ctx.GetLogger().Debugf("REST service /timerSwitch request channel=%s, duration=%s, user=%s", channel, duration, ctx.GetUser().GetEmail())

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

	deviceService, err := restService.getDeviceService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	deviceType := deviceService.GetDeviceType()
	message := fmt.Sprintf("User %s switching on %s channel %s for %s seconds",
		ctx.GetUser().GetEmail(), deviceType, channel, duration)
	eventEntity, err := deviceService.TimerSwitch(_channel, _duration, message)
	if err != nil {
		ctx.GetLogger().Error("Error: %s", err.Error())
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, eventEntity)
}

func (restService *DefaultDeviceRestService) History(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
	}
	defer ctx.Close()

	params := mux.Vars(r)
	metric := params["metric"]

	ctx.GetLogger().Debugf("REST service /history request. metric=%s", metric)

	deviceService, err := restService.getDeviceService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	history, err := deviceService.GetHistory(metric)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
	}

	restService.jsonWriter.Write(w, http.StatusOK, history)
}

/*
func (restService *DefaultDeviceRestService) SetMetric(w http.ResponseWriter, r *http.Request) {

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
		ctx.GetLogger().Errorf("Error: %s", err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	ctx.GetLogger().Debugf("REST service / request. metric=%s, value=%s", metric, value)

	deviceService, err := restService.getDeviceService(r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	err = deviceService.SetMetricValue(metric, fvalue)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
*/
