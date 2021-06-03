// +build InAppPurchase

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/service"
)

type InAppPurchaseRestService interface {
	Verify(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultInAppPurchaseRestService struct {
	iapService service.InAppPurchaseService
	middleware service.Middleware
	jsonWriter common.HttpWriter
	InAppPurchaseRestService
}

func NewInAppPurchaseRestService(iapService service.InAppPurchaseService, middleware service.Middleware,
	jsonWriter common.HttpWriter) InAppPurchaseRestService {

	return &DefaultInAppPurchaseRestService{
		iapService: iapService,
		middleware: middleware,
		jsonWriter: jsonWriter}
}

func (restService *DefaultInAppPurchaseRestService) RegisterEndpoints(router *mux.Router, baseURI string) []string {
	iapVerifyEndpoint := fmt.Sprintf("%s/iap/verify", baseURI)
	router.Handle(iapVerifyEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.Verify)),
	)).Methods("POST")
	return []string{iapVerifyEndpoint}
}

func (restService *DefaultInAppPurchaseRestService) Verify(w http.ResponseWriter, r *http.Request) {

	scope, err := restService.middleware.CreateScope(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer scope.Close()

	var iap model.InAppPurchase
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&iap); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	verified, err := restService.iapService.Verify(scope, &iap)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	scope.GetLogger().Debugf("[InAppPurchaseRestService.Verify] verified=%t", verified)

	restService.jsonWriter.Write(w, http.StatusOK, verified)
}
