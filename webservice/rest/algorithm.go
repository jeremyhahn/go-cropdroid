package rest

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/service"
)

type AlgorithmRestService interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultAlgorithmRestService struct {
	algorithmService  service.AlgorithmService
	middlewareService service.Middleware
	jsonWriter        common.HttpWriter
	AlgorithmRestService
}

func NewAlgorithmRestService(algorithmService service.AlgorithmService, middlewareService service.Middleware,
	jsonWriter common.HttpWriter) AlgorithmRestService {

	return &DefaultAlgorithmRestService{
		algorithmService:  algorithmService,
		middlewareService: middlewareService,
		jsonWriter:        jsonWriter}
}

func (restService *DefaultAlgorithmRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	algorithmsEndpoint := fmt.Sprintf("%s/algorithms", baseFarmURI)
	router.Handle(algorithmsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetAll)),
	)).Methods("GET")
	return []string{algorithmsEndpoint}
}

func (restService *DefaultAlgorithmRestService) GetAll(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer ctx.Close()

	algorithms, err := restService.algorithmService.GetAll()
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	ctx.GetLogger().Debugf("[AlgorithmRestService.GetAll] algorithms=%+v", algorithms)

	restService.jsonWriter.Write(w, http.StatusOK, algorithms)

}
