package shoppingcart

import (
	"errors"
	"strconv"

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
	"github.com/stripe/stripe-go/v78/price"
	"github.com/stripe/stripe-go/v78/product"
	"github.com/stripe/stripe-go/v78/taxrate"
	"gorm.io/gorm"
)

var ErrInvoiceNotPaid = errors.New("invoice not paid")
var STRIPE_DEFAULT_LIMIT = int64(10)

type ShoppingCartService interface {
	GetPaymentMethods(customerID string) []*stripe.PaymentMethod
	GetProducts() []Product
	CreatePaymentIntent(createPaymentIntentRequest CreatePaymentIntentRequest) (PaymentIntentResponse, error)
	GetPaymentIntent(id string) (PaymentIntentResponse, error)
	GetCustomer(id uint64) (*config.Customer, error)
	CreateCustomer(customerConfig *config.Customer) (*config.Customer, error)
	UpdateCustomer(customerConfig *config.Customer) (*config.Customer, error)
	CreateInvoice(userAccount common.UserAccount, createInvoiceRequest CreateInvoiceRequest) (PaymentIntentResponse, error)
	PayInvoice(invoiceID string) (bool, error)
	GetTaxRates() []*stripe.TaxRate
}

type StripeClient struct {
	app         *app.App
	params      *config.Stripe
	customerDAO dao.CustomerDAO
	ShoppingCartService
}

func NewStripeService(app *app.App, customerDAO dao.CustomerDAO) ShoppingCartService {
	stripe.Key = app.Stripe.Key.Secret
	return &StripeClient{
		app:         app,
		params:      app.Stripe,
		customerDAO: customerDAO}
}

func (client *StripeClient) GetProducts() []Product {
	products := make([]Product, 0)
	params := &stripe.ProductListParams{}
	params.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
	result := product.List(params)
	for result.Iter.Next() {
		item := result.Iter.Current()
		if !item.(*stripe.Product).Active {
			continue
		}
		priceParams := &stripe.PriceListParams{}
		priceParams.Product = stripe.String(item.(*stripe.Product).ID)
		priceParams.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
		result := price.List(priceParams)
		for result.Iter.Next() {
			price := result.Iter.Current()
			if !price.(*stripe.Price).Active {
				continue
			}
			imageUrl := ""
			if len(item.(*stripe.Product).Images) > 0 {
				imageUrl = item.(*stripe.Product).Images[0]
			}
			products = append(products, Product{
				Id:          item.(*stripe.Product).ID,
				Name:        item.(*stripe.Product).Name,
				Description: item.(*stripe.Product).Description,
				ImageUrl:    imageUrl,
				Price:       price.(*stripe.Price).UnitAmount})
		}
	}
	return products
}

func (client *StripeClient) GetCustomer(id uint64) (*config.Customer, error) {

	customerConfig, err := client.customerDAO.Get(id, common.CONSISTENCY_LOCAL)
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
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

	return client.mapStripeCustomerToCustomerConfig(stripeCustomer), nil
}

func (client *StripeClient) UpdateCustomer(customerConfig *config.Customer) (*config.Customer, error) {

	customerParams := client.mapCustomerConfigToStripeCustomer(customerConfig)
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

func (client *StripeClient) CreateCustomer(customerConfig *config.Customer) (*config.Customer, error) {

	customerParams := client.mapCustomerConfigToStripeCustomer(customerConfig)
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

	// Try to look up the customer from the database (the client does not store the processor ID / stripe customer ID)
	customer, err := client.customerDAO.GetByEmail(userAccount.GetEmail(), common.CONSISTENCY_LOCAL)
	if err != nil {
		client.app.Logger.Error(err)
		return EmptyPaymentIntentResponse, err
	}

	if customer.ID == 0 {
		// If the customer doesn't exist, this is a new customer. Create a stripe
		// customer and save them to the database.
		customserConfig := &config.Customer{
			Name:  userAccount.GetEmail(),
			Email: userAccount.GetEmail()}
		customer, err = client.CreateCustomer(customserConfig)
		if err != nil {
			client.app.Logger.Error(err)
			return EmptyPaymentIntentResponse, err
		}
		if err = client.customerDAO.Save(customer); err != nil {
			client.app.Logger.Error(err)
			return EmptyPaymentIntentResponse, err
		}
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

	// Look up the payment intent to retreive the secret key (not included in the invoice payment intent)
	paymentIntentResponse, err := client.GetPaymentIntent(newInvoice.PaymentIntent.ID)
	paymentIntentResponse.Customer = customer
	paymentIntentResponse.InvoiceID = newInvoice.ID
	return paymentIntentResponse, err
}

func (client StripeClient) PayInvoice(invoiceID string) (bool, error) {
	params := &stripe.InvoicePayParams{}
	result, err := invoice.Pay(invoiceID, params)
	if err != nil {
		client.app.Logger.Error(err)
		return false, err
	}
	if !result.Paid {
		return false, ErrInvoiceNotPaid
	}
	return true, err
}

func (client *StripeClient) GetPaymentMethods(customerID string) []*stripe.PaymentMethod {
	paymentMethods := make([]*stripe.PaymentMethod, 0)
	params := &stripe.CustomerListPaymentMethodsParams{
		Customer: stripe.String(customerID),
	}
	params.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
	result := customer.ListPaymentMethods(params)
	for result.Iter.Next() {
		item := result.Iter.Current()

		client.app.Logger.Infof("%+v", item)
	}
	return paymentMethods
}

func (client *StripeClient) GetTaxRates() []*stripe.TaxRate {
	taxRates := make([]*stripe.TaxRate, 0)
	params := &stripe.TaxRateListParams{}
	params.Limit = stripe.Int64(STRIPE_DEFAULT_LIMIT)
	result := taxrate.List(params)
	for result.Iter.Next() {
		item := result.Iter.Current()
		if !item.(*stripe.TaxRate).Active {
			continue
		}
		taxRates = append(taxRates, item.(*stripe.TaxRate))
		client.app.Logger.Infof("%+v", item.(*stripe.TaxRate))
	}
	return taxRates
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

func (Client *StripeClient) CreateCheckoutSession() (*stripe.CheckoutSession, error) {
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

// func (client *StripeClient) CreateTaxRates() error {

// 	stripeTax := client.app.Stripe.Tax
// 	if stripeTax == nil {
// 		return
// 	}
// 	for _, rate :- range client.app.Stripe.Tax.Rates {

// 	}

// 	params := &stripe.TaxRateParams{
// 		DisplayName: stripe.String("VAT"),
// 		Description: stripe.String("VAT Germany"),
// 		Percentage: stripe.Float64(16),
// 		Jurisdiction: stripe.String("DE"),
// 		Inclusive: stripe.Bool(false),
// 	  };
// 	  result, err := taxrate.New(params);
// }

func (client *StripeClient) createPaymentIntentResponse(customer *config.Customer, invoiceID,
	paymentIntentID, clientSecret string) (PaymentIntentResponse, error) {

	ekparams := &stripe.EphemeralKeyParams{
		Customer:      stripe.String(customer.ProcessorID),
		StripeVersion: stripe.String("2020-08-27"),
	}
	ephemeralKey, err := ephemeralkey.New(ekparams)
	if err != nil {
		client.app.Logger.Error(err)
		return PaymentIntentResponse{}, nil
	}
	return PaymentIntentResponse{
		Customer:       customer,
		InvoiceID:      invoiceID,
		PaymentIntent:  paymentIntentID,
		ClientSecret:   clientSecret,
		EphemeralKey:   ephemeralKey.Secret,
		PublishableKey: client.app.Stripe.Key.Publishable}, nil

}

func (client *StripeClient) mapCustomerConfigToStripeCustomer(customerConfig *config.Customer) *stripe.CustomerParams {

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

	return &stripe.CustomerParams{
		Description: &customerConfig.Description,
		Name:        stripe.String(customerConfig.Name),
		Email:       stripe.String(customerConfig.Email),
		Phone:       &customerConfig.Phone,
		Address:     address,
		Shipping:    shipping}
}

func (client *StripeClient) mapStripeCustomerToCustomerConfig(stripeCustomer *stripe.Customer) *config.Customer {

	var customer = &config.Customer{}
	var billingAddress *config.Address
	var shippingAddress *config.ShippingAddress

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

	customer.Name = stripeCustomer.Name
	customer.Email = stripeCustomer.Email
	customer.Phone = stripeCustomer.Phone
	customer.Address = billingAddress
	customer.Shipping = shippingAddress

	return customer
}
