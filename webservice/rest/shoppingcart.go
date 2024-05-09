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
	GetSetupIntent(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultShoppingCartRestService struct {
	shoppingCartService shoppingcart.ShoppingCartService
	webhookSecret       string
	middlewareService   service.Middleware
	jsonWriter          common.HttpWriter
	ShoppingCartRestService
}

func NewShoppingCartRestService(
	shoppingCartService shoppingcart.ShoppingCartService,
	webhookSecret string,
	middlewareService service.Middleware,
	jsonWriter common.HttpWriter) ShoppingCartRestService {

	return &DefaultShoppingCartRestService{
		shoppingCartService: shoppingCartService,
		webhookSecret:       webhookSecret,
		middlewareService:   middlewareService,
		jsonWriter:          jsonWriter}
}

func (restService *DefaultShoppingCartRestService) RegisterEndpoints(
	router *mux.Router, baseURI, baseFarmURI string) []string {

	shoppingCartEndpoint := fmt.Sprintf("%s/shoppingcart", baseURI)
	publishableKeuEndpoint := fmt.Sprintf("%s/publishable-key", shoppingCartEndpoint)
	customerEndpoint := fmt.Sprintf("%s/customer", shoppingCartEndpoint)
	customerEphemeralKeyEndpoint := fmt.Sprintf("%s/ephemeral-key/{customerID}", customerEndpoint)
	customerSetupIntentWithSecretEndpoint := fmt.Sprintf("%s/setup-intent/secret", customerEndpoint)
	customerSetupIntentEndpoint := fmt.Sprintf("%s/setup-intent", customerEndpoint)
	attachAndSetDefaultPaymentMethodEndpoint := fmt.Sprintf("%s/attach-and-set-default-payment-method", customerEndpoint)
	attachPaymentMethodEndpoint := fmt.Sprintf("%s/attach-payment-method", customerEndpoint)
	getDefaultPaymentMethodEndpoint := fmt.Sprintf("%s/payment-methods/{customerID}", customerEndpoint)
	setCustomerDefaultPaymentMethodEndpoint := fmt.Sprintf("%s/default-payment-method", customerEndpoint)
	getCustomerEndpoint := fmt.Sprintf("%s/{id}", customerEndpoint)
	invoiceEndpoint := fmt.Sprintf("%s/invoice", shoppingCartEndpoint)
	// invoicePaymentEndpoint := fmt.Sprintf("%s/payment-webhook", invoiceEndpoint)
	paymentIntentEndpoint := fmt.Sprintf("%s/payment-intent", shoppingCartEndpoint)
	taxRateEndpoint := fmt.Sprintf("%s/tax-rate", shoppingCartEndpoint)
	webhooksEndpoint := fmt.Sprintf("%s/webhook", shoppingCartEndpoint)

	router.Handle(publishableKeuEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetPublishableKey)),
	)).Methods("GET")

	// router.Handle(invoicePaymentEndpoint, negroni.New(
	// 	negroni.HandlerFunc(restService.middlewareService.Validate),
	// 	negroni.Wrap(http.HandlerFunc(restService.InvoicePaymentWebHook)),
	// )).Methods("GET")

	router.Handle(attachAndSetDefaultPaymentMethodEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.AttachAndSetDefaultPaymentMethod)),
	)).Methods("POST")

	router.Handle(attachPaymentMethodEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.AttachPaymentMethod)),
	)).Methods("POST")

	router.Handle(getDefaultPaymentMethodEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetPaymentMethods)),
	)).Methods("GET")

	router.Handle(setCustomerDefaultPaymentMethodEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.SetDefaultPaymentMethod)),
	)).Methods("POST")

	router.Handle(getCustomerEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetCustomer)),
	)).Methods("GET")

	router.Handle(customerEphemeralKeyEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetOrCreateCustomerWithEphemeralKey)),
	)).Methods("GET")

	router.Handle(customerSetupIntentWithSecretEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetSetupIntent)),
	)).Methods("POST")

	router.Handle(customerSetupIntentEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.CreateSetupIntent)),
	)).Methods("POST")

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

	router.Handle(webhooksEndpoint, negroni.New(
		//negroni.HandlerFunc(restService.middlewareService.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Webhook)),
	)).Methods("POST")

	return []string{shoppingCartEndpoint, publishableKeuEndpoint, customerEndpoint,
		customerEphemeralKeyEndpoint, customerSetupIntentWithSecretEndpoint, customerSetupIntentEndpoint,
		attachAndSetDefaultPaymentMethodEndpoint, attachPaymentMethodEndpoint,
		getDefaultPaymentMethodEndpoint, setCustomerDefaultPaymentMethodEndpoint, getCustomerEndpoint,
		invoiceEndpoint, paymentIntentEndpoint, taxRateEndpoint, webhooksEndpoint}
}

// @Summary GetProducts
// @Description Returns a list of products from the merchange provider / credit card processor
// @Produce  json
// @Success 200
// @Router /api/v1/shoppingcart [get]
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

func (restService *DefaultShoppingCartRestService) GetPublishableKey(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	publishableKey := restService.shoppingCartService.GetPublishableKey()

	session.GetLogger().Debugf("publishableKey=%+v", publishableKey)

	restService.jsonWriter.Success200(w, publishableKey)
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

func (restService *DefaultShoppingCartRestService) GetOrCreateCustomerWithEphemeralKey(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	user := session.GetUser()
	session.GetLogger().Debugf("user=%+v", user)

	response, err := restService.shoppingCartService.GetOrCreateCustomerWithEphemeralKey(user)
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

	response, err := restService.shoppingCartService.CreateCustomer(customerConfig)
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

	response, err := restService.shoppingCartService.UpdateCustomer(customerConfig)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("response=%+v", response)

	restService.jsonWriter.Success200(w, response)
}

// This is GET resource, but doing a POST to keep the client secret out of HTTP (Request URI) logs
func (restService *DefaultShoppingCartRestService) GetSetupIntent(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	var setupIntentRequest shoppingcart.SetupIntentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&setupIntentRequest); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("setupIntentRequest=%+v", setupIntentRequest)

	response, err := restService.shoppingCartService.GetSetupIntent(&setupIntentRequest)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("response=%+v", response)

	restService.jsonWriter.Success200(w, response)
}

func (restService *DefaultShoppingCartRestService) CreateSetupIntent(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	user := session.GetUser()

	session.GetLogger().Debugf("user=%+v", user)

	response, err := restService.shoppingCartService.CreateSetupIntent(user)
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

func (restService *DefaultShoppingCartRestService) AttachPaymentMethod(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	var attachPaymentRequest shoppingcart.AttachPaymentMethodRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&attachPaymentRequest); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("attachPaymentRequest=%+v", attachPaymentRequest)

	response, err := restService.shoppingCartService.AttachPaymentMethod(&attachPaymentRequest)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

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

	var setDefaultPaymentMethodRequest shoppingcart.SetDefaultPaymentMethodRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&setDefaultPaymentMethodRequest); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("setDefaultPaymentMethodRequest=%+v", setDefaultPaymentMethodRequest)

	response, err := restService.shoppingCartService.SetDefaultPaymentMethod(&setDefaultPaymentMethodRequest)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("response=%+v", response)

	restService.jsonWriter.Success200(w, response)
}

func (restService *DefaultShoppingCartRestService) AttachAndSetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	var setDefaultPaymentMethodRequest shoppingcart.SetDefaultPaymentMethodRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&setDefaultPaymentMethodRequest); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("AttachAndSetDefaultPaymentMethod=%+v", setDefaultPaymentMethodRequest)

	response, err := restService.shoppingCartService.AttachAndSetDefaultPaymentMethod(&setDefaultPaymentMethodRequest)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("response=%+v", response)

	restService.jsonWriter.Success200(w, response)
}

func (restService *DefaultShoppingCartRestService) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middlewareService.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	processorID := params["processorID"]

	session.GetLogger().Debugf("processorID=%s", processorID)

	paymentMethods := restService.shoppingCartService.GetPaymentMethods(processorID)

	session.GetLogger().Debugf("response=%+v", paymentMethods)

	restService.jsonWriter.Success200(w, paymentMethods)
}

// This is a public endpoint unprotected by JWT. There is no Token present in the request and therefore
// no way to look up a user or create a Session.
func (restService *DefaultShoppingCartRestService) Webhook(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(os.Stdout, "webhook called")

	b, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "webhook: %+v", err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), restService.webhookSecret)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "customer.subscription.created":
		// Then define and call a function to handle the event customer.subscription.created
	case "customer.subscription.deleted":
		// Then define and call a function to handle the event customer.subscription.deleted
	case "customer.subscription.paused":
		// Then define and call a function to handle the event customer.subscription.paused
	case "customer.subscription.resumed":
		// Then define and call a function to handle the event customer.subscription.resumed
	case "customer.subscription.trial_will_end":
		// Then define and call a function to handle the event customer.subscription.trial_will_end
	case "invoice.overdue":
		// Then define and call a function to handle the event invoice.overdue
	case "invoice.paid":
		// Then define and call a function to handle the event invoice.paid
	case "invoice.payment_action_required":
		// Then define and call a function to handle the event invoice.payment_action_required
	case "invoice.payment_failed":
		// Then define and call a function to handle the event invoice.payment_failed
	case "invoice.payment_succeeded":
		// Then define and call a function to handle the event invoice.payment_succeeded
		var invoice *stripe.Invoice
		err = json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}
	case "invoice.will_be_due":
		// Then define and call a function to handle the event invoice.will_be_due
	case "payment_intent.payment_failed":
		// Then define and call a function to handle the event payment_intent.payment_failed
	case "payment_intent.succeeded":
		// Then define and call a function to handle the event payment_intent.succeeded
	case "setup_intent.created":
		// Then define and call a function to handle the event setup_intent.created
	case "setup_intent.succeeded":
		// Then define and call a function to handle the event setup_intent.succeeded
		fmt.Fprintln(os.Stdout, "webhook setup_intent.succeeded")
		fmt.Fprintf(os.Stdout, "webhook %+v\n", event)
		// fmt.Fprintf(os.Stdout, "webhook %+v\n", string(event.Data.Raw))
		// var setupIntent *stripe.SetupIntent
		// err := json.Unmarshal(event.Data.Raw, &setupIntent)
		// if err != nil {
		// 	BadRequestError(w, r, err, restService.jsonWriter)
		// 	return
		// }
		// if setupIntent.Metadata == nil {
		// 	BadRequestError(w, r, errors.New("missing required customer_id metadata"), restService.jsonWriter)
		// 	return
		// }
		// customerIdMetadata, ok := setupIntent.Metadata["app_customer_id"]
		// if !ok {
		// 	BadRequestError(w, r, errors.New("missing required app_customer_id metadata"), restService.jsonWriter)
		// 	return
		// }
		// customerID, err := strconv.ParseUint(customerIdMetadata, 0, 64)
		// if err != nil {
		// 	BadRequestError(w, r, err, restService.jsonWriter)
		// 	return
		// }
		// restService.shoppingCartService.SetDefaultPaymentMethod(&shoppingcart.SetDefaultPaymentMethodRequest{
		// 	CustomerID:      customerID,
		// 	ProcessorID:     setupIntent.Customer.ID,
		// 	PaymentMethodID: setupIntent.PaymentMethod.ID})

	default:
		fmt.Fprintf(os.Stderr, "webhook unhandled event type: %s\n", event.Type)
	}

	restService.jsonWriter.Success200(w, event.Type)
}
