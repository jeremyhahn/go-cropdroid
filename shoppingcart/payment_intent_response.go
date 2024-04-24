package shoppingcart

import "github.com/jeremyhahn/go-cropdroid/config"

type PaymentIntentResponse struct {
	Customer       *config.Customer `json:"customer"`
	InvoiceID      string           `json:"invoiceId"`
	PaymentIntent  string           `json:"paymentIntent"`
	ClientSecret   string           `json:"clientSecret"`
	EphemeralKey   string           `json:"ephemeralKey"`
	PublishableKey string           `json:"publishableKey"`
}
