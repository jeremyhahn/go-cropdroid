package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	logging "github.com/op/go-logging"
)

type MetricRestServicer interface {
	SetMetrics(w http.ResponseWriter, r *http.Request)
	GetMetricsByDeviceID(w http.ResponseWriter, r *http.Request)
	RestService
}

type MetricRestService struct {
	logger        *logging.Logger
	metricService service.MetricService
	middleware    middleware.JsonWebTokenMiddleware
	httpWriter    response.HttpWriter
	MetricRestServicer
}

func NewMetricRestService(
	logger *logging.Logger,
	metricService service.MetricService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) MetricRestServicer {

	return &MetricRestService{
		logger:        logger,
		metricService: metricService,
		middleware:    middleware,
		httpWriter:    httpWriter}
}

func (restService *MetricRestService) SetMetrics(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var metric model.Metric
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&metric); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err = restService.metricService.Update(session, metric); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

func (restService *MetricRestService) GetMetricsByDeviceID(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	deviceID := params["id"]
	id, err := strconv.ParseUint(deviceID, 0, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	metrics, err := restService.metricService.GetAll(session, id)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, metrics)

}
