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

type ScheduleRouter struct {
	middleware          middleware.JsonWebTokenMiddleware
	scheduleRestService rest.ScheduleRestServicer
	WebServiceRouter
}

// Creates a new web service schedule router
func NewScheduleRouter(
	scheduleService service.ScheduleService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &ScheduleRouter{
		middleware: middleware,
		scheduleRestService: rest.NewScheduleRestService(
			scheduleService,
			middleware,
			httpWriter)}
}

// Registers all of the schedule endpoints at the root of the farm (/api/v1/farm/{farmID})
func (scheduleRouter *ScheduleRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	return []string{
		scheduleRouter.get(router, baseFarmURI),
		scheduleRouter.create(router, baseFarmURI),
		scheduleRouter.update(router, baseFarmURI),
		scheduleRouter.delete(router, baseFarmURI)}
}

// @Summary Get channel schedule
// @Description Returns the requested channel schedule
// @Tags Farms
// @Produce  json
// @Param   farmID		path	integer	true	"string valid"
// @Param   channelID	path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farm/{farmID}/schedule/channel/{channelID} [get]
// @Security JWT
func (scheduleRouter *ScheduleRouter) get(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/schedule", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(scheduleRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(scheduleRouter.scheduleRestService.GetSchedule)),
	))
	return endpoint
}

// @Summary Create channel schedule
// @Description Creates a new channel schedule
// @Tags Farms
// @Produce  json
// @Param   farmID		path	integer	true	"string valid"
// @Param   Schedule	body	config.Schedule	true	"config.Schedule struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farm/{farmID}/schedule [post]
// @Security JWT
func (scheduleRouter *ScheduleRouter) create(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/schedule", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(scheduleRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(scheduleRouter.scheduleRestService.Create)),
	)).Methods("POST")
	return endpoint
}

// @Summary Update channel schedule
// @Description Updates a new channel schedule
// @Tags Farms
// @Produce  json
// @Param   farmID		path	integer			true	"string valid"
// @Param   Schedule	body	config.Schedule	true	"config.Schedule struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farm/{farmID}/schedule [put]
// @Security JWT
func (scheduleRouter *ScheduleRouter) update(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/schedule", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(scheduleRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(scheduleRouter.scheduleRestService.Update)),
	)).Methods("PUT")
	return endpoint
}

// @Summary Delete channel schedule
// @Description Deletes a channel schedule
// @Tags Farms
// @Produce  json
// @Param   farmID	path	integer	true	"string valid"
// @Param   id		path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /farm/{farmID}/schedule/{id} [delete]
// @Security JWT
func (scheduleRouter *ScheduleRouter) delete(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/schedule/{id}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(scheduleRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(scheduleRouter.scheduleRestService.Delete)),
	)).Methods("DELETE")
	return endpoint
}
