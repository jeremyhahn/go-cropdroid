package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type MetricRestService interface {
	SetMetrics(w http.ResponseWriter, r *http.Request)
	GetMetricsByDeviceId(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultMetricRestService struct {
	metricService service.MetricService
	metricMapper  mapper.MetricMapper
	middleware    service.Middleware
	jsonWriter    common.HttpWriter
	MetricRestService
}

func NewMetricRestService(metricService service.MetricService, metricMapper mapper.MetricMapper,
	middleware service.Middleware, jsonWriter common.HttpWriter) MetricRestService {

	return &DefaultMetricRestService{
		metricService: metricService,
		metricMapper:  metricMapper,
		middleware:    middleware,
		jsonWriter:    jsonWriter}
}

func (restService *DefaultMetricRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	putMetricsEndpoint := fmt.Sprintf("%s/metrics", baseFarmURI)
	getMetricsEndpoint := fmt.Sprintf("%s/{id}", putMetricsEndpoint)
	router.Handle(putMetricsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.SetMetrics)),
	)).Methods("PUT")
	router.Handle(getMetricsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetMetricsByDeviceId)),
	)).Methods("GET")
	return []string{putMetricsEndpoint, getMetricsEndpoint}
}

func (restService *DefaultMetricRestService) SetMetrics(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	var metric model.Metric
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&metric); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("metric=%+v", metric)

	if err = restService.metricService.Update(session, &metric); err != nil {
		session.GetLogger().Errorf("Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}

func (restService *DefaultMetricRestService) GetMetricsByDeviceId(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	deviceID := params["id"]

	session.GetLogger().Debugf("deviceID=%s", deviceID)

	id, err := strconv.ParseUint(deviceID, 0, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	metrics, err := restService.metricService.GetAll(session, id)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("metrics=%+v", metrics)

	restService.jsonWriter.Write(w, http.StatusOK, metrics)

}
