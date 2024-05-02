package shoppingcart

import "github.com/jeremyhahn/go-cropdroid/config"

type CreateInvoiceRequest struct {
	Description string     `json:"description"`
	Products    []*Product `json:"products"`
}

type PaymentIntentResponse struct {
	Customer       *config.Customer `json:"customer"`
	InvoiceID      string           `json:"invoiceId"`
	PaymentIntent  string           `json:"paymentIntent"`
	ClientSecret   string           `json:"clientSecret"`
	EphemeralKey   string           `json:"ephemeralKey"`
	PublishableKey string           `json:"publishableKey"`
}

type CreatePaymentIntentRequest struct {
	CustomerID   string `json:"customerId"`
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
}

type SetupIntentRequest struct {
	ClientSecret string `json:"client_secret"`
	//ProcessorID string `json:"processor_id"`
}

type SetupIntentResponse struct {
	Customer       *config.Customer `json:"customer"`
	EphemeralKey   string           `json:"ephemeral_key"`
	ClientSecret   string           `json:"client_secret"`
	PublishableKey string           `json:"publishable_key"`
}

type CustomerEphemeralKeyResponse struct {
	Customer     *config.Customer `json:"customer"`
	EphemeralKey string           `json:"ephemeral_key"`
}

type AttachPaymentMethodRequest struct {
	ProcessorID     string `json:"processor_id"`
	PaymentMethodID string `json:"payment_method_id"`
}

type SetDefaultPaymentMethodRequest struct {
	CustomerID      uint64 `json:"customer_id"`
	ProcessorID     string `json:"processor_id"`
	PaymentMethodID string `json:"payment_method_id"`
}

type Invoice struct {
	ID              string     `json:"id"`
	Description     string     `json:"description"`
	AutoAdvance     bool       `json:"auto_advance"`
	PaymentIntentID string     `json:"payment_intent"`
	DefaultTaxRates []*TaxRate `json:"default_tax_rates"`
}

type Cart struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
}

type Product struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	ImageUrl    string            `json:"imageUrl"`
	Price       int64             `json:"price"`
	Quantity    int64             `json:"quantity"`
	Metadata    map[string]string `json:"metadata"`
}

type Subscription struct {
	ID              string     `json:"id"`
	Description     string     `json:"description"`
	AutomaticTax    bool       `json:"automatic_tax"`
	DefaultTaxRates []*TaxRate `json:"default_tax_rates"`
	PriceID         string     `json:"price_id"`
	LatestInvoiceID string     `json:"latest_invoice"`
	//DefaultPaymentMethodID string           `json:"default_payment_method"`
	PaymentIntentID string           `json:"payment_intent"`
	Customer        *config.Customer `json:"customer"`
}

type TaxRate struct {
	ID           string  `json:"id"`
	DisplayName  string  `json:"display_name"`
	Description  string  `json:"description"`
	State        string  `json:"state"`
	Country      string  `json:"country"`
	Inclusive    bool    `json:"inclusive"`
	Jurisdiction string  `json:"jurisdiction"`
	Percentage   float64 `json:"percentage"`
}

type CreditCard struct {
	ID       string `json:"id"`
	Number   string `json:"number"`
	ExpMonth int64  `json:"exp_month"`
	ExpYear  int64  `json:"exp_year"`
	CVC      string `json:"cvc"`
}
