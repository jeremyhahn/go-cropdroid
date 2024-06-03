package router

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/shoppingcart"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
	"github.com/op/go-logging"
)

type ShoppingCartRouter struct {
	middleware              middleware.JsonWebTokenMiddleware
	shoppingCartRestService rest.ShoppingCartRestServicer
	WebServiceRouter
}

// Creates a new web service shopping cart router
func NewShoppingCartRouter(
	logger *logging.Logger,
	webhookSecret string,
	shoppingCartService shoppingcart.ShoppingCartService,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &ShoppingCartRouter{
		middleware: middleware,
		shoppingCartRestService: rest.NewShoppingCartRestService(
			logger,
			webhookSecret,
			shoppingCartService,
			middleware,
			httpWriter)}
}

// Registers all of the shoppingCart endpoints at the root of the webservice (/api/v1)
func (cartRouter *ShoppingCartRouter) RegisterRoutes(router *mux.Router, baseFarmURI string) []string {
	baseCartURI := fmt.Sprintf("%s/shoppingcart", baseFarmURI)
	return []string{
		cartRouter.publishableKey(router, baseCartURI),
		cartRouter.getOrCreateCustomerWithEphemeralKey(router, baseCartURI),
		cartRouter.products(router, baseCartURI),
		cartRouter.taxRates(router, baseCartURI),
		cartRouter.customer(router, baseCartURI),
		cartRouter.defaultPaymentMethods(router, baseCartURI),
		cartRouter.attachAndSetDefaultPaymentMethod(router, baseCartURI),
		cartRouter.attachPaymentMethod(router, baseCartURI),
		cartRouter.setDefaultPaymentMethod(router, baseCartURI),
		cartRouter.getSetupIntent(router, baseCartURI),
		cartRouter.createSetupIntent(router, baseCartURI),
		cartRouter.createPaymentIntent(router, baseCartURI),
		cartRouter.createCustomer(router, baseCartURI),
		cartRouter.updateCustomer(router, baseCartURI),
		cartRouter.createInvoice(router, baseCartURI),
		cartRouter.webhook(router, baseCartURI)}
}

// @Summary Get stripe publishable key
// @Description Returns the Stripe API publishable key that corresponds with the configured Stripe account
// @Tags Stripe
// @Produce  json
// @Param   page	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/publishable-key [get]
// @Security JWT
func (cartRouter *ShoppingCartRouter) publishableKey(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/publishable-key", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.GetPublishableKey)),
	))
	return endpoint
}

// @Summary Get stripe customer ephemeral key
// @Description Returns the stripe customer ephemeral key for the customer/user session
// @Tags Stripe
// @Produce  json
// @Param   customerID	path	integer	true	"string valid" "This users ID from the session is used instead of the passed customerID"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/ephemeral-key/{customerID} [get]
// @Security JWT
func (cartRouter *ShoppingCartRouter) getOrCreateCustomerWithEphemeralKey(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/ephemeral-key/{customerID}", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.GetOrCreateCustomerWithEphemeralKey)),
	))
	return endpoint
}

// @Summary Get product list
// @Description Returns all products configured in Stripe
// @Tags Stripe
// @Produce  json
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/products [get]
// @Security JWT
func (cartRouter *ShoppingCartRouter) products(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/products", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.GetProducts)),
	))
	return endpoint
}

// @Summary Get tax rate
// @Description Returns the configured tax rates. When a tax rate is not configured, Stripe Automatic Tax is used instead.
// @Tags Stripe
// @Produce  json
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/tax-rate [get]
// @Security JWT
func (cartRouter *ShoppingCartRouter) taxRates(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/tax-rate", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.GetTaxRates)),
	))
	return endpoint
}

// @Summary Get customer default payment method
// @Description Returns the customers default payment method
// @Tags Stripe
// @Produce  json
// @Param   processorID	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/payment-methods/{processorID} [get]
// @Security JWT
func (cartRouter *ShoppingCartRouter) defaultPaymentMethods(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/payment-methods/{processorID}", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.GetPaymentMethods)),
	))
	return endpoint
}

// @Summary Attach and set default payment method
// @Description Attaches the specified payment method and sets it as the customers default
// @Tags Stripe
// @Produce  json
// @Param SetDefaultPaymentMethodRequest body shoppingcart.SetDefaultPaymentMethodRequest true "shoppingcart.SetDefaultPaymentMethodRequest struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/attach-and-set-default-payment-method [post]
// @Security JWT
func (cartRouter *ShoppingCartRouter) attachAndSetDefaultPaymentMethod(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/attach-and-set-default-payment-method", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.AttachAndSetDefaultPaymentMethod)),
	)).Methods("POST")
	return endpoint
}

// @Summary Attach new payment method
// @Description Attaches a payment method to the customer
// @Tags Stripe
// @Produce  json
// @Param AttachPaymentMethodRequest body shoppingcart.AttachPaymentMethodRequest true "shoppingcart.AttachPaymentMethodRequest struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/attach-payment-method [post]
// @Security JWT
func (cartRouter *ShoppingCartRouter) attachPaymentMethod(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/attach-payment-method", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.AttachPaymentMethod)),
	)).Methods("POST")
	return endpoint
}

// @Summary Set default payment method
// @Description Sets an attached payment method as the default
// @Tags Stripe
// @Produce  json
// @Param SetDefaultPaymentMethodRequest body shoppingcart.SetDefaultPaymentMethodRequest true "shoppingcart.SetDefaultPaymentMethodRequest struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/default-payment-method [post]
// @Security JWT
func (cartRouter *ShoppingCartRouter) setDefaultPaymentMethod(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/default-payment-method", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.SetDefaultPaymentMethod)),
	)).Methods("POST")
	return endpoint
}

// @Summary Retrieve a setup-intent
// @Description Retrieves an existing setup-intent using a POST request to protect the client secret
// @Tags Stripe
// @Produce  json
// @Param SetupIntentRequest body shoppingcart.SetupIntentRequest true "shoppingcart.SetupIntentRequest struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/setup-intent/secret [post]
// @Security JWT
func (cartRouter *ShoppingCartRouter) getSetupIntent(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/setup-intent/secret", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.GetSetupIntent)),
	)).Methods("POST")
	return endpoint
}

// @Summary Create setup-intent
// @Description Creates a new setup-intent
// @Tags Stripe
// @Produce  json
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/setup-intent [post]
// @Security JWT
func (cartRouter *ShoppingCartRouter) createSetupIntent(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/setup-intent", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.CreateSetupIntent)),
	)).Methods("POST")
	return endpoint
}

// @Summary Get customer
// @Description Returns a Customer with the unique Stripe assigned cust_* ID
// @Tags Stripe
// @Produce  json
// @Param   id	path	integer	true	"string valid" "The local user / customer ID"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/customers/{id} [get]
// @Security JWT
func (cartRouter *ShoppingCartRouter) customer(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/customers/{id}", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.GetCustomer)),
	))
	return endpoint
}

// @Summary Create new payment-intent
// @Description Creates a new payment-intent for the user in the JWT request
// @Tags Stripe
// @Produce  json
// @Param CreatePaymentIntentRequest body shoppingcart.CreatePaymentIntentRequest true "shoppingcart.CreatePaymentIntentRequest struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/payment-intent [post]
// @Security JWT
func (cartRouter *ShoppingCartRouter) createPaymentIntent(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/payment-intent", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.CreatePaymentIntent)),
	)).Methods("POST")
	return endpoint
}

// @Summary Create customer
// @Description Creates a new Stripe and local customer
// @Tags Stripe
// @Produce  json
// @Param Customer body config.Customer true "config.Customer struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/customers [post]
// @Security JWT
func (cartRouter *ShoppingCartRouter) createCustomer(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/customers", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.CreateCustomer)),
	)).Methods("POST")
	return endpoint
}

// @Summary Update customer
// @Description Updates an existing Stripe and local customer
// @Tags Stripe
// @Produce  json
// @Param Customer body config.Customer true "config.Customer struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/customers [put]
// @Security JWT
func (cartRouter *ShoppingCartRouter) updateCustomer(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/customers", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.UpdateCustomer)),
	)).Methods("PUT")
	return endpoint
}

// @Summary Create invoice
// @Description Creates a new customer invoice
// @Tags Stripe
// @Produce  json
// @Param CreateInvoiceRequest body shoppingcart.CreateInvoiceRequest true "shoppingcart.CreateInvoiceRequest struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/invoice [post]
// @Security JWT
func (cartRouter *ShoppingCartRouter) createInvoice(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/invoice", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(cartRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.CreateInvoice)),
	)).Methods("POST")
	return endpoint
}

// @Summary WebHook Handler
// @Description Handles public webhook callbacks from Stripe
// @Tags Stripe
// @Produce  json
// @Param Event body stripe.Event true "stripe.Event struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /shoppingcart/webhook [post]
func (cartRouter *ShoppingCartRouter) webhook(router *mux.Router, baseCartURI string) string {
	endpoint := fmt.Sprintf("%s/webhook", baseCartURI)
	router.Handle(endpoint, negroni.New(
		negroni.Wrap(http.HandlerFunc(cartRouter.shoppingCartRestService.Webhook)),
	)).Methods("POST")
	return endpoint
}
