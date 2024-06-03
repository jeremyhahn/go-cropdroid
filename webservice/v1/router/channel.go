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

type ChannelRouter struct {
	middleware         middleware.JsonWebTokenMiddleware
	channelRestService rest.ChannelRestServicer
	WebServiceRouter
}

// Creates a new web service channel router
func NewChannelRouter(
	channelService service.ChannelServicer,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &ChannelRouter{
		middleware: middleware,
		channelRestService: rest.NewChannelRestService(
			channelService,
			middleware,
			httpWriter)}
}

// Registers all device channel endpoints
func (channelRouter *ChannelRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	return []string{
		channelRouter.list(router, baseFarmURI),
		channelRouter.update(router, baseFarmURI)}
}

// @Summary List device channels
// @Description Returns all channels for the requested farm
// @Tags Channels
// @Accept json
// @Produce  json
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Param   farmID		path	integer	true	"string valid"
// @Param   deviceID	path	integer	false	"string valid"
// @Router /farms/{farmID}/channels/{deviceID} [get]
// @Security JWT
func (channelRouter *ChannelRouter) list(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/channels", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(channelRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(channelRouter.channelRestService.List)),
	))
	return endpoint
}

// @Summary Update device channel
// @Description Updates a farm channel
// @Tags Channels
// @Accept json
// @Produce  json
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Param   farmID	path	integer	true	"string valid"
// @Param   id		path	integer	true	"string valid"
// @Router /farms/{farmID}/channel/{id} [put]
// @Security JWT
func (channelRouter *ChannelRouter) update(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/channel/{id}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(channelRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(channelRouter.channelRestService.Update)),
	))
	return endpoint
}
