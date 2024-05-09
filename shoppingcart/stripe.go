package shoppingcart

// https://stripe.com/legal/restricted-businesses
// https://docs.stripe.com/testing?testing-method=payment-methods#visa
// https://docs.stripe.com/billing/taxes/tax-rates
// https://docs.stripe.com/tax/tax-codes
// https://github.com/stripe-samples/subscription-use-cases/tree/main/fixed-price-subscriptions

// Rate limiting
// https://docs.stripe.com/rate-limits#load-testing

// Test credit card numbers
// https://docs.stripe.com/testing

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/ephemeralkey"
	"github.com/stripe/stripe-go/v78/invoice"
	"github.com/stripe/stripe-go/v78/invoiceitem"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/paymentmethod"
	"github.com/stripe/stripe-go/v78/price"
	"github.com/stripe/stripe-go/v78/product"
	"github.com/stripe/stripe-go/v78/setupintent"
	"github.com/stripe/stripe-go/v78/subscription"
	"github.com/stripe/stripe-go/v78/taxrate"
	"gorm.io/gorm"
)

var ErrInvoiceNotPaid = errors.New("invoice not paid")
var ErrInvoiceOutstanding = errors.New("outstanding invoice")
var ErrMissingPaymentMethod = errors.New("missing payment method")
var STRIPE_DEFAULT_LIMIT = int64(10)

type ShoppingCartService interface {
	GetPublishableKey() string
	GetProducts() []Product
	GetTaxRates() []*TaxRate

	CreatePaymentIntent(createPaymentIntentRequest CreatePaymentIntentRequest) (PaymentIntentResponse, error)
	GetPaymentIntent(id string) (PaymentIntentResponse, error)

	GetSetupIntent(setupIntentRequest *SetupIntentRequest) (*stripe.SetupIntent, error)
	CreateSetupIntent(user common.UserAccount) (*SetupIntentResponse, error)
	ListSetupIntents() []*stripe.SetupIntent

	GetCustomer(id uint64) (*config.Customer, error)
	GetOrCreateCustomerWithEphemeralKey(user common.UserAccount) (*CustomerEphemeralKeyResponse, error)
	CreateCustomer(customerConfig *config.Customer) (*config.Customer, error)
	UpdateCustomer(customerConfig *config.Customer) (*config.Customer, error)

	GetInvoice(invoiceID string) (*Invoice, error)
	CreateInvoice(userAccount common.UserAccount, createInvoiceRequest CreateInvoiceRequest) (PaymentIntentResponse, error)
	PayInvoice(invoiceID string) (*Invoice, error)
	InvoicePreview(customerID, subscriptionID, priceID string) (*Invoice, error)

	CreateSubscription(subscriptionConfig *Subscription) (*Subscription, error)
	UpdateSubscription(subscriptionConfig *Subscription) (*Subscription, error)
	CancelSubscription(subscriptionID string) (*Subscription, error)
	ListSubscriptions(processorID string) []*Subscription

	GetPaymentMethod(paymentMethodID string) (*stripe.PaymentMethod, error)
	GetPaymentMethods(customerID string) []*stripe.PaymentMethod
	CreateCardPaymentMethod(card *CreditCard) (string, error)
	AttachPaymentMethod(attachPaymentRequest *AttachPaymentMethodRequest) (*stripe.PaymentMethod, error)
	SetDefaultPaymentMethod(setDefaultPaymentMethodRequest *SetDefaultPaymentMethodRequest) (*config.Customer, error)
	AttachAndSetDefaultPaymentMethod(setDefaultPaymentMethodRequest *SetDefaultPaymentMethodRequest) (*config.Customer, error)
}

type StripeClient struct {
	app         *app.App
	params      *config.Stripe
	customerDAO dao.CustomerDAO
	ShoppingCartService
}

func NewStripeService(_app *app.App, customerDAO dao.CustomerDAO) ShoppingCartService {
	stripe.Key = _app.Stripe.Key.Secret
	// stripe.SetAppInfo(&stripe.AppInfo{
	// 	Name:    _app.Name,
	// 	Version: app.Release,
	// 	URL:     "https://www.cropdroid.com",
	// })
	return &StripeClient{
		app:         _app,
		params:      _app.Stripe,
		customerDAO: customerDAO}
}

func (client *StripeClient) GetProducts() []Product {
	products := make([]Product, 0)
	params := &stripe.ProductListParams{}
	params.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
	result := product.List(params)
	for result.Iter.Next() {
		product := result.Iter.Current().(*stripe.Product)
		if !product.Active {
			continue
		}
		var hasFilterTags = false
		for k, v := range product.Metadata {
			// if k == "category" && v == "subscription" {
			// 	hasFilterTags = true
			// 	break
			// }
			if k == "visible" && v == "false" {
				hasFilterTags = true
				break
			}
		}
		if hasFilterTags {
			continue
		}
		priceParams := &stripe.PriceListParams{}
		priceParams.Product = stripe.String(product.ID)
		priceParams.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
		result := price.List(priceParams)
		for result.Iter.Next() {
			price := result.Iter.Current()
			if !price.(*stripe.Price).Active {
				continue
			}
			imageUrl := ""
			if len(product.Images) > 0 {
				imageUrl = product.Images[0]
			}
			products = append(products, Product{
				Id:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				ImageUrl:    imageUrl,
				Price:       price.(*stripe.Price).UnitAmount,
				Metadata:    product.Metadata})
		}
	}
	return products
}

func (client *StripeClient) GetPublishableKey() string {
	return client.app.Stripe.Key.Publishable
}

func (client *StripeClient) CreateSetupIntent(user common.UserAccount) (*SetupIntentResponse, error) {
	customerEphemeralKeyResponse, err := client.GetOrCreateCustomerWithEphemeralKey(user)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	params := &stripe.SetupIntentParams{
		Customer:           stripe.String(customerEphemeralKeyResponse.Customer.ProcessorID),
		PaymentMethodTypes: []*string{stripe.String("card")},
		Metadata: map[string]string{
			"app_customer_id": strconv.Itoa(int(user.GetID())),
		},
	}
	setupIntent, err := setupintent.New(params)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	return &SetupIntentResponse{
		Customer:       customerEphemeralKeyResponse.Customer,
		EphemeralKey:   customerEphemeralKeyResponse.EphemeralKey,
		ClientSecret:   setupIntent.ClientSecret,
		PublishableKey: client.app.Stripe.Key.Publishable}, nil
}

func (client *StripeClient) ListSetupIntents() []*stripe.SetupIntent {
	setupIntents := make([]*stripe.SetupIntent, 0)
	params := &stripe.SetupIntentListParams{}
	params.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
	result := setupintent.List(params)
	for result.Iter.Next() {
		setupIntent := result.Iter.Current().(*stripe.SetupIntent)
		setupIntents = append(setupIntents, setupIntent)

	}
	return setupIntents
}

// Retrieves a SetupIntent using the ClientSecret that was used to confirm it
func (client *StripeClient) GetSetupIntent(setupIntentRequest *SetupIntentRequest) (*stripe.SetupIntent, error) {

	pieces := strings.Split(setupIntentRequest.ClientSecret, "_secret_")
	setupIntentID := pieces[0]
	//secret := pieces[1]

	params := &stripe.SetupIntentParams{}
	setupIntent, err := setupintent.Get(setupIntentID, params)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	return setupIntent, nil
}

// Retrieve a user with their ephemeral key. A new user is created if they don't already exist
func (client *StripeClient) GetOrCreateCustomerWithEphemeralKey(user common.UserAccount) (*CustomerEphemeralKeyResponse, error) {
	customerConfig, err := client.customerDAO.Get(user.GetID(), common.CONSISTENCY_LOCAL)
	if err == gorm.ErrRecordNotFound {
		customerConfig, err = client.CreateCustomer(&config.Customer{
			ID:    user.GetID(),
			Name:  user.GetEmail(),
			Email: user.GetEmail(),
			Description: fmt.Sprintf("This is a placeholder for %s, who is the process of becoming a new customer %s. DO NOT DELETE!",
				user.GetEmail(), time.Now().Format(common.TIME_FORMAT))})
		if err != nil {
			client.app.Logger.Error(err)
			return nil, err
		}
	}
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	ephemeralKey, err := client.createEphemeralKey(customerConfig.ProcessorID)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	return &CustomerEphemeralKeyResponse{
		Customer:     *customerConfig,
		EphemeralKey: ephemeralKey.Secret}, nil
}

func (client *StripeClient) GetCustomer(id uint64) (*config.Customer, error) {
	customerConfig, err := client.customerDAO.Get(id, common.CONSISTENCY_LOCAL)
	// if err == gorm.ErrRecordNotFound {
	// 	return emptyCustomer, nil
	// }
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	params := &stripe.CustomerParams{}
	stripeCustomer, err := customer.Get(customerConfig.ProcessorID, params)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	if stripeCustomer.InvoiceSettings != nil && stripeCustomer.InvoiceSettings.DefaultPaymentMethod != nil {
		paymentMethod, err := client.GetPaymentMethod(stripeCustomer.InvoiceSettings.DefaultPaymentMethod.ID)
		if err != nil {
			client.app.Logger.Error(err)
			return nil, err
		}
		customerConfig.PaymentMethodID = paymentMethod.ID
		customerConfig.PaymentMethodLast4 = paymentMethod.Card.Last4
	} else {
		customerConfig.PaymentMethodID = ""
		customerConfig.PaymentMethodLast4 = ""
	}
	hydratedCustomerConfig := client.mapStripeCustomerToCustomerConfig(stripeCustomer)
	hydratedCustomerConfig.ID = customerConfig.ID
	// The stripe customer doesnt have the last 4 set, so it cant be mapped
	hydratedCustomerConfig.PaymentMethodLast4 = customerConfig.PaymentMethodLast4
	return hydratedCustomerConfig, nil
}

func (client *StripeClient) CreateCustomer(customerConfig *config.Customer) (*config.Customer, error) {

	customerParams := client.mapCustomerConfigToStripeCustomerParams(customerConfig)
	stripeCustomer, err := customer.New(customerParams)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	customerConfig.ProcessorID = stripeCustomer.ID
	if err := client.customerDAO.Save(customerConfig); err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	return customerConfig, err
}

func (client *StripeClient) UpdateCustomer(customerConfig *config.Customer) (*config.Customer, error) {
	customerParams := client.mapCustomerConfigToStripeCustomerParams(customerConfig)
	_, err := customer.Update(customerConfig.ProcessorID, customerParams)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	if err := client.customerDAO.Update(customerConfig); err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	return customerConfig, err
}

func (client *StripeClient) ListInvoices(processorID string, status *string) []*stripe.Invoice {
	invoices := make([]*stripe.Invoice, 0)
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(processorID),
		Status:   status,
	}
	params.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
	stripeInvoices := invoice.List(params)
	for stripeInvoices.Iter.Next() {
		stripeInvoice := stripeInvoices.Iter.Current().(*stripe.Invoice)
		invoices = append(invoices, stripeInvoice)

	}
	return invoices
}

func (client *StripeClient) GetInvoice(invoiceID string) (*Invoice, error) {

	params := &stripe.InvoiceParams{}
	stripeInvoice, err := invoice.Get(invoiceID, params)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	return client.mapStripeInvoiceToInvoice(stripeInvoice), nil
}

func (client *StripeClient) CreateInvoice(userAccount common.UserAccount,
	createInvoiceRequest CreateInvoiceRequest) (PaymentIntentResponse, error) {

	var EmptyPaymentIntentResponse = PaymentIntentResponse{}
	var fixedTaxRates []*string
	//var dynamicTaxRates []*string

	if client.app.Stripe.Tax != nil && client.app.Stripe.Tax.Fixed != nil {
		fixedTaxRates = client.app.Stripe.Tax.Fixed
	}
	// if client.app.Stripe.Tax != nil && client.app.Stripe.Tax.Dynamic != nil {
	// 	dynamicTaxRates = client.app.Stripe.Tax.Dynamic
	// }

	customer, err := client.customerDAO.GetByEmail(userAccount.GetEmail(), common.CONSISTENCY_LOCAL)
	if err != nil {
		client.app.Logger.Error(err)
		return EmptyPaymentIntentResponse, err
	}

	// Make sure the user doesn't already have an outstanding invoice
	currentInvoices := client.ListInvoices(customer.ProcessorID, stripe.String("open"))
	if len(currentInvoices) > 0 {
		return EmptyPaymentIntentResponse, ErrInvoiceOutstanding
	}

	var shippingDetails *stripe.InvoiceShippingDetailsParams
	if customer.Shipping != nil {
		shippingDetails = &stripe.InvoiceShippingDetailsParams{
			Name:  stripe.String(customer.Shipping.Name),
			Phone: stripe.String(customer.Shipping.Phone),
			Address: &stripe.AddressParams{
				Line1:      stripe.String(customer.Shipping.Address.Line1),
				Line2:      stripe.String(customer.Shipping.Address.Line2),
				City:       stripe.String(customer.Shipping.Address.City),
				State:      stripe.String(customer.Shipping.Address.State),
				PostalCode: stripe.String(customer.Shipping.Address.PostalCode),
				Country:    stripe.String(customer.Shipping.Address.Country)}}
	}

	// Create a new invoice
	invoiceParams := &stripe.InvoiceParams{
		Customer:        stripe.String(customer.ProcessorID),
		Description:     stripe.String(createInvoiceRequest.Description),
		AutoAdvance:     stripe.Bool(false),
		ShippingDetails: shippingDetails}

	// Apply tax rate(s)
	if fixedTaxRates != nil {
		invoiceParams.DefaultTaxRates = fixedTaxRates
	} else {
		invoiceParams.AutomaticTax = &stripe.InvoiceAutomaticTaxParams{
			Enabled: stripe.Bool(true)}
	}

	newInvoice, err := invoice.New(invoiceParams)
	if err != nil {
		client.app.Logger.Error(err)
		return EmptyPaymentIntentResponse, err
	}

	// Add products to the invoice
	for _, product := range createInvoiceRequest.Products {
		invoiceItem := &stripe.InvoiceItemParams{
			Customer:    stripe.String(customer.ProcessorID),
			Invoice:     stripe.String(newInvoice.ID),
			Description: stripe.String(product.Name),
			UnitAmount:  &product.Price,
			Quantity:    &product.Quantity}

		if _, err := invoiceitem.New(invoiceItem); err != nil {
			client.app.Logger.Error(err)
			return EmptyPaymentIntentResponse, err
		}
	}

	// Finalize the invoice
	params := &stripe.InvoiceFinalizeInvoiceParams{AutoAdvance: stripe.Bool(false)}
	if newInvoice, err = invoice.FinalizeInvoice(newInvoice.ID, params); err != nil {
		return EmptyPaymentIntentResponse, err
	}

	// Email the invoice to the customer if configured to do so
	if newInvoice.CollectionMethod == stripe.InvoiceCollectionMethodSendInvoice {
		sentInvoice, err := invoice.SendInvoice(newInvoice.ID, &stripe.InvoiceSendInvoiceParams{})
		if err != nil {
			client.app.Logger.Error(err)
			return EmptyPaymentIntentResponse, err
		}
		return client.createPaymentIntentResponse(customer, newInvoice.ID,
			newInvoice.PaymentIntent.ID, sentInvoice.PaymentIntent.ClientSecret)
	}

	// Look up the payment intent to retreive the secret key (not included in the invoice PaymentIntent)
	paymentIntentResponse, err := client.GetPaymentIntent(newInvoice.PaymentIntent.ID)
	paymentIntentResponse.Customer = customer
	paymentIntentResponse.InvoiceID = newInvoice.ID
	return paymentIntentResponse, err
}

func (client StripeClient) PayInvoice(invoiceID string) (*Invoice, error) {
	params := &stripe.InvoicePayParams{}
	_invoice, err := invoice.Pay(invoiceID, params)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	if !_invoice.Paid {
		return nil, ErrInvoiceNotPaid
	}
	return client.mapStripeInvoiceToInvoice(_invoice), err
}

func (client *StripeClient) GetPaymentIntent(id string) (PaymentIntentResponse, error) {
	// params := &stripe.PaymentIntentParams{
	// 	ClientSecret: stripe.String(secret),
	// }
	paymentIntent, err := paymentintent.Get(id, &stripe.PaymentIntentParams{})
	if err != nil {
		client.app.Logger.Error(err)
		return PaymentIntentResponse{}, nil
	}
	customer := &config.Customer{ProcessorID: paymentIntent.Customer.ID}
	return client.createPaymentIntentResponse(customer, "", paymentIntent.ID, paymentIntent.ClientSecret)
}

func (client *StripeClient) CreatePaymentIntent(createPaymentIntentRequest CreatePaymentIntentRequest) (PaymentIntentResponse, error) {

	var c *stripe.Customer
	var err error
	cparams := &stripe.CustomerParams{}

	if createPaymentIntentRequest.CustomerID != "" {
		c, err = customer.Get(createPaymentIntentRequest.CustomerID, cparams)
	} else {
		c, _ = customer.New(cparams) // Creates a new customer
	}

	if err != nil {
		client.app.Logger.Error(err)
		return PaymentIntentResponse{}, err
	}

	amount64, err := strconv.ParseInt(createPaymentIntentRequest.Amount, 10, 64)
	if err != nil {
		client.app.Logger.Error(err)
		return PaymentIntentResponse{}, err
	}

	var receiptEmail = stripe.String(c.Email)
	if c.Email == "" {
		receiptEmail = nil
	}
	piparams := &stripe.PaymentIntentParams{
		Amount:           stripe.Int64(amount64),
		Currency:         stripe.String(createPaymentIntentRequest.CurrencyCode),
		Customer:         stripe.String(c.ID),
		Confirm:          stripe.Bool(false),
		ReceiptEmail:     receiptEmail,
		SetupFutureUsage: stripe.String("off_session"),
		// In the latest version of the API, specifying the `automatic_payment_methods` parameter
		// is optional because Stripe enables its functionality by default.
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
	}
	paymentIntent, err := paymentintent.New(piparams)
	if err != nil {
		client.app.Logger.Error(err)
		return PaymentIntentResponse{}, err
	}

	customer := &config.Customer{ProcessorID: c.ID}
	return client.createPaymentIntentResponse(customer, "", paymentIntent.ID, paymentIntent.ClientSecret)
}

func (client *StripeClient) CreateSubscription(csubscriptionConfig *Subscription) (*Subscription, error) {
	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(csubscriptionConfig.Customer.ProcessorID),
		Items: []*stripe.SubscriptionItemsParams{{
			Price: stripe.String(csubscriptionConfig.PriceID)}},
		PaymentBehavior: stripe.String("default_incomplete"),
	}
	subscriptionParams.AddExpand("latest_invoice.payment_intent")
	stripeSubscription, err := subscription.New(subscriptionParams)
	customer := client.mapStripeCustomerToCustomerConfig(stripeSubscription.Customer)
	customer.ID = csubscriptionConfig.Customer.ID
	return &Subscription{
		ID:              stripeSubscription.ID,
		Customer:        customer,
		LatestInvoiceID: stripeSubscription.LatestInvoice.ID}, err
}

func (client *StripeClient) UpdateSubscription(_subscription *Subscription) (*Subscription, error) {

	// Fetch the subscription to access the related subscription item's ID
	// that will be updated. In practice, you might want to store the
	// Subscription Item ID in your database to avoid this API call.
	s, err := subscription.Get(_subscription.ID, &stripe.SubscriptionParams{})
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	var defaultPaymentMethodID *string
	if _subscription.PaymentIntentID != "" {

		params := &stripe.PaymentIntentParams{}
		paymentIntentResult, err := paymentintent.Get(_subscription.PaymentIntentID, params)
		if err != nil {
			client.app.Logger.Error(err)
			return nil, err
		}

		defaultPaymentMethodID = stripe.String(paymentIntentResult.PaymentMethod.ID)
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{{
			ID:    stripe.String(s.Items.Data[0].ID),
			Price: stripe.String(_subscription.PriceID),
		}}}
	if defaultPaymentMethodID != nil {
		params.DefaultPaymentMethod = defaultPaymentMethodID
	}

	updatedSubscription, err := subscription.Update(_subscription.ID, params)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	return &Subscription{
		ID:              updatedSubscription.ID,
		Description:     updatedSubscription.Description,
		PriceID:         _subscription.PriceID,
		AutomaticTax:    updatedSubscription.AutomaticTax.Enabled,
		DefaultTaxRates: client.mapStripeTaxRatesToTaxRates(updatedSubscription.DefaultTaxRates),
		LatestInvoiceID: updatedSubscription.LatestInvoice.ID,
		PaymentIntentID: _subscription.PaymentIntentID,
	}, nil
}

func (client *StripeClient) CancelSubscription(subscriptionID string) (*Subscription, error) {
	_, err := subscription.Cancel(subscriptionID, &stripe.SubscriptionCancelParams{})
	return nil, err
}

func (client *StripeClient) ListSubscriptions(processorID string) []*Subscription {
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(processorID),
		Status:   stripe.String("all")}
	params.AddExpand("data.default_payment_method")
	result := subscription.List(params)
	subscriptions := make([]*Subscription, 0)
	for result.Iter.Next() {
		item := result.Iter.Current().(*stripe.Subscription)
		subscriptions = append(subscriptions, &Subscription{
			ID:           item.ID,
			Description:  item.Description,
			AutomaticTax: item.AutomaticTax.Enabled,
			//PriceID: item.(*stripe.Subscription).,
		})
	}
	return subscriptions
}

func (client *StripeClient) CreateCheckoutSession() (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		SuccessURL:         stripe.String("https://example.com/success"),
		CancelURL:          stripe.String("https://example.com/cancel"),
		Mode:               stripe.String("payment"),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{{
			Quantity: stripe.Int64(2),
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency:    stripe.String("usd"),
				UnitAmount:  stripe.Int64(10000),
				TaxBehavior: stripe.String("exclusive"),
				Product:     stripe.String("prod_q23fxaHasd"),
			},
		}},
	}
	return session.New(params)
}

func (client *StripeClient) InvoicePreview(customerID, subscriptionID, priceID string) (*Invoice, error) {

	_subscription, err := subscription.Get(subscriptionID, &stripe.SubscriptionParams{})
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	// params := &stripe.InvoiceParams{
	// 	Customer:     stripe.String(customerID),
	// 	Subscription: stripe.String(subscriptionID),
	// }

	stripeInvoice, err := invoice.Upcoming(&stripe.InvoiceUpcomingParams{
		Customer:     stripe.String(customerID),
		Subscription: stripe.String(_subscription.ID)})
	// SubscriptionItems: []*stripe.SubscriptionItemsParams{{
	// 	ID:    stripe.String(_subscription.ID),
	// 	Price: stripe.String(priceID),
	// }}})
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	return &Invoice{
		ID:              stripeInvoice.ID,
		Description:     stripeInvoice.Description,
		AutoAdvance:     stripeInvoice.AutoAdvance,
		DefaultTaxRates: client.mapStripeTaxRatesToTaxRates(stripeInvoice.DefaultTaxRates)}, nil

}

func (client *StripeClient) GetPaymentMethod(paymentMethodID string) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodParams{}
	result, err := paymentmethod.Get(paymentMethodID, params)
	return result, err
}

func (client *StripeClient) GetPaymentMethods(processorID string) []*stripe.PaymentMethod {
	paymentMethods := make([]*stripe.PaymentMethod, 0)
	params := &stripe.CustomerListPaymentMethodsParams{
		Customer: stripe.String(processorID),
		Type:     stripe.String(string(stripe.PaymentMethodTypeCard)),
	}
	params.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
	result := customer.ListPaymentMethods(params)
	for result.Iter.Next() {
		item := result.Iter.Current()
		paymentMethods = append(paymentMethods, item.(*stripe.PaymentMethod))
	}
	return paymentMethods
}

func (client *StripeClient) CreateCardPaymentMethod(card *CreditCard) (string, error) {
	params := &stripe.PaymentMethodParams{
		Type: stripe.String(string(stripe.PaymentMethodTypeCard)),
		Card: &stripe.PaymentMethodCardParams{
			Number:   stripe.String(card.Number),
			ExpMonth: stripe.Int64(card.ExpMonth),
			ExpYear:  stripe.Int64(card.ExpYear),
			CVC:      stripe.String(card.CVC)}}

	paymentMethod, err := paymentmethod.New(params)
	if err != nil {
		client.app.Logger.Error(err)
	}
	return paymentMethod.ID, err
}

func (client *StripeClient) AttachAndSetDefaultPaymentMethod(
	setDefaultPaymentMethodRequest *SetDefaultPaymentMethodRequest) (*config.Customer, error) {
	_, err := client.AttachPaymentMethod(&AttachPaymentMethodRequest{
		ProcessorID:     setDefaultPaymentMethodRequest.ProcessorID,
		PaymentMethodID: setDefaultPaymentMethodRequest.PaymentMethodID})
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	return client.SetDefaultPaymentMethod(setDefaultPaymentMethodRequest)
}

func (client *StripeClient) AttachPaymentMethod(attachPaymentRequest *AttachPaymentMethodRequest) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(attachPaymentRequest.ProcessorID)}
	stripePaymentMethod, err := paymentmethod.Attach(attachPaymentRequest.PaymentMethodID, params)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}
	return stripePaymentMethod, nil
}

func (client *StripeClient) SetDefaultPaymentMethod(
	setDefaultPaymentMethodRequest *SetDefaultPaymentMethodRequest) (*config.Customer, error) {

	var defaultPaymentMethod *stripe.PaymentMethod = nil
	stripeCustomer, err := customer.Get(setDefaultPaymentMethodRequest.ProcessorID, nil)
	if err != nil {
		client.app.Logger.Error(err)
		return nil, err
	}

	if setDefaultPaymentMethodRequest.PaymentMethodID == "" {
		// This is a hack to workaround confirming the SetupIntent using the Android SDK.
		// There is no way to define a callback for stripe.confirmSetupIntent and there
		// appears to be a race condition somewhere because an immediate call to GetCustomer
		// does not show the newly added payment method. It could be network or system latency,
		// the confirmSetupIntent call might be async, or some other problem.
		// This condition handles setting a newly added card during a SetupIntent as the default
		// knowing what the payment method id is.
		paymentMethods := client.GetPaymentMethods(stripeCustomer.ID)
		if len(paymentMethods) == 0 {
			client.app.Logger.Error(ErrMissingPaymentMethod)
			return nil, ErrMissingPaymentMethod
		}
		for _, method := range paymentMethods {
			if method.Card.Checks.AddressPostalCodeCheck != "pass" ||
				method.Card.Checks.CVCCheck != "pass" {
				continue
			} else {
				defaultPaymentMethod = method
				break
			}
		}
		if defaultPaymentMethod == nil {
			client.app.Logger.Error(ErrMissingPaymentMethod)
			return nil, ErrMissingPaymentMethod
		}
	} else {
		defaultPaymentMethod, err = client.GetPaymentMethod(setDefaultPaymentMethodRequest.PaymentMethodID)
		if err != nil {
			client.app.Logger.Error(err)
			return nil, err
		}
		if defaultPaymentMethod == nil {
			client.app.Logger.Error(ErrMissingPaymentMethod)
			return nil, ErrMissingPaymentMethod
		}
	}

	stripeCustomer.InvoiceSettings = &stripe.CustomerInvoiceSettings{
		DefaultPaymentMethod: &stripe.PaymentMethod{
			ID: defaultPaymentMethod.ID}}

	customerConfigWithDefaultPaymentMethod := client.mapStripeCustomerToCustomerConfig(stripeCustomer)
	customerConfigWithDefaultPaymentMethod.ID = setDefaultPaymentMethodRequest.CustomerID
	customerConfigWithDefaultPaymentMethod.PaymentMethodLast4 = defaultPaymentMethod.Card.Last4

	return client.UpdateCustomer(customerConfigWithDefaultPaymentMethod)
}

func (client *StripeClient) GetTaxRates() []*TaxRate {
	taxRates := make([]*TaxRate, 0)
	params := &stripe.TaxRateListParams{}
	params.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
	result := taxrate.List(params)
	for result.Iter.Next() {
		item := result.Iter.Current()
		if !item.(*stripe.TaxRate).Active {
			continue
		}
		taxRates = append(taxRates, &TaxRate{
			ID:           item.(*stripe.TaxRate).ID,
			DisplayName:  item.(*stripe.TaxRate).DisplayName,
			Description:  item.(*stripe.TaxRate).Description,
			State:        item.(*stripe.TaxRate).State,
			Country:      item.(*stripe.TaxRate).Country,
			Inclusive:    item.(*stripe.TaxRate).Inclusive,
			Jurisdiction: item.(*stripe.TaxRate).Jurisdiction,
			Percentage:   item.(*stripe.TaxRate).Percentage})

		client.app.Logger.Infof("%+v", taxRates)
	}
	return taxRates
}

func (client *StripeClient) createPaymentIntentResponse(customer *config.Customer, invoiceID,
	paymentIntentID, clientSecret string) (PaymentIntentResponse, error) {

	ephemeralKey, err := client.createEphemeralKey(customer.ProcessorID)
	if err != nil {
		client.app.Logger.Error(err)
		return PaymentIntentResponse{}, err
	}

	return PaymentIntentResponse{
		Customer:       customer,
		InvoiceID:      invoiceID,
		PaymentIntent:  paymentIntentID,
		ClientSecret:   clientSecret,
		EphemeralKey:   ephemeralKey.Secret,
		PublishableKey: client.app.Stripe.Key.Publishable}, nil

}

func (client *StripeClient) createEphemeralKey(processorID string) (*stripe.EphemeralKey, error) {
	ekparams := &stripe.EphemeralKeyParams{
		Customer:      stripe.String(processorID),
		StripeVersion: stripe.String("2020-08-27"),
	}
	return ephemeralkey.New(ekparams)
}

func (client *StripeClient) mapCustomerConfigToStripeCustomerParams(customerConfig *config.Customer) *stripe.CustomerParams {

	var address *stripe.AddressParams
	var shipping *stripe.CustomerShippingParams

	if customerConfig.Address != nil {
		address = &stripe.AddressParams{
			Line1:      &customerConfig.Address.Line1,
			Line2:      &customerConfig.Address.Line2,
			City:       &customerConfig.Address.City,
			State:      &customerConfig.Address.State,
			PostalCode: &customerConfig.Address.PostalCode,
			Country:    &customerConfig.Address.Country,
		}
	}

	if customerConfig.Shipping != nil {
		shipping = &stripe.CustomerShippingParams{
			Name:  &customerConfig.Shipping.Name,
			Phone: &customerConfig.Shipping.Phone,
			Address: &stripe.AddressParams{
				Line1:      &customerConfig.Shipping.Address.Line1,
				Line2:      &customerConfig.Shipping.Address.Line2,
				City:       &customerConfig.Shipping.Address.City,
				State:      &customerConfig.Shipping.Address.State,
				PostalCode: &customerConfig.Shipping.Address.PostalCode,
				Country:    &customerConfig.Shipping.Address.Country}}
	}

	params := &stripe.CustomerParams{
		Name:        stripe.String(customerConfig.Name),
		Email:       stripe.String(customerConfig.Email),
		Description: &customerConfig.Description,
		Phone:       &customerConfig.Phone,
		Address:     address,
		Shipping:    shipping}

	if customerConfig.PaymentMethodID != "" {
		params.InvoiceSettings = &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(customerConfig.PaymentMethodID)}
	}

	return params
}

func (client *StripeClient) mapStripeCustomerToCustomerConfig(stripeCustomer *stripe.Customer) *config.Customer {

	var customer = config.NewCustomer()
	var billingAddress *config.Address
	var shippingAddress *config.ShippingAddress

	paymentMethodID := ""
	paymentMethodLast4 := ""
	if stripeCustomer.InvoiceSettings != nil && stripeCustomer.InvoiceSettings.DefaultPaymentMethod != nil {
		paymentMethodID = stripeCustomer.InvoiceSettings.DefaultPaymentMethod.ID
		if stripeCustomer.InvoiceSettings.DefaultPaymentMethod.Card != nil {
			paymentMethodLast4 = stripeCustomer.InvoiceSettings.DefaultPaymentMethod.Card.Last4
		}
	}

	if stripeCustomer.Address != nil {
		billingAddress = &config.Address{
			Line1:      stripeCustomer.Address.Line1,
			Line2:      stripeCustomer.Address.Line2,
			City:       stripeCustomer.Address.City,
			State:      stripeCustomer.Address.State,
			PostalCode: stripeCustomer.Address.PostalCode,
			Country:    stripeCustomer.Address.Country}
	}

	if stripeCustomer.Shipping != nil {
		shippingAddress = &config.ShippingAddress{
			Name:  stripeCustomer.Shipping.Name,
			Phone: stripeCustomer.Shipping.Phone,
			Address: &config.Address{
				Line1:      stripeCustomer.Shipping.Address.Line1,
				Line2:      stripeCustomer.Shipping.Address.Line2,
				City:       stripeCustomer.Shipping.Address.City,
				State:      stripeCustomer.Shipping.Address.State,
				PostalCode: stripeCustomer.Shipping.Address.PostalCode,
				Country:    stripeCustomer.Shipping.Address.Country}}
	}

	customer.ProcessorID = stripeCustomer.ID
	customer.Name = stripeCustomer.Name
	customer.Email = stripeCustomer.Email
	customer.Phone = stripeCustomer.Phone
	customer.Address = billingAddress
	customer.Shipping = shippingAddress
	customer.PaymentMethodID = paymentMethodID
	customer.PaymentMethodLast4 = paymentMethodLast4

	return customer
}

func (client *StripeClient) mapStripeTaxRatesToTaxRates(stripeTaxRates []*stripe.TaxRate) []*TaxRate {
	taxRates := make([]*TaxRate, 0)
	for _, rate := range stripeTaxRates {
		taxRates = append(taxRates, &TaxRate{
			ID:           rate.ID,
			DisplayName:  rate.DisplayName,
			Description:  rate.Description,
			State:        rate.State,
			Country:      rate.Country,
			Inclusive:    rate.Inclusive,
			Jurisdiction: rate.Jurisdiction,
			Percentage:   rate.Percentage})
	}
	return taxRates
}

func (client *StripeClient) createTestCardPaymentMethod() *stripe.PaymentMethodParams {
	return &stripe.PaymentMethodParams{
		Type: stripe.String(string(stripe.PaymentMethodTypeCard)),
		Card: &stripe.PaymentMethodCardParams{
			Number:   stripe.String("4242424242424242"),
			ExpMonth: stripe.Int64(1),
			ExpYear:  stripe.Int64(2050),
			CVC:      stripe.String("123")}}
}

func (client *StripeClient) mapStripeInvoiceToInvoice(stripeInvoice *stripe.Invoice) *Invoice {
	return &Invoice{
		ID:              stripeInvoice.ID,
		Description:     stripeInvoice.Description,
		AutoAdvance:     stripeInvoice.AutoAdvance,
		PaymentIntentID: stripeInvoice.PaymentIntent.ID,
		DefaultTaxRates: client.mapStripeTaxRatesToTaxRates(stripeInvoice.DefaultTaxRates)}
}

func (client *StripeClient) hasDefaultPaymentSource(stripeCustomer *stripe.Customer) bool {
	return stripeCustomer.InvoiceSettings != nil && stripeCustomer.InvoiceSettings.DefaultPaymentMethod != nil
}
