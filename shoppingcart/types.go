package shoppingcart

type CreateInvoiceRequest struct {
	Description string     `json:"description"`
	Products    []*Product `json:"products"`
}
