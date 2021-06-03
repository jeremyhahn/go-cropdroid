package model

import "github.com/jeremyhahn/cropdroid/common"

type InAppPurchase struct {
	//OrderID              string `json:"orderId"`
	ProductID            string `json:"productId"`
	PurchaseToken        string `json:"purchaseToken"`
	PurchaseTimeMillis   int64  `json:"purchaseTime"`
	common.InAppPurchase `json:"-"`
}

func NewInAppPurchase(productID, token string, time int64) common.InAppPurchase {
	return &InAppPurchase{
		//	OrderID:            orderID,
		ProductID:          productID,
		PurchaseToken:      token,
		PurchaseTimeMillis: time}
}

/*
func (iap *InAppPurchase) GetOrderID() string {
	return iap.OrderID
}

func (iap *InAppPurchase) SetOrderID(id string) {
	iap.OrderID = id
}*/

func (iap *InAppPurchase) GetProductID() string {
	return iap.ProductID
}

func (iap *InAppPurchase) GetPurchaseToken() string {
	return iap.PurchaseToken
}

func (iap *InAppPurchase) GetPurchaseTimeMillis() int64 {
	return iap.PurchaseTimeMillis
}
