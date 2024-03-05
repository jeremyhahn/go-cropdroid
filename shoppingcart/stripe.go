package shoppingcart

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
)

type ShoppingCartService interface {
	GetProducts() []Product
}

type StripeClient struct {
	app    *app.App
	params *config.Stripe
	ShoppingCartService
}

func NewStripeService(app *app.App) ShoppingCartService {
	stripe.Key = app.Stripe.Key
	return &StripeClient{
		app:    app,
		params: app.Stripe}
}

func (client *StripeClient) GetProducts() []Product {
	products := make([]Product, 0)
	params := &stripe.ProductListParams{}
	params.Limit = stripe.Int64(10)
	result := product.List(params)
	for result.Iter.Next() {
		item := result.Iter.Current()
		priceParams := &stripe.PriceListParams{}
		priceParams.Product = stripe.String(item.(*stripe.Product).ID)
		priceParams.Limit = stripe.Int64(10)
		result := price.List(priceParams)
		for result.Iter.Next() {
			price := result.Iter.Current()
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
