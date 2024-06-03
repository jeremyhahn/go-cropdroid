package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/shoppingcart"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	logging "github.com/op/go-logging"

	"io"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
)

type ShoppingCartRestServicer interface {
	GetPublishableKey(w http.ResponseWriter, r *http.Request)
	GetOrCreateCustomerWithEphemeralKey(w http.ResponseWriter, r *http.Request)
	GetProducts(w http.ResponseWriter, r *http.Request)
	GetTaxRates(w http.ResponseWriter, r *http.Request)
	GetCustomer(w http.ResponseWriter, r *http.Request)
	GetPaymentMethods(w http.ResponseWriter, r *http.Request)
	AttachAndSetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request)
	AttachPaymentMethod(w http.ResponseWriter, r *http.Request)
	SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request)
	GetSetupIntent(w http.ResponseWriter, r *http.Request)
	CreateSetupIntent(w http.ResponseWriter, r *http.Request)
	CreatePaymentIntent(w http.ResponseWriter, r *http.Request)
	CreateCustomer(w http.ResponseWriter, r *http.Request)
	UpdateCustomer(w http.ResponseWriter, r *http.Request)
	CreateInvoice(w http.ResponseWriter, r *http.Request)
	Webhook(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultShoppingCartRestService struct {
	logger              *logging.Logger
	webhookSecret       string
	shoppingCartService shoppingcart.ShoppingCartService
	middleware          middleware.JsonWebTokenMiddleware
	httpWriter          response.HttpWriter
	ShoppingCartRestServicer
}

func NewShoppingCartRestService(
	logger *logging.Logger,
	webhookSecret string,
	shoppingCartService shoppingcart.ShoppingCartService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) ShoppingCartRestServicer {

	return &DefaultShoppingCartRestService{
		logger:              logger,
		webhookSecret:       webhookSecret,
		shoppingCartService: shoppingCartService,
		middleware:          middleware,
		httpWriter:          httpWriter}
}

func (restService *DefaultShoppingCartRestService) GetProducts(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	products := restService.shoppingCartService.GetProducts()

	// Set a default product image if one is not configured
	// Doing this in the rest layer instead of service layer
	// because the baseURI is created in the webserver class
	// and not available in the service layer.
	for i, product := range products {
		if product.ImageUrl == "" {
			proto := "http://"
			if r.TLS != nil {
				proto = "https://"
			}
			products[i].ImageUrl = fmt.Sprintf("%s%s/images/logo-512px.png",
				proto, r.Host)
		}
	}
	restService.httpWriter.Success200(w, r, products)
}

func (restService *DefaultShoppingCartRestService) GetPublishableKey(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	publishableKey := restService.shoppingCartService.GetPublishableKey()
	restService.httpWriter.Success200(w, r, publishableKey)
}

func (restService *DefaultShoppingCartRestService) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var createPaymentIntentRequest shoppingcart.CreatePaymentIntentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createPaymentIntentRequest); err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	paymentIntent, err := restService.shoppingCartService.CreatePaymentIntent(createPaymentIntentRequest)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, paymentIntent)
}

func (restService *DefaultShoppingCartRestService) CreateInvoice(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var createInvoiceRequest shoppingcart.CreateInvoiceRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createInvoiceRequest); err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	paymentIntentResponse, err := restService.shoppingCartService.CreateInvoice(session.GetUser(), createInvoiceRequest)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, paymentIntentResponse)
}

func (restService *DefaultShoppingCartRestService) GetCustomer(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	params := mux.Vars(r)
	customerID := params["id"]

	id, err := strconv.ParseUint(customerID, 0, 64)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}
	response, err := restService.shoppingCartService.GetCustomer(id)
	if err == datastore.ErrRecordNotFound {
		restService.httpWriter.Success404(w, r, nil, err)
		return
	}
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, response)
}

func (restService *DefaultShoppingCartRestService) GetOrCreateCustomerWithEphemeralKey(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()
	user := session.GetUser()

	response, err := restService.shoppingCartService.GetOrCreateCustomerWithEphemeralKey(user)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, response)
}

func (restService *DefaultShoppingCartRestService) CreateCustomer(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var customerConfig *config.CustomerStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(customerConfig); err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	response, err := restService.shoppingCartService.CreateCustomer(customerConfig)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	session.GetLogger().Debugf("response=%+v", response)

	restService.httpWriter.Success200(w, r, response)
}

func (restService *DefaultShoppingCartRestService) UpdateCustomer(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var customerConfig *config.CustomerStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&customerConfig); err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	response, err := restService.shoppingCartService.UpdateCustomer(customerConfig)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, response)
}

// This is GET resource, but doing a POST to keep the client secret out of HTTP (Request URI) logs
func (restService *DefaultShoppingCartRestService) GetSetupIntent(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var setupIntentRequest shoppingcart.SetupIntentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&setupIntentRequest); err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	response, err := restService.shoppingCartService.GetSetupIntent(&setupIntentRequest)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, response)
}

func (restService *DefaultShoppingCartRestService) CreateSetupIntent(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()
	user := session.GetUser()

	response, err := restService.shoppingCartService.CreateSetupIntent(user)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, response)
}

func (restService *DefaultShoppingCartRestService) GetTaxRates(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	response := restService.shoppingCartService.GetTaxRates()
	restService.httpWriter.Success200(w, r, response)
}

func (restService *DefaultShoppingCartRestService) AttachPaymentMethod(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var attachPaymentRequest shoppingcart.AttachPaymentMethodRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&attachPaymentRequest); err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	response, err := restService.shoppingCartService.AttachPaymentMethod(&attachPaymentRequest)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, response)
}

func (restService *DefaultShoppingCartRestService) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var setDefaultPaymentMethodRequest shoppingcart.SetDefaultPaymentMethodRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&setDefaultPaymentMethodRequest); err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	response, err := restService.shoppingCartService.SetDefaultPaymentMethod(&setDefaultPaymentMethodRequest)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, response)
}

func (restService *DefaultShoppingCartRestService) AttachAndSetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var setDefaultPaymentMethodRequest shoppingcart.SetDefaultPaymentMethodRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&setDefaultPaymentMethodRequest); err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	response, err := restService.shoppingCartService.AttachAndSetDefaultPaymentMethod(&setDefaultPaymentMethodRequest)
	if err != nil {
		logger.Errorf("session: %+v, error: %s", session, err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	restService.httpWriter.Success200(w, r, response)
}

func (restService *DefaultShoppingCartRestService) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	processorID := params["processorID"]
	paymentMethods := restService.shoppingCartService.GetPaymentMethods(processorID)
	restService.httpWriter.Success200(w, r, paymentMethods)
}

// This is a public endpoint unprotected by JWT. There is no Token present in the request and therefore
// no way to look up a user or create a Session.
func (restService *DefaultShoppingCartRestService) Webhook(w http.ResponseWriter, r *http.Request) {

	restService.logger.Debugf("[Stripe WebHook] url: %s, method: %s, remoteAddress: %s, requestUri: %s",
		r.URL.Path, r.Method, r.RemoteAddr, r.RequestURI)

	b, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), restService.webhookSecret)
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
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
			restService.logger.Error(err)
			restService.httpWriter.Error400(w, r, err)
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
		restService.logger.Info("webhook setup_intent.succeeded")
		restService.logger.Infof("webhook %+v\n", event)
		// fmt.Fprintf(os.Stdout, "webhook %+v\n", string(event.Data.Raw))
		// var setupIntent *stripe.SetupIntent
		// err := json.Unmarshal(event.Data.Raw, &setupIntent)
		// if err != nil {
		// 	BadRequestError(w, r, err, restService.httpWriter)
		// 	return
		// }
		// if setupIntent.Metadata == nil {
		// 	BadRequestError(w, r, errors.New("missing required customer_id metadata"), restService.httpWriter)
		// 	return
		// }
		// customerIdMetadata, ok := setupIntent.Metadata["app_customer_id"]
		// if !ok {
		// 	BadRequestError(w, r, errors.New("missing required app_customer_id metadata"), restService.httpWriter)
		// 	return
		// }
		// customerID, err := strconv.ParseUint(customerIdMetadata, 0, 64)
		// if err != nil {
		// 	BadRequestError(w, r, err, restService.httpWriter)
		// 	return
		// }
		// restService.shoppingCartService.SetDefaultPaymentMethod(&shoppingcart.SetDefaultPaymentMethodRequest{
		// 	CustomerID:      customerID,
		// 	ProcessorID:     setupIntent.Customer.ID,
		// 	PaymentMethodID: setupIntent.PaymentMethod.ID})

	default:
		restService.logger.Errorf("[Stripe WebHook] unhandled event: %s\n", event.Type)
		return
	}

	restService.httpWriter.Success200(w, r, event.Type)
}
