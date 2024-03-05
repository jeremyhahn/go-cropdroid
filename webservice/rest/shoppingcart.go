package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/shoppingcart"
)

type ShoppingCartRestService interface {
	GetProducts(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultShoppingCartRestService struct {
	shoppingCartService shoppingcart.ShoppingCartService
	middlewareService   service.Middleware
	jsonWriter          common.HttpWriter
	ShoppingCartRestService
}

func NewShoppingCartRestService(
	shoppingCartService shoppingcart.ShoppingCartService,
	middlewareService service.Middleware,
	jsonWriter common.HttpWriter) ShoppingCartRestService {

	return &DefaultShoppingCartRestService{
		shoppingCartService: shoppingCartService,
		middlewareService:   middlewareService,
		jsonWriter:          jsonWriter}
}

func (restService *DefaultShoppingCartRestService) RegisterEndpoints(
	router *mux.Router, baseURI, baseFarmURI string) []string {

	shoppingCartEndpoint := fmt.Sprintf("%s/shoppingcart", baseURI)
	router.Handle(shoppingCartEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetProducts)),
	)).Methods("GET")
	return []string{shoppingCartEndpoint}
}

func (restService *DefaultShoppingCartRestService) GetProducts(w http.ResponseWriter, r *http.Request) {

	ctx, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer ctx.Close()

	products := restService.shoppingCartService.GetProducts()
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	// Set a default product image if one is not configured
	// Doing this in the rest layer instead of service layer
	// because the baseURI is created in the webserver class
	// and not available in the service layer.
	for i, product := range products {
		if product.ImageUrl == "" {
			proto := "http://"
			if strings.Contains(r.Proto, "HTTPS") {
				proto = "https://"
			}
			products[i].ImageUrl = fmt.Sprintf("%s%s/images/logo-512px.png",
				proto, r.Host)
		}
	}

	ctx.GetLogger().Debugf("products=%+v", products)

	restService.jsonWriter.Success200(w, products)
}
