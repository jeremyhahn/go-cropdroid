package shoppingcart

import (
	"encoding/json"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
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

	assert.Equal(t, len(products), 5)
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

func TestCreateCustomer(t *testing.T) {

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)
	customer := createTestCustomer()

	stripeCustomer, err := service.CreateCustomer(customer)
	assert.Nil(t, err)
	assert.Equal(t, stripeCustomer.Name, customer.Name)
	assert.Equal(t, stripeCustomer.Email, customer.Email)
	assert.NotEmpty(t, stripeCustomer.ProcessorID)

	customers, err := customerDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(customers))
	assert.Equal(t, customer.Name, customers[0].Name)
	assert.Equal(t, customer.Email, customers[0].Email)
	assert.Equal(t, customer.ProcessorID, stripeCustomer.ProcessorID)
	assert.NotEmpty(t, customer.ProcessorID, customers[0].ProcessorID)

	assert.Equal(t, customer.Address, customers[0].Address)
	assert.Equal(t, customer.Shipping, customers[0].Shipping)
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

	customers, err := customerDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(customers))
	assert.Equal(t, newCustomerName, customers[0].Name)
	assert.Equal(t, customer.Email, customers[0].Email)
	assert.NotEmpty(t, customerConfig.ProcessorID, customers[0].ProcessorID)
	assert.NotEmpty(t, updatedCustomerConfig.ProcessorID, customers[0].ProcessorID)

	assert.Equal(t, customer.Address, customers[0].Address)
	assert.Equal(t, customer.Shipping, customers[0].Shipping)
}

func TestCreateInvoice(t *testing.T) {

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)
	testCustomer := createTestCustomer()

	// Create new customer
	customer, err := service.CreateCustomer(testCustomer)
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
	userAccount := &model.User{
		ID:    customer.ID,
		Email: customer.Email}
	createInvoiceRequest := CreateInvoiceRequest{
		Description: customer.Description,
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

	// // Pay the invoice
	//
	// TODO: Fix the error "There is no `default_payment_method` set on this Customer or Invoice."
	//
	// success, err := service.PayInvoice(paymentIntentResponse.InvoiceID)
	// assert.Nil(t, err)
	// assert.True(t, success)
}

func TestGetPaymentMethods(t *testing.T) {

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)
	customer := createTestCustomer()

	stripeCustomer, err := service.CreateCustomer(customer)
	assert.Nil(t, err)
	assert.Equal(t, stripeCustomer.Name, customer.Name)
	assert.Equal(t, stripeCustomer.Email, customer.Email)
	assert.NotEmpty(t, stripeCustomer.ProcessorID)

	paymentMethods := service.GetPaymentMethods(stripeCustomer.ProcessorID)

	assert.NotEmpty(t, paymentMethods)
}

func TestGetTaxRates(t *testing.T) {

	customerDAO := createCustomerDAO()
	service := NewStripeService(CurrentTest.app, customerDAO)

	taxRates := service.GetTaxRates()
	assert.NotEmpty(t, taxRates, "tax rates not found")
	assert.Greater(t, len(taxRates), 0)
}

func createCustomerDAO() dao.CustomerDAO {
	return gorm.NewCustomerDAO(CurrentTest.logger, CurrentTest.gorm)
}

func createTestCustomer() *config.Customer {
	address := &config.Address{
		Line1:      "123 test street",
		Line2:      "Apt 1",
		City:       "TestCity",
		State:      "TestState",
		PostalCode: "12345",
		Country:    "TestCountry"}

	shipping := &config.ShippingAddress{
		Name:  "Test Receiver",
		Phone: "9876543210",
		Address: &config.Address{
			Line1:      "456 receiver ave",
			Line2:      "Apt 2",
			City:       "ShipCity",
			State:      "ShipState",
			PostalCode: "54321",
			Country:    "ShipCountry"}}

	return &config.Customer{
		ID:          1,
		Description: "Integration test user",
		Name:        "test",
		Email:       "root@test.com",
		Phone:       "11234567890",
		Address:     address,
		Shipping:    shipping,
	}

}
