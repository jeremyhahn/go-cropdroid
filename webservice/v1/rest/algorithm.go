package rest

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type AlgorithmRestServicer interface {
	SetService(algorithmService service.AlgorithmServicer)
	Page(w http.ResponseWriter, r *http.Request)
}

type AlgorithmRestService struct {
	algorithmService service.AlgorithmServicer
	middleware       middleware.JsonWebTokenMiddleware
	httpWriter       response.HttpWriter
	AlgorithmRestServicer
}

func NewAlgorithmRestService(
	algorithmService service.AlgorithmServicer,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) AlgorithmRestServicer {

	return &AlgorithmRestService{
		algorithmService: algorithmService,
		middleware:       middleware,
		httpWriter:       httpWriter}
}

// Dependency injection to set mocked algorithm service
func (restService *AlgorithmRestService) SetService(algorithmService service.AlgorithmServicer) {
	restService.algorithmService = algorithmService
}

// Returns a page of algorithms
func (restService *AlgorithmRestService) Page(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}

	logger := session.GetLogger()

	params := mux.Vars(r)
	page := params["page"]

	p, err := strconv.Atoi(page)
	if err != nil {
		logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	pageQuery := query.NewPageQuery()
	pageQuery.Page = p
	pageQuery.SortOrder = query.SORT_DESCENDING

	defer session.Close()
	algorithms, err := restService.algorithmService.Page(query.NewPageQuery(), session.GetConsistencyLevel())
	if err != nil {
		logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, algorithms)
}
