package router

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/pki/ca"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

type FarmRouter struct {
	baseFarmURI              string
	middleware               middleware.JsonWebTokenMiddleware
	farmWebSocketRestService rest.FarmWebSocketRestServicer
	farmRestService          rest.FarmRestServicer
	responseWriter           response.ResponseWriter
	WebServiceRouter
}

// Creates a new web service authentication router
func NewFarmRouter(
	domain string,
	certificateAuthority ca.CertificateAuthority,
	baseFarmURI string,
	serviceRegistry service.ServiceRegistry,
	middleware middleware.JsonWebTokenMiddleware,
	farmWebSocketRestService rest.FarmWebSocketRestServicer,
	responseWriter response.HttpWriter) WebServiceRouter {

	return &FarmRouter{
		middleware:               middleware,
		farmWebSocketRestService: farmWebSocketRestService,
		farmRestService: rest.NewFarmRestService(
			domain,
			certificateAuthority,
			serviceRegistry.GetFarmFactory(),
			serviceRegistry.GetUserService(),
			serviceRegistry.GetNotificationService(),
			middleware,
			responseWriter)}
}

// Registers all of the authentication endpoints at the root of the webservice (/api/v1)
func (farmRouter *FarmRouter) RegisterRoutes(router *mux.Router, baseURI string) []string {
	farmRouter.baseFarmURI = fmt.Sprintf("%s/farms/{farmID}", baseURI)
	return []string{
		farmRouter.farms(router, baseURI),
		farmRouter.farmTicker(router, baseURI),
		farmRouter.devices(router, farmRouter.baseFarmURI),
		farmRouter.users(router, farmRouter.baseFarmURI),
		farmRouter.resetPassword(router, farmRouter.baseFarmURI),
		farmRouter.deleteFarmUser(router, farmRouter.baseFarmURI),
		farmRouter.config(router, farmRouter.baseFarmURI),
		farmRouter.state(router, farmRouter.baseFarmURI),
		farmRouter.pubkey(router, farmRouter.baseFarmURI),
		farmRouter.setDeviceConfig(router, farmRouter.baseFarmURI),
		farmRouter.sendNotification(router, farmRouter.baseFarmURI),
		farmRouter.notificationTicker(router, farmRouter.baseFarmURI)}
}

// @Summary List farms
// @Description Returns a paginated list of farm entites
// @Tags Farms
// @Produce json
// @Consume json
// @Success 200 {object} config.Farm
// @Failure 400 {object} response.WebServiceResponse
// @Failure 401 {object} response.WebServiceResponse
// @Router /farms [get]
// @Security JWT
func (farmRouter *FarmRouter) farms(router *mux.Router, baseURI string) string {
	endpoint := fmt.Sprintf("%s/farms", baseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.Farms)),
	)).Methods("GET")
	return endpoint
}

// @Summary List farm devices
// @Description Returns all devices associated with the farm
// @Tags Devices
// @Accept json
// @Produce json
// @Param   farmID	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farms/{farmID}/devices [get]
// @Security JWT
func (farmRouter *FarmRouter) devices(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/devices", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.Devices)),
	))
	return endpoint
}

// @Summary List farm user membership
// @Description Returns all users who have are members and able to access the farm.
// @Tags Farms
// @Produce json
// @Consume json
// @Param   farmID	path	integer	true	"string valid"
// @Success 200 {object} config.Farm
// @Failure 400 {object} response.WebServiceResponse
// @Failure 401 {object} response.WebServiceResponse
// @Router /farms/{farmID}/users [get]
// @Security JWT
func (farmRouter *FarmRouter) users(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/users", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.FarmUsers)),
	)).Methods("GET")
	return endpoint
}

// @Summary Reset user password
// @Description Updates a users password
// @Tags Farms
// @Produce json
// @Consume json
// @Param 	farmID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Param 	userID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Param UserCredentials body service.UserCredentials true "UserCredentials struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Failure 401 {object} response.WebServiceResponse
// @Failure 500 {object} response.WebServiceResponse
// @Router /farms/{farmID}/users/{userID} [post]
// @Security JWT
func (farmRouter *FarmRouter) resetPassword(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/users/{userID}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.ResetPassword)),
	)).Methods("POST")
	return endpoint
}

// @Summary Disassociate a user
// @Description Deletes the users permission to access the farm
// @Tags Farms
// @Produce json
// @Consume json
// @Param 	farmID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Param 	userID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200 {object} config.Farm
// @Failure 400 {object} response.WebServiceResponse
// @Failure 401 {object} response.WebServiceResponse
// @Router /farms/{farmID}/users/{userID} [delete]
// @Security JWT
func (farmRouter *FarmRouter) deleteFarmUser(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/users/{userID}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.DeleteFarmUser)),
	)).Methods("DELETE")
	return endpoint
}

// @Summary Get farm configuration
// @Description Returns the complete farm configuration.
// @Tags Farms
// @Produce json
// @Consume json
// @Param 	farmID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200 {object} config.Farm
// @Failure 400 {object} response.WebServiceResponse
// @Failure 401 {object} response.WebServiceResponse
// @Router /farms/{farmID}/config [get]
// @Security JWT
func (farmRouter *FarmRouter) config(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/config", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.Config)),
	)).Methods("GET")
	return endpoint
}

// @Summary Get the farm state
// @Description Returns the complete farm state
// @Tags Farms
// @Produce json
// @Consume json
// @Param 	farmID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Param 	userID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200 {object} config.Farm
// @Failure 400 {object} response.WebServiceResponse
// @Failure 401 {object} response.WebServiceResponse
// @Router /farms/{farmID}/users/{userID} [get]
// @Security JWT
func (farmRouter *FarmRouter) state(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/state", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.State)),
	)).Methods("GET")
	return endpoint
}

// @Summary Returns the farm public key
// @Description Returns the unique public key for the farm.
// @Tags Farms
// @Produce json
// @Consume json
// @Param 	farmID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200 {string} pubkey
// @Router /farms/{farmID}/pubkey [get]
func (farmRouter *FarmRouter) pubkey(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/pubkey", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.PublicKey)),
	)).Methods("GET")
	return endpoint
}

// @Summary Set device configuration
// @Description Sets a farm device configuration
// @Tags Farms
// @Produce json
// @Consume json
// @Param 	farmID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Param 	deviceID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Param   key	query	string	false	"string valid"	minlength(1)	maxlength(255)
// @Success 200 {string} pubkey
// @Failure 400 {object} response.WebServiceResponse
// @Failure 401 {object} response.WebServiceResponse
// @Router /farms/{farmID}/config/{deviceID}/{key} [get]
// @Security JWT
func (farmRouter *FarmRouter) setDeviceConfig(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/config/{deviceID}/{key}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.State)),
	)).Methods("GET")
	return endpoint
}

// @Summary Sends a push notification
// @Description Sends a push notification to all clients connected to the farm websocket
// @Tags Farms
// @Produce  json
// @Param   farmID	path	integer	true	"string valid"
// @Param	type	path	string	true	"Message type"
// @Param	message	path	string	true	"Message text"
// @Param	priority	path	int	false	"Message priority"
// @Success 200
// @Router /farms/{farmID}/notification/{type}/{message}/{priority} [get]
func (farmRouter *FarmRouter) sendNotification(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/notification/{type}/{message}/{priority}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmRestService.SendMessage)),
	)).Methods("GET")
	return endpoint
}

// @Summary Stream real-time farm configuration and state updates via websocket
// @Description Create a websocket and start receiving real-time farm updates
// @Tags Farms
// @Produce  json
// @Param farmID	path	integer		true	"string valid"
// @Param user 		body	model.User	true	"user model"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Failure 401 {object} response.WebServiceResponse
// @Router /farmticker/{farmID} [get]
// @Security JWT
func (farmRouter *FarmRouter) farmTicker(router *mux.Router, baseURI string) string {
	endpoint := fmt.Sprintf("%s/farmticker/{farmID}", baseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmWebSocketRestService.FarmTickerConnect)),
	))
	return endpoint
}

// @Summary Stream real-time farm push notifications via websocket
// @Description Stream real-time farm push notifications via websocket
// @Tags Farms
// @Produce  json
// @Param farmID	path	integer		true	"string valid"
// @Param user 		body	model.User	true	"user model"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Failure 401 {object} response.WebServiceResponse
// @Router /farms/{farmID}/notifications [get]
// @Security JWT
func (farmRouter *FarmRouter) notificationTicker(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/notifications", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(farmRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(farmRouter.farmWebSocketRestService.PushNotificationConnect)),
	))
	return endpoint
}
