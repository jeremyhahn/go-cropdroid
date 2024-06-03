package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/builder"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
	"github.com/jeremyhahn/go-cropdroid/webservice"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

var App *app.App
var WebServerV1 *webservice.WebServerV1
var ServiceRegistry service.ServiceRegistry
var RestServiceRegistry rest.RestServiceRegistry
var JWT string

func TestMain(m *testing.M) {
	setup()
	login()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func teardown() {
	App.ShutdownChan <- true
	os.Remove(fmt.Sprintf("%s/%s.db",
		App.GORMInitParams.DataDir, App.Name))
}

func setup() {

	App = app.NewApp()
	App.Init(&app.AppInitParams{
		Debug:     true,
		LogDir:    "/var/log",
		ConfigDir: "."})
	App.DataStoreEngine = "sqlite"

	serviceMapper, serviceRegistry, restServiceRegistry,
		farmTickerProvisionerChan, err := builder.NewGormConfigBuilder(App).Build()
	if err != nil {
		App.Logger.Fatal(err)
	}

	farmServices := serviceRegistry.GetFarmServices()
	for _, farmService := range farmServices {
		go farmService.Run()
	}

	WebServerV1 = webservice.NewWebServerV1(
		App, serviceMapper,
		serviceRegistry,
		restServiceRegistry,
		farmTickerProvisionerChan)
	go WebServerV1.Run()
	go WebServerV1.RunProvisionerConsumer()

	serviceRegistry.GetEventLogService(0).Create(0, common.CONTROLLER_TYPE_SERVER, "System", "Integration Test Startup")

	// Don't block on integration tests

	// <-App.ShutdownChan
	// close(App.ShutdownChan)

	// serviceRegistry.GetEventLogService(0).Create(0, common.CONTROLLER_TYPE_SERVER, "System", "Shutdown")

	// Set up test variables
	ServiceRegistry = serviceRegistry
	RestServiceRegistry = restServiceRegistry
}

// Obtain a JWT using the default username and password
func login() {
	if JWT == "" {
		userCredentials := service.UserCredentials{
			Email:    common.DEFAULT_USER,
			Password: common.DEFAULT_PASSWORD,
		}
		jsonCreds, err := json.Marshal(userCredentials)
		if err != nil {
			App.Logger.Fatal(err)
		}
		req, err := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonCreds))
		if err != nil {
			App.Logger.Fatal(err)
		}
		rr := httptest.NewRecorder()

		// Set the handler function an serve the request
		handler := http.HandlerFunc(RestServiceRegistry.JsonWebTokenService().GenerateToken)

		// Send the request and record the response
		handler.ServeHTTP(rr, req)

		var jwt viewmodel.JsonWebToken
		err = json.Unmarshal(rr.Body.Bytes(), &jwt)
		if err != nil {
			App.Logger.Error(err)
		}
		JWT = jwt.Value
	}
}

func createFakeSession() service.Session {
	return service.CreateSession(
		App.Logger,
		[]service.OrganizationClaim{},
		[]service.FarmClaim{},
		nil,
		0,
		0,
		common.CONSISTENCY_LOCAL,
		model.NewUser())
}
