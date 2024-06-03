package router

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

type AlgorithmRouter struct {
	algorithmRestService rest.AlgorithmRestServicer
	WebServiceRouter
}

// Creates a new web service algorithm router
func NewAlgorithmRouter(algorithmRestService rest.AlgorithmRestServicer) WebServiceRouter {
	return &AlgorithmRouter{algorithmRestService: algorithmRestService}
}

// Registers all of the algorithm endpoints at the root of the webservice (/api/v1)
func (algorithmRouter *AlgorithmRouter) RegisterRoutes(router *mux.Router, baseURI string) []string {
	return []string{
		algorithmRouter.page(router, baseURI)}
}

// @Summary List algorithms
// @Description Returns a page of algorithm entries
// @Tags Algorithms
// @Accept json
// @Produce  json
// @Param   page	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /algorithms/{page} [get]
// @Security JWT
func (algorithmRouter *AlgorithmRouter) page(router *mux.Router, baseURI string) string {
	endpoint := fmt.Sprintf("%s/algorithms/{page}", baseURI)
	router.HandleFunc(endpoint, algorithmRouter.algorithmRestService.Page)
	// router.Handle(endpoint, negroni.New(
	// 	negroni.HandlerFunc(algorithmRouter.middleware.Validate),
	// 	negroni.Wrap(http.HandlerFunc(algorithmRouter.algorithmRestService.GetPage)),
	// ))
	return endpoint
}
