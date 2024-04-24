package shoppingcart

type Cart struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
}
