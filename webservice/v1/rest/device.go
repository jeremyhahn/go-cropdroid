package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

var (
	ErrDeviceNotFound = errors.New("device not found")
)

type DeviceRestServicer interface {
	SetServiceRegistry(serviceRegistry service.ServiceRegistry)
	SetMiddleware(middleware middleware.JsonWebTokenMiddleware)
	View(w http.ResponseWriter, r *http.Request)
	State(w http.ResponseWriter, r *http.Request)
	Switch(w http.ResponseWriter, r *http.Request)
	Metric(w http.ResponseWriter, r *http.Request)
	TimerSwitch(w http.ResponseWriter, r *http.Request)
	History(w http.ResponseWriter, r *http.Request)
	RestService
}

type DeviceRestService struct {
	serviceRegistry service.ServiceRegistry
	middleware      middleware.JsonWebTokenMiddleware
	httpWriter      response.HttpWriter
	DeviceRestServicer
}

func NewDeviceRestService(
	serviceRegistry service.ServiceRegistry,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) DeviceRestServicer {

	return &DeviceRestService{
		serviceRegistry: serviceRegistry,
		middleware:      middleware,
		httpWriter:      httpWriter}
}

// Dependency injection to set mocked service registry
func (restService *DeviceRestService) SetServiceRegistry(serviceRegistry service.ServiceRegistry) {
	restService.serviceRegistry = serviceRegistry
}

// Dependency injection to set mocked JWT middleware
func (restService *DeviceRestService) SetMiddleware(middleware middleware.JsonWebTokenMiddleware) {
	restService.middleware = middleware
}

// Local helper to get the device service
func (restService *DeviceRestService) deviceService(r *http.Request) (service.DeviceServicer, error) {
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
		if service.DeviceType() == deviceType {
			return service, nil
		}
	}
	return nil, ErrDeviceNotFound
}

// Returns a device UI view that contains a complete device model with
// it's configuration and current state.
func (restService *DeviceRestService) View(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	deviceService, err := restService.deviceService(r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	view, err := deviceService.View()
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, view)
}

// Sets a device metric value using the "key" and "value" HTTP GET parmeters
func (restService *DeviceRestService) Metric(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	key := params["key"]
	value := params["value"]
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	deviceService, err := restService.deviceService(r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err = deviceService.SetMetricValue(key, floatValue); err != nil {
		session.GetLogger().Errorf("Error: Unable to set metric %s for %s device",
			key, deviceService.DeviceType())
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, floatValue)
}

// Returns the current device state
func (restService *DeviceRestService) State(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	deviceService, err := restService.deviceService(r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	state, err := deviceService.State()
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, state)
}

// Switch device channel on / off
func (restService *DeviceRestService) Switch(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	channel := params["channel"]
	position := params["position"]
	_channel, err := strconv.Atoi(channel)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	_position, err := strconv.Atoi(position)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	deviceService, err := restService.deviceService(r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	channelConfig, err := deviceService.ChannelConfig(_channel)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	deviceType := deviceService.DeviceType()
	switchPosition := util.NewSwitchPosition(_position)
	message := fmt.Sprintf("%s switching %s %s %s", session.GetUser().GetEmail(),
		switchPosition.ToLowerString(), deviceType, channelConfig.GetName())
	eventEntity, err := deviceService.Switch(_channel, _position, message)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, eventEntity)
}

// Switch device channel on / off for specified duration in seconds.
func (restService *DeviceRestService) TimerSwitch(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	channel := params["channel"]
	duration := params["duration"]
	_channel, err := strconv.Atoi(channel)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	_duration, err := strconv.Atoi(duration)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	deviceService, err := restService.deviceService(r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	deviceType := deviceService.DeviceType()
	message := fmt.Sprintf("%s switching on %s channel %s for %s seconds",
		session.GetUser().GetEmail(), deviceType, channel, duration)
	eventEntity, err := deviceService.TimerSwitch(_channel, _duration, message)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, eventEntity)
}

// Retrieve metric data history
func (restService *DeviceRestService) History(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	metric := params["metric"]
	deviceService, err := restService.deviceService(r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	history, err := deviceService.History(metric)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, history)
}
