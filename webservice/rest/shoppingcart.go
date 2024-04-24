package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
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
	customerEndpoint := fmt.Sprintf("%s/customer", shoppingCartEndpoint)
	getCustomerEndpoint := fmt.Sprintf("%s/{id}", customerEndpoint)
	invoiceEndpoint := fmt.Sprintf("%s/invoice", shoppingCartEndpoint)
	paymentIntentEndpoint := fmt.Sprintf("%s/paymentIntent", shoppingCartEndpoint)
	taxRateEndpoint := fmt.Sprintf("%s/taxrate", shoppingCartEndpoint)

	router.Handle(getCustomerEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetCustomer)),
	)).Methods("GET")

	router.Handle(customerEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.CreateCustomer)),
	)).Methods("POST")

	router.Handle(customerEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.UpdateCustomer)),
	)).Methods("PUT")

	router.Handle(shoppingCartEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetProducts)),
	)).Methods("GET")

	router.Handle(invoiceEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.CreateInvoice)),
	)).Methods("POST")

	router.Handle(paymentIntentEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.CreatePaymentIntent)),
	)).Methods("POST")

	router.Handle(taxRateEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetTaxRates)),
	)).Methods("GET")

	return []string{shoppingCartEndpoint, paymentIntentEndpoint, invoiceEndpoint}
}

func (restService *DefaultShoppingCartRestService) GetProducts(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

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

	session.GetLogger().Debugf("products=%+v", products)

	restService.jsonWriter.Success200(w, products)
}

func (restService *DefaultShoppingCartRestService) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	var createPaymentIntentRequest shoppingcart.CreatePaymentIntentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createPaymentIntentRequest); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	paymentIntent, err := restService.shoppingCartService.CreatePaymentIntent(createPaymentIntentRequest)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("paymentIntent=%+v", paymentIntent)

	restService.jsonWriter.Success200(w, paymentIntent)
}

func (restService *DefaultShoppingCartRestService) CreateInvoice(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	var createInvoiceRequest shoppingcart.CreateInvoiceRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createInvoiceRequest); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	paymentIntentResponse, err := restService.shoppingCartService.CreateInvoice(session.GetUser(), createInvoiceRequest)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("paymentIntentResponse=%+v", paymentIntentResponse)

	restService.jsonWriter.Success200(w, paymentIntentResponse)
}

func (restService *DefaultShoppingCartRestService) GetCustomer(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	customerID := params["id"]

	session.GetLogger().Debugf("customerID=%s", customerID)

	id, err := strconv.ParseUint(customerID, 0, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	response, err := restService.shoppingCartService.GetCustomer(id)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("response=%+v", response)

	restService.jsonWriter.Success200(w, response)
}

func (restService *DefaultShoppingCartRestService) CreateCustomer(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	var customerConfig config.Customer
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&customerConfig); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	response, err := restService.shoppingCartService.CreateCustomer(&customerConfig)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("response=%+v", response)

	restService.jsonWriter.Success200(w, response)
}

func (restService *DefaultShoppingCartRestService) UpdateCustomer(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	var customerConfig config.Customer
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&customerConfig); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	response, err := restService.shoppingCartService.UpdateCustomer(&customerConfig)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("response=%+v", response)

	restService.jsonWriter.Success200(w, response)
}

func (restService *DefaultShoppingCartRestService) GetTaxRates(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	response := restService.shoppingCartService.GetTaxRates()

	session.GetLogger().Debugf("response=%+v", response)

	restService.jsonWriter.Success200(w, response)
}
