package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/shoppingcart"

	"io"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
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
	customerDefaultPaymentMethodEndpoint := fmt.Sprintf("%s/default-payment-method/{customerID}/{processorID}", customerEndpoint)
	getCustomerEndpoint := fmt.Sprintf("%s/{id}", customerEndpoint)
	invoiceEndpoint := fmt.Sprintf("%s/invoice", shoppingCartEndpoint)
	invoicePaymentEndpoint := fmt.Sprintf("%s/payment-webhook", invoiceEndpoint)
	paymentIntentEndpoint := fmt.Sprintf("%s/payment-intent", shoppingCartEndpoint)
	taxRateEndpoint := fmt.Sprintf("%s/tax-rate", shoppingCartEndpoint)

	router.Handle(invoicePaymentEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.InvoicePaymentWebHook)),
	)).Methods("GET")

	router.Handle(customerDefaultPaymentMethodEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.SetDefaultPaymentMethod)),
	)).Methods("GET")

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

	return []string{shoppingCartEndpoint, customerEndpoint, customerDefaultPaymentMethodEndpoint,
		getCustomerEndpoint, invoiceEndpoint, paymentIntentEndpoint,
		taxRateEndpoint}
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

func (restService *DefaultShoppingCartRestService) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	customerID := params["customerID"]
	processorID := params["processorID"]

	session.GetLogger().Debugf("customerID=%d, processorID=%s", customerID, processorID)

	id, err := strconv.ParseUint(customerID, 0, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	response, err := restService.shoppingCartService.SetDefaultPaymentMethod(id, processorID)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("response=%+v", response)

	restService.jsonWriter.Success200(w, response)
}

func (restService *DefaultShoppingCartRestService) InvoicePaymentWebHook(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	b, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	var invoice *stripe.Invoice
	err = json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	if event.Type == "invoice.payment_succeeded" {
		var invoice *stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}

		restService.shoppingCartService.UpdateSubscription(&shoppingcart.Subscription{
			ID:              invoice.Subscription.ID,
			PaymentIntentID: invoice.PaymentIntent.ID,
			LatestInvoiceID: invoice.ID})

		// pi, _ := paymentintent.Get(
		// 	invoice.PaymentIntent.ID,
		// 	nil,
		// )

		// params := &stripe.SubscriptionParams{
		// 	DefaultPaymentMethod: stripe.String(pi.PaymentMethod.ID),
		// }
		// subscription.Update(invoice.Subscription.ID, params)
		// fmt.Println("Default payment method set for subscription: ", pi.PaymentMethod)
	}

	if event.Type == "invoice.paid" {
		// Used to provision services after the trial has ended.
		// The status of the invoice will show up as paid. Store the status in your
		// database to reference when a user accesses your service to avoid hitting rate
		// limits.
		_, err = restService.shoppingCartService.UpdateSubscription(&shoppingcart.Subscription{
			ID:              invoice.Subscription.ID,
			PaymentIntentID: invoice.PaymentIntent.ID,
			LatestInvoiceID: invoice.ID})
		if err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}
	}

	if event.Type == "invoice.payment_failed" {
		// If the payment fails or the customer does not have a valid payment method,
		// an invoice.payment_failed event is sent, the subscription becomes past_due.
		// Use this webhook to notify your user that their payment has
		// failed and to retrieve new card details.
		return
	}

	if event.Type == "customer.subscription.deleted" {
		// handle subscription canceled automatically based
		// upon your subscription settings. Or if the user cancels it. {
		_, err = restService.shoppingCartService.CancelSubscription(invoice.Subscription.ID)
		if err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}
		return
	}

	restService.jsonWriter.Success200(w, nil)
}
