package shoppingcart

import (
	"encoding/json"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v78"
)

func TestGetProducts(t *testing.T) {

	service := NewStripeService(CurrentTest.app, nil)
	products := service.GetProducts()

	jsonData, err := json.Marshal(products)
	assert.Nil(t, err)
	assert.NotNil(t, jsonData)

	var jsonProducts []Product
	err = json.Unmarshal(jsonData, &jsonProducts)
	assert.Nil(t, err)
	assert.Greater(t, len(jsonProducts), 0)

	for _, product := range products {
		if product.Metadata != nil {
			for k, v := range product.Metadata {
				assert.NotEqual(t, "subscription", v)
				assert.NotEqual(t, "visible", k)
			}
		}
	}

	assert.Greater(t, len(products), 0)
}

func TestGetPublishableKey(t *testing.T) {
	service := NewStripeService(CurrentTest.app, nil)
	assert.Equal(t, service.GetPublishableKey(), CurrentTest.app.Stripe.Key.Publishable)
}

func TestCreatePaymentIntent(t *testing.T) {
	service := NewStripeService(CurrentTest.app, nil)
	createPaymentIntentRequest := CreatePaymentIntentRequest{
		Amount:       "123",
		CurrencyCode: string(stripe.CurrencyUSD),
	}
	paymentIntent, err := service.CreatePaymentIntent(createPaymentIntentRequest)
	assert.Nil(t, err)
	assert.Equal(t, paymentIntent.PublishableKey, CurrentTest.app.Stripe.Key.Publishable)
}

func TestListSetupIntents(t *testing.T) {
	service := NewStripeService(CurrentTest.app, nil)
	setupIntents := service.ListSetupIntents()
	assert.Greater(t, 0, len(setupIntents))
	assert.NotEmpty(t, setupIntents)
}

func TestGetPaymentIntent(t *testing.T) {
	service := NewStripeService(CurrentTest.app, nil)
	createPaymentIntentRequest := CreatePaymentIntentRequest{

		Amount:       "123",
		CurrencyCode: string(stripe.CurrencyUSD),
	}
	paymentIntent, err := service.CreatePaymentIntent(createPaymentIntentRequest)
	assert.Nil(t, err)
	assert.Equal(t, paymentIntent.PublishableKey, CurrentTest.app.Stripe.Key.Publishable)
	assert.Equal(t, uint64(0), paymentIntent.Customer.ID)
	assert.NotEmpty(t, paymentIntent.PaymentIntent)
	assert.NotEmpty(t, paymentIntent.ClientSecret)
	assert.NotEmpty(t, paymentIntent.EphemeralKey)
	assert.NotEmpty(t, paymentIntent.PublishableKey)

	persistedPaymentIntent, err := service.GetPaymentIntent(paymentIntent.PaymentIntent)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), persistedPaymentIntent.Customer.ID)
	assert.NotEmpty(t, persistedPaymentIntent.PaymentIntent)
	assert.NotEmpty(t, persistedPaymentIntent.ClientSecret)
	assert.NotEmpty(t, persistedPaymentIntent.EphemeralKey)
	assert.NotEmpty(t, persistedPaymentIntent.PublishableKey)
}

func TestGetOrCreateCustomerWithEphemeralKey(t *testing.T) {
	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)
	customer := createTestCustomer()
	customerWithEphemeralKeyResponse, err := service.GetOrCreateCustomerWithEphemeralKey(
		&model.UserStruct{
			ID:    customer.ID,
			Email: customer.Email})
	assert.Nil(t, err)
	assert.Equal(t, customerWithEphemeralKeyResponse.Customer.Name, "SetupIntent")
	assert.NotEmpty(t, customerWithEphemeralKeyResponse.Customer.ProcessorID)
	assert.NotEmpty(t, customerWithEphemeralKeyResponse.EphemeralKey)
}

func TestCreateCustomer(t *testing.T) {

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)
	customer := createTestCustomer()

	stripeCustomer, err := service.CreateCustomer(customer)
	assert.Nil(t, err)
	assert.Equal(t, stripeCustomer.Name, customer.Name)
	assert.Equal(t, stripeCustomer.Email, customer.Email)
	assert.NotEmpty(t, stripeCustomer.ProcessorID)

	page1, err := customerDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(page1.Entities))
	assert.Equal(t, customer.Name, page1.Entities[0].Name)
	assert.Equal(t, customer.Email, page1.Entities[0].Email)
	assert.Equal(t, customer.ProcessorID, stripeCustomer.ProcessorID)
	assert.NotEmpty(t, customer.ProcessorID, page1.Entities[0].ProcessorID)

	assert.Equal(t, customer.Address, page1.Entities[0].Address)
	assert.Equal(t, customer.Shipping, page1.Entities[0].Shipping)
}

func TestUpdateCustomer(t *testing.T) {

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)
	customer := createTestCustomer()

	customerConfig, err := service.CreateCustomer(customer)
	assert.Nil(t, err)
	assert.Equal(t, customerConfig.Name, customer.Name)
	assert.Equal(t, customerConfig.Email, customer.Email)
	assert.NotEmpty(t, customerConfig.ProcessorID)

	newCustomerName := "New Test Customer"
	customerConfig.Name = newCustomerName
	customerConfig.Shipping.Name = "test"
	customerConfig.Shipping.Phone = "1111111111"
	customerConfig.Shipping.Address.Line1 = "test"
	customerConfig.Shipping.Address.Line2 = "test"
	customerConfig.Shipping.Address.City = "test"
	customerConfig.Shipping.Address.State = "test"
	customerConfig.Shipping.Address.PostalCode = "11111"
	customerConfig.Shipping.Address.Country = "test"
	updatedCustomerConfig, err := service.UpdateCustomer(customerConfig)
	assert.Nil(t, err)
	assert.Equal(t, updatedCustomerConfig.Name, newCustomerName)
	assert.Equal(t, updatedCustomerConfig.Email, customer.Email)
	assert.NotEmpty(t, updatedCustomerConfig.ProcessorID)
	assert.Equal(t, updatedCustomerConfig.Shipping.Address.Line1, customer.Shipping.Address.Line1)
	assert.Equal(t, updatedCustomerConfig.Shipping.Address.Line2, customer.Shipping.Address.Line2)
	assert.Equal(t, updatedCustomerConfig.Shipping.Address.City, customer.Shipping.Address.City)
	assert.Equal(t, updatedCustomerConfig.Shipping.Address.State, customer.Shipping.Address.State)
	assert.Equal(t, updatedCustomerConfig.Shipping.Address.PostalCode, customer.Shipping.Address.PostalCode)
	assert.Equal(t, updatedCustomerConfig.Shipping.Address.Country, customer.Shipping.Address.Country)

	page1, err := customerDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(page1.Entities))
	assert.Equal(t, newCustomerName, page1.Entities[0].Name)
	assert.Equal(t, customer.Email, page1.Entities[0].Email)
	assert.NotEmpty(t, customerConfig.ProcessorID, page1.Entities[0].ProcessorID)
	assert.NotEmpty(t, updatedCustomerConfig.ProcessorID, page1.Entities[0].ProcessorID)

	assert.Equal(t, customer.Address, page1.Entities[0].Address)
	assert.Equal(t, customer.Shipping, page1.Entities[0].Shipping)
}

func TestGetPaymentMethods(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	customerConfig, err := service.GetCustomer(1)
	assert.Nil(t, err)
	assert.NotNil(t, customerConfig)

	paymentMethods := service.GetPaymentMethods(customerConfig.ProcessorID)

	assert.NotEmpty(t, paymentMethods)
}

// NOTE: This method requires the payment method to already be saved to the customer
//
//	account, which must be submitted from the client. After the client has submitted
//	the payment method, the server can look it up.
func TestGetPaymentMethod(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	customerConfig, err := service.GetCustomer(TEST_CUSTOMER_ID)
	assert.Nil(t, err)
	assert.NotNil(t, customerConfig)

	stripePaymentMethonds := service.GetPaymentMethods(customerConfig.ProcessorID)
	assert.Greater(t, len(stripePaymentMethonds), 0)
}

// NOTE: This method requires the payment method to already be saved to the customer
//
//	account, which must be submitted from the client. After the client has submitted
//	the payment method, the server can look it up, set it as the default method and
//	update the local database.
func TestSetDefaultPaymentMethod(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	customerConfig, err := service.GetCustomer(TEST_CUSTOMER_ID)
	assert.Nil(t, err)

	customerConfig, err = service.SetDefaultPaymentMethod(&SetDefaultPaymentMethodRequest{
		CustomerID:      customerConfig.ID,
		ProcessorID:     customerConfig.ProcessorID,
		PaymentMethodID: ""})
	assert.Nil(t, err)
	assert.NotEmpty(t, customerConfig.PaymentMethodID)
	assert.NotEmpty(t, customerConfig.PaymentMethodLast4)
}

// NOTE: This method requires the payment method to already be saved to the customer
//
//	account as the default payment method, After the client has a default payment
//	method set, the invoice can be paid.
func TestCreateInvoice(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	// Create new customer
	customerConfig, err := service.GetCustomer(TEST_CUSTOMER_ID)
	assert.Nil(t, err)

	// Get a list of products for the cart
	products := service.GetProducts()

	// Create a list of products for the invoice
	lineItems := make([]*Product, 0)
	lineItems = append(lineItems, &products[0])
	lineItems = append(lineItems, &products[1])

	// Set product quantities
	lineItems[0].Quantity = 1
	lineItems[1].Quantity = 2

	// Create a new invoice for the customer
	userAccount := &model.UserStruct{
		ID:    customerConfig.ID,
		Email: customerConfig.Email}
	createInvoiceRequest := CreateInvoiceRequest{
		Description: customerConfig.Description,
		Products:    lineItems}
	paymentIntentResponse, err := service.CreateInvoice(userAccount, createInvoiceRequest)
	assert.Nil(t, err)
	assert.NotNil(t, paymentIntentResponse)
	assert.NotEmpty(t, paymentIntentResponse.Customer.ID)
	assert.NotEmpty(t, paymentIntentResponse.PaymentIntent)
	assert.NotEmpty(t, paymentIntentResponse.ClientSecret)
	assert.NotEmpty(t, paymentIntentResponse.EphemeralKey)
	assert.NotEmpty(t, paymentIntentResponse.PublishableKey)

	// Make sure objects are serializable
	jsonData, err := json.Marshal(lineItems)
	assert.Nil(t, err)
	assert.NotNil(t, jsonData)

	var jsonProducts []Product
	err = json.Unmarshal(jsonData, &jsonProducts)
	assert.Nil(t, err)
	assert.NotNil(t, jsonProducts)
	assert.Greater(t, len(jsonProducts), 0)

	// Set the default payment method
	customerWithDefaultPaymentMethod, err := service.SetDefaultPaymentMethod(&SetDefaultPaymentMethodRequest{
		CustomerID:      paymentIntentResponse.Customer.ID,
		ProcessorID:     paymentIntentResponse.Customer.ProcessorID,
		PaymentMethodID: ""})
	assert.Nil(t, err)
	assert.NotEmpty(t, customerWithDefaultPaymentMethod.PaymentMethodID)
	assert.NotEmpty(t, customerWithDefaultPaymentMethod.PaymentMethodLast4)

	// Pay the invoice
	invoice, err := service.PayInvoice(paymentIntentResponse.InvoiceID)
	assert.Nil(t, err)
	assert.NotNil(t, invoice)
	assert.NotEmpty(t, invoice.ID)
}

func TestGetTaxRates(t *testing.T) {

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	taxRates := service.GetTaxRates()
	assert.NotEmpty(t, taxRates, "tax rates not found")
	assert.Greater(t, len(taxRates), 0)
}

func TestCreateSubscription(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	// Create new customer
	customerConfig, err := service.GetCustomer(TEST_CUSTOMER_ID)
	assert.Nil(t, err)

	subscriptionConfig := &Subscription{
		Customer: customerConfig,
		PriceID:  "price_1P9BsMDsf61j8i6epI3RjXO7"}
	subscription, err := service.CreateSubscription(subscriptionConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, subscription)
}

func TestUpdateSubscription(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	customerConfig, err := service.GetCustomer(TEST_CUSTOMER_ID)
	assert.Nil(t, err)

	subscriptionConfig := &Subscription{
		Customer: customerConfig,
		PriceID:  "price_1P9BsMDsf61j8i6epI3RjXO7"}
	savedSubscription, err := service.CreateSubscription(subscriptionConfig)
	assert.Nil(t, err)
	assert.NotNil(t, savedSubscription)
	assert.NotEmpty(t, savedSubscription.LatestInvoiceID)

	// You cannot update a subscription in `incomplete` status. Pay the invoice
	// so the subscription can be updated
	paidInvoice, err := service.PayInvoice(savedSubscription.LatestInvoiceID)
	assert.Nil(t, err)
	assert.NotEmpty(t, paidInvoice)

	// Simulate the webhook being called to set the default payment method
	// for this subscription.
	newSubscriptionConfig := &Subscription{
		ID:              savedSubscription.ID,
		Customer:        customerConfig,
		PaymentIntentID: paidInvoice.PaymentIntentID,
		PriceID:         "price_1P9BsMDsf61j8i6epI3RjXO7"}
	updatedSubscription, err := service.UpdateSubscription(newSubscriptionConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, updatedSubscription)
	assert.Equal(t, savedSubscription.LatestInvoiceID, updatedSubscription.LatestInvoiceID)

	// Update the subscription to a new price/product
	newPriceSubscriptionConfig := &Subscription{
		ID:              savedSubscription.ID,
		Customer:        customerConfig,
		PaymentIntentID: paidInvoice.PaymentIntentID,
		PriceID:         "price_1P9BfkDsf61j8i6exUaKxe3C"}
	updatedSubscriptionWithNewPriceID, err := service.UpdateSubscription(newPriceSubscriptionConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, updatedSubscriptionWithNewPriceID)
	assert.Equal(t, updatedSubscriptionWithNewPriceID.LatestInvoiceID, updatedSubscription.LatestInvoiceID)
	assert.NotEqual(t, updatedSubscriptionWithNewPriceID.PriceID, updatedSubscription.PriceID)
	assert.Equal(t, updatedSubscriptionWithNewPriceID.PriceID, newPriceSubscriptionConfig.PriceID)
}

func TestInvoicePaymentWebhook(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	// Create new customer
	customerConfig, err := service.GetCustomer(TEST_CUSTOMER_ID)
	assert.Nil(t, err)

	subscription := &Subscription{
		Customer: customerConfig,
		PriceID:  "price_1P9BsMDsf61j8i6epI3RjXO7"}
	savedSubscription, err := service.CreateSubscription(subscription)
	assert.Nil(t, err)
	assert.NotNil(t, savedSubscription)
}

// func TestUpdateSubscription(t *testing.T) {

// 	customerDAO := createCustomerDAO()
// 	service := NewStripeService(CurrentTest.app, customerDAO)
// 	testCustomer := createTestCustomer()

// 	customerConfig, err := service.CreateCustomer(testCustomer)
// 	assert.Nil(t, err)
// 	assert.Equal(t, customerConfig.Name, testCustomer.Name)
// 	assert.Equal(t, customerConfig.Email, testCustomer.Email)
// 	assert.NotEmpty(t, customerConfig.ProcessorID)

// 	paymentMethod, err := service.AttachPaymentMethond(testCustomer.ProcessorID, "pm_card_visa")
// 	assert.Nil(t, err)
// 	assert.Equal(t, "pm_card_visa", paymentMethod.ID)

// 	subscription, err := service.CreateSubscription(customerConfig, "price_1P9BfkDsf61j8i6exUaKxe3C")
// 	assert.Nil(t, err)
// 	assert.NotNil(t, subscription)

// 	customer, err := service.GetCustomer(customerConfig.ID)

// 	assert.Nil(t, err)
// 	assert.Equal(t, customer.ID, customerConfig.ID)
// 	//assert.Equal(t, "pm_card_visa", "")

// 	updatedSubscription, err := service.UpdateSubscription(subscription, "price_1P9BsMDsf61j8i6epI3RjXO7")
// 	assert.Nil(t, err)
// 	assert.NotNil(t, updatedSubscription)
// 	//assert.Equal(t, updatedSubscription.Items[0].Price.ID, "price_1P9BsMDsf61j8i6epI3RjXO7")
// }

func TestCancelSubscription(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	// Create new customer
	customerConfig, err := service.GetCustomer(TEST_CUSTOMER_ID)
	assert.Nil(t, err)

	subscription := &Subscription{
		Customer: customerConfig,
		PriceID:  "price_1P9BsMDsf61j8i6epI3RjXO7"}
	savedSubscription, err := service.CreateSubscription(subscription)
	assert.Nil(t, err)
	assert.NotNil(t, savedSubscription)

	_, err = service.CancelSubscription(savedSubscription.ID)
	assert.Nil(t, err)
}

func TestListSubscriptions(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	// Create new customer
	customerConfig, err := service.GetCustomer(TEST_CUSTOMER_ID)
	assert.Nil(t, err)

	subscription := &Subscription{
		Customer: customerConfig,
		PriceID:  "price_1P9BsMDsf61j8i6epI3RjXO7"}
	savedSubscription, err := service.CreateSubscription(subscription)
	assert.Nil(t, err)
	assert.NotNil(t, savedSubscription)

	subscriptions := service.ListSubscriptions(customerConfig.ProcessorID)
	assert.Nil(t, err)
	assert.NotNil(t, subscriptions)
	assert.Greater(t, len(subscriptions), 0)

	// for _, _subscription := range subscriptions {
	// 	_, err = service.CancelSubscription(_subscription.ID)
	// 	assert.Nil(t, err)
	// }
}

func TestInvoicePreview(t *testing.T) {

	SKIP_TEARDOWN_FLAG = true

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	// Create new customer
	customerConfig, err := service.GetCustomer(TEST_CUSTOMER_ID)
	assert.Nil(t, err)

	subscription := &Subscription{
		Customer: customerConfig,
		PriceID:  "price_1P9BsMDsf61j8i6epI3RjXO7"}
	savedSubscription, err := service.CreateSubscription(subscription)
	assert.Nil(t, err)
	assert.NotNil(t, savedSubscription)

	// You cannot update a subscription in `incomplete` status. Pay the invoice
	// so the subscription can be updated
	// Set the default payment method
	// success, err := service.PayInvoice(savedSubscription.LatestInvoiceID)
	// assert.Nil(t, err)
	// assert.True(t, success)

	invoiceToPreview, err := service.InvoicePreview(customerConfig.ProcessorID,
		savedSubscription.ID, subscription.PriceID)
	assert.Nil(t, err)
	assert.NotNil(t, invoiceToPreview)
}

func createCustomerDAO() dao.CustomerDAO {
	return gorm.NewCustomerDAO(CurrentTest.logger, CurrentTest.gorm)
}

func createTestCustomer() *config.CustomerStruct {
	address := &config.AddressStruct{
		Line1:      "123 test street",
		Line2:      "Apt 1",
		City:       "TestCity",
		State:      "TestState",
		PostalCode: "12345",
		Country:    "TestCountry"}

	shipping := &config.ShippingAddressStruct{
		Name:  "Test Receiver",
		Phone: "9876543210",
		Address: &config.AddressStruct{
			Line1:      "456 receiver ave",
			Line2:      "Apt 2",
			City:       "ShipCity",
			State:      "ShipState",
			PostalCode: "54321",
			Country:    "ShipCountry"}}

	return &config.CustomerStruct{
		ID:          1,
		Description: "Integration test user",
		Name:        "test",
		Email:       "root@test.com",
		Phone:       "11234567890",
		Address:     address,
		Shipping:    shipping,
	}

}
