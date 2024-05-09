package rest

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type AlgorithmRestService interface {
	GetPage(w http.ResponseWriter, r *http.Request)
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
		negroni.Wrap(http.HandlerFunc(restService.GetPage)),
	)).Methods("GET")
	return []string{algorithmsEndpoint}
}

func (restService *DefaultAlgorithmRestService) GetPage(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	algorithms, err := restService.algorithmService.GetPage(session.GetConsistencyLevel(), 1, 10)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("algorithms=%+v", algorithms)

	restService.jsonWriter.Write(w, http.StatusOK, algorithms)

}
