package shoppingcart

type CreatePaymentIntentRequest struct {
	CustomerID   string `json:"customerId"`
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
}
