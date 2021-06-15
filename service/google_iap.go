// +build cloud

package service

// Google Play In-App Purchase Integration

// client:
// https://medium.com/@vleonovs8/tutorial-google-play-billing-in-app-purchases-6143bda8d290

// server:
// http://www.evanlin.com/server-side-iap-verification-google-play/
// https://stackoverflow.com/questions/43536904/google-play-developer-api-the-current-user-has-insufficient-permissions-to-pe
// //https://godoc.org/google.golang.org/api/androidpublisher/v3#ProductPurchase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	androidpublisher "google.golang.org/api/androidpublisher/v2"
	"google.golang.org/api/option"
)

type GoogleIAP struct {
	Kind               string `json:"kind"`
	PurchaseTimeMillis string `json:"purchaseTimeMillis"`
	PurchaseState      string `json:"purchaseState"`
	ConsumptionState   bool   `json:"consumptionState"`
	OrderId            string `json:"orderId"`
	DeveloperPayload   string `json:"developerPayload"`
}

type InAppPurchaseService interface {
	Verify(session Session, purchasedItem common.InAppPurchase) (bool, error)
}

type AndroidInAppPurchaseService struct {
	app              *app.App
	session          Session
	publisherService *androidpublisher.Service
	InAppPurchaseService
}

func NewAndroidInAppPurchaseService(app *app.App, session Session) (InAppPurchaseService, error) {
	oauthFile := fmt.Sprintf("%s/googleplay-oauth2.json", app.ConfigDir)
	if _, err := os.Stat(oauthFile); os.IsNotExist(err) {
		app.Logger.Fatalf("[AndroidInAppPurchaseService] Google Play OAuth JSON config not found: %s", oauthFile)
		return nil, err
	}
	context := context.Background()
	androidpublisherService, err := androidpublisher.NewService(context, option.WithCredentialsFile(oauthFile))
	if err != nil {
		return nil, err
	}
	return &AndroidInAppPurchaseService{
		app:              app,
		session:          session,
		publisherService: androidpublisherService}, nil
}

func (iap *AndroidInAppPurchaseService) Verify(session Session, purchasedItem common.InAppPurchase) (bool, error) {

	iap.app.Logger.Debugf("Verifying purchased item. user=%+v, product=%+v", session.GetUser().GetEmail(), purchasedItem)

	purchasesService := androidpublisher.NewPurchasesProductsService(iap.publisherService)
	productPurchase, err := purchasesService.Get(common.PACKAGE, purchasedItem.GetProductID(), purchasedItem.GetPurchaseToken()).Do()
	if err != nil {
		return false, err
	}

	if productPurchase.PurchaseTimeMillis != purchasedItem.GetPurchaseTimeMillis() {
		return false, fmt.Errorf("Purchase time mismatch")
	}

	if productPurchase.ConsumptionState != 1 {
		// 0 = Yet to be consumed
		// 1 = Consumed
		return false, fmt.Errorf("Product yet to be consumed")
	}

	if productPurchase.PurchaseState != 0 {
		// 0 = Purchased
		// 1 = Canceled
		// 2 = Pending
		state := "cancelled"
		if productPurchase.PurchaseState == 2 {
			state = "pending"
		}
		return false, fmt.Errorf("Purchase state %s", state)
	}

	iap.app.Logger.Debug("orderId: ", productPurchase.OrderId)
	iap.app.Logger.Debug("productPurchase: ", productPurchase)
	iap.app.Logger.Debug("purchaseTimeMillis: ", productPurchase.PurchaseTimeMillis)
	iap.app.Logger.Debug("consumptionState: ", productPurchase.ConsumptionState)
	iap.app.Logger.Debug("purchaseState: ", productPurchase.PurchaseState)

	//purchaseTime := time.Unix(productPurchase.PurchaseTimeMillis/1000, 0)
	//log.Println("time purchase (local): ", timePurchase.Local())

	purchasedItemJSON, err := json.Marshal(purchasedItem)
	if err != nil {
		return false, nil
	}

	// 	resp, err := iap.etcd.Put(ctx, iap.createEtcdKey(scope, productPurchase.OrderId), string(purchasedItemJSON))

	iap.app.Logger.Debugf("purchasedItemJSON: %+v", string(purchasedItemJSON))

	return true, nil
}

func (iap *AndroidInAppPurchaseService) createEtcdKey(session Session, orderID string) string {
	return fmt.Sprintf("iap_%d_%s", session.GetUser().GetID(), orderID)
}
