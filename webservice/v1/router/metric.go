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
	"github.com/op/go-logging"
)

type MetricRouter struct {
	middleware        middleware.JsonWebTokenMiddleware
	metricRestService rest.MetricRestServicer
	WebServiceRouter
}

// Creates a new web service metric router
func NewMetricRouter(
	logger *logging.Logger,
	metricService service.MetricService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &MetricRouter{
		middleware: middleware,
		metricRestService: rest.NewMetricRestService(
			logger,
			metricService,
			middleware,
			httpWriter)}
}

// Registers all of the metric endpoints at the root of the farm (/api/v1/farms/{farmID})
func (metricRouter *MetricRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	return []string{
		metricRouter.set(router, baseFarmURI),
		metricRouter.getByDeviceID(router, baseFarmURI)}
}

// @Summary List metrics
// @Description Returns a page of metric entries
// @Tags Metric
// @Produce  json
// @Param   page	path	string	false	"string valid"	minlength(1)	maxlength(20)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /metrics [put]
// @Security JWT
func (metricRouter *MetricRouter) set(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/metrics", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(metricRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(metricRouter.metricRestService.SetMetrics)),
	))
	return endpoint
}

// @Summary Get metrics for device
// @Description Returns all metrics for the specified device ID
// @Tags Metric
// @Produce  json
// @Param   page	path	string	false	"string valid"	minlength(1)	maxlength(20)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /metrics/{id} [get]
// @Security JWT
func (metricRouter *MetricRouter) getByDeviceID(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/metrics/{id}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(metricRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(metricRouter.metricRestService.GetMetricsByDeviceID)),
	))
	return endpoint
}
