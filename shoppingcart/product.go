package shoppingcart

type Product struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageUrl    string `json:"imageUrl"`
	Price       int64  `json:"price"`
	Quantity    int64  `json:"quantity"`
}
