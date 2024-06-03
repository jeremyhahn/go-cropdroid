package router

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

type DeviceRouter struct {
	middleware        middleware.JsonWebTokenMiddleware
	deviceRestService rest.DeviceRestServicer
	WebServiceRouter
}

func NewDeviceRouter(
	serviceRegistry service.ServiceRegistry,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &DeviceRouter{
		middleware: middleware,
		deviceRestService: rest.NewDeviceRestService(
			serviceRegistry,
			middleware,
			httpWriter)}
}

// Create a new farm device router and register all of the device endpoints
func (authenticationRouter *DeviceRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	return []string{
		authenticationRouter.view(router, baseFarmURI),
		authenticationRouter.state(router, baseFarmURI),
		authenticationRouter.metric(router, baseFarmURI),
		authenticationRouter.history(router, baseFarmURI),
		authenticationRouter._switch(router, baseFarmURI),
		authenticationRouter.timerSwitch(router, baseFarmURI)}
}

// @Summary Returns a UI view of the device
// @Description Returns a user-interface view of the device
// @Tags Devices
// @Accept json
// @Produce json
// @Param   farmID		path	integer	true	"string valid"
// @Param   deviceType	path	string	true	"string valid"	minlength(1)	maxlength(255)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/devices/{deviceType}/view [get]
// @Security JWT
func (deviceRouter *DeviceRouter) view(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/devices/{deviceType}/view", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(deviceRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(deviceRouter.deviceRestService.View)),
	))
	return endpoint
}

// @Summary Get current device state
// @Description Returns the current device state
// @Tags Devices
// @Accept json
// @Produce json
// @Param   farmID		path	integer	true	"string valid"
// @Param   deviceType	path	string	true	"string valid"	minlength(1)	maxlength(255)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/devices/{deviceType} [get]
// @Security JWT
func (deviceRouter *DeviceRouter) state(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/devices/{deviceType}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(deviceRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(deviceRouter.deviceRestService.State)),
	))
	return endpoint
}

// @Summary Sets metric value
// @Description Sets the value of a metric
// @Tags Devices
// @Accept json
// @Produce json
// @Param   farmID		path	integer	true	"string valid"
// @Param   deviceType	path	string	true	"string valid"	minlength(1)	maxlength(255)
// @Param   key			path	string	true	"string valid"	minlength(1)	maxlength(255)
// @Param   value		path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/devices/{deviceType}/metrics/{key}/{value} [get]
// @Security JWT
func (deviceRouter *DeviceRouter) metric(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/devices/{deviceType}/metrics/{key}/{value}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(deviceRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(deviceRouter.deviceRestService.Metric)),
	))
	return endpoint
}

// @Summary Get metric history
// @Description Returns a historical data set for the requested metric
// @Tags Devices
// @Accept json
// @Produce  json
// @Param   farmID		path	integer	true	"string valid"
// @Param   deviceType	path	string	true	"string valid"	minlength(1)	maxlength(255)
// @Param   metric		path	string	true	"string valid"	minlength(1)	maxlength(255)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/devices/{deviceType}/history/{metric} [get]
// @Security JWT
func (deviceRouter *DeviceRouter) history(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/devices/{deviceType}/history/{metric}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(deviceRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(deviceRouter.deviceRestService.History)),
	))
	return endpoint
}

// @Summary Switch channel on / off
// @Description Switch a farm channel on / off
// @Tags Devices
// @Accept json
// @Produce  json
// @Param   farmID		path	integer	true	"string valid"
// @Param   deviceType	path	string	true	"string valid"	minlength(1)	maxlength(255)
// @Param   channel		path	integer	true	"string valid"
// @Param   position	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/devices/{deviceType}/switch/{channel}/{position} [get]
// @Security JWT
func (deviceRouter *DeviceRouter) _switch(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/devices/{deviceType}/switch/{channel}/{position}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(deviceRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(deviceRouter.deviceRestService.Switch)),
	))
	return endpoint
}

// @Summary Switch channel for precise number of seconds
// @Description Switches the specified channel on for the specified number of seconds
// @Tags Devices
// @Accept json
// @Produce  json
// @Param   farmID		path	integer	true	"string valid"
// @Param   deviceType	path	string	true	"string valid"	minlength(1)	maxlength(255)
// @Param   channel		path	integer	true	"string valid"
// @Param   duration	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/devices/{deviceType}/timerSwitch/{channel}/{duration} [get]
// @Security JWT
func (deviceRouter *DeviceRouter) timerSwitch(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/devices/{deviceType}/timerSwitch/{channel}/{duration} ", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(deviceRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(deviceRouter.deviceRestService.TimerSwitch)),
	))
	return endpoint
}
