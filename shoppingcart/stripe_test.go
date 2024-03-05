package shoppingcart

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpGetType(t *testing.T) {

	stripeService := NewStripeService(CurrentTest.app)
	products := stripeService.GetProducts()

	jsonData, _ := json.Marshal(products)
	fmt.Println(string(jsonData))

	assert.Equal(t, len(products), 5)
}
