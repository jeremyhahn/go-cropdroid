package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/stretchr/testify/assert"

	restmocks "github.com/jeremyhahn/go-cropdroid/test/mocks/restservice"
	servicemocks "github.com/jeremyhahn/go-cropdroid/test/mocks/service"
)

// View(w http.ResponseWriter, r *http.Request)
// State(w http.ResponseWriter, r *http.Request)
// Switch(w http.ResponseWriter, r *http.Request)
// Metric(w http.ResponseWriter, r *http.Request)
// TimerSwitch(w http.ResponseWriter, r *http.Request)
// History(w http.ResponseWriter, r *http.Request)

// Send a request to the real device channel's device list
// endpoint to ensure the router is sending requests to the proper
// REST and business logic service methods.
func TestDeviceViewEndpoint(t *testing.T) {

	// mocked / expected response
	farmID := uint64(1)

	// create mock REST and business logic services with stubs used by the devices endpoint
	deviceRestService := RestServiceRegistry.DeviceRestService()

	mockDeviceService := new(servicemocks.MockDeviceService)
	ServiceRegistry.SetDeviceService(farmID, mockDeviceService)

	// create mock JWT middleware service with stub for CreateSession
	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
	mockJsonWebTokenService.SetApp(App)

	// Replace the real device service with the mocked service
	deviceRestService.SetServiceRegistry(ServiceRegistry)
	deviceRestService.SetMiddleware(mockJsonWebTokenService)

	// Build the request for the devices endpoint with a new httptest.Recorder
	req, err := http.NewRequest("GET", "/api/v1/farms/1/devices/test-device/1", nil)
	req.Header.Set("Authorization", JWT)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// set mock mux parameters
	vars := map[string]string{
		"farmID":     "1",
		"deviceType": "test-device",
	}
	req = mux.SetURLVars(req, vars)

	// create a fake session
	fakeSession := createFakeSession()

	deviceView := mockDeviceService.ExpectedView()

	// mock the service call, passing a fake session and channel ID to match the channelID used in the GET request
	mockDeviceService.On("DeviceType").Return("test-device")
	mockServiceCall := mockDeviceService.On("View").Return(deviceView, nil)

	// mock the JWT CreateSession call and return the fake session
	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

	// set the HTTP handler
	handler := http.HandlerFunc(deviceRestService.View)

	// send the request and record the response
	handler.ServeHTTP(rr, req)

	// assert that the service expectations were met
	mockDeviceService.AssertExpectations(t)

	// remove the handler so another one can be added to take precedence
	mockServiceCall.Unset()

	// build the exepected web service response
	webServiceResponse := &response.WebServiceResponse{
		Code:    200,
		Success: true,
		Payload: deviceView}
	jsonWebServiceResponse, err := json.Marshal(webServiceResponse)
	assert.Nil(t, err)

	// assert the expected response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, string(jsonWebServiceResponse), rr.Body.String())
}

// // Send a request to the real device device create endpoint to
// // ensure the router is sending requests to the proper REST and
// // business logic service methods.
// func TestDeviceCreateEndpoint(t *testing.T) {

// 	// Fake device Update
// 	fakeDevice := &config.DeviceStruct{
// 		ID:         uint64(1),
// 		WorkflowID: uint64(1),
// 		MetricID:   uint64(1),
// 		ChannelID:  uint64(1),
// 		Comparator: ">",
// 		Threshold:  5,
// 	}

// 	// create mock REST and business logic services that stub methods used by the devices endpoint
// 	deviceRestService := RestServiceRegistry.DeviceRestService()
// 	mockDeviceService := new(servicemocks.MockDeviceService)
// 	// create mock JWT middleware service to stub CreateSession
// 	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
// 	mockJsonWebTokenService.SetApp(App)

// 	// Replace the real device service with the mocked service
// 	deviceRestService.SetService(mockDeviceService)
// 	deviceRestService.SetMiddleware(mockJsonWebTokenService)

// 	jsonFakeDevice, err := json.Marshal(fakeDevice)
// 	assert.Nil(t, err)

// 	// Build the request for the GetPage endpoint with a new httptest.Recorder
// 	req, err := http.NewRequest("POST", "/api/v1/farms/1/devices/1", strings.NewReader(string(jsonFakeDevice)))
// 	req.Header.Set("Authorization", JWT)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rr := httptest.NewRecorder()

// 	// set mock mux parameters
// 	vars := map[string]string{
// 		"farmID": "1",
// 	}
// 	req = mux.SetURLVars(req, vars)

// 	// create a fake session
// 	fakeSession := createFakeSession()

// 	// mock the service call, passing a fake session and device ID to match the deviceID used in the GET request
// 	mockServiceCall := mockDeviceService.On("Create", fakeSession, fakeDevice).Return(nil)

// 	// mock the JWT CreateSession call and return the fake session
// 	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

// 	// set the HTTP handler
// 	handler := http.HandlerFunc(deviceRestService.Create)

// 	// send the request and record the response
// 	handler.ServeHTTP(rr, req)

// 	// assert that the service expectations were met
// 	mockDeviceService.AssertExpectations(t)

// 	// remove the handler so another one can be added to take precedence
// 	mockServiceCall.Unset()

// 	// build the exepected web service response
// 	webServiceResponse := &response.WebServiceResponse{
// 		Code:    200,
// 		Success: true,
// 		Payload: fakeDevice}
// 	jsonWebServiceResponse, err := json.Marshal(webServiceResponse)
// 	assert.Nil(t, err)

// 	// assert the expected response
// 	assert.Equal(t, http.StatusOK, rr.Code)
// 	assert.Equal(t, string(jsonWebServiceResponse), rr.Body.String())
// }

// // Send a request to the real device device update endpoint to
// // ensure the router is sending requests to the proper REST and
// // business logic service methods.
// func TestDeviceUpdateEndpoint(t *testing.T) {

// 	// Fake Update
// 	fakeDevice := &config.DeviceStruct{
// 		ID:         uint64(1),
// 		WorkflowID: uint64(1),
// 		MetricID:   uint64(1),
// 		ChannelID:  uint64(1),
// 		Comparator: ">",
// 		Threshold:  5,
// 	}

// 	// create mock REST and business logic services that stub methods used by the devices endpoint
// 	deviceRestService := RestServiceRegistry.DeviceRestService()
// 	mockDeviceService := new(servicemocks.MockDeviceService)
// 	// create mock JWT middleware service to stub CreateSession
// 	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
// 	mockJsonWebTokenService.SetApp(App)

// 	// Replace the real device service with the mocked service
// 	deviceRestService.SetService(mockDeviceService)
// 	deviceRestService.SetMiddleware(mockJsonWebTokenService)

// 	jsonFakeDevice, err := json.Marshal(fakeDevice)
// 	assert.Nil(t, err)

// 	// Build the request for the GetPage endpoint with a new httptest.Recorder
// 	req, err := http.NewRequest("PUT", "/api/v1/farms/1/devices/1", strings.NewReader(string(jsonFakeDevice)))
// 	req.Header.Set("Authorization", JWT)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rr := httptest.NewRecorder()

// 	// set mock mux parameters
// 	vars := map[string]string{
// 		"farmID": "1",
// 	}
// 	req = mux.SetURLVars(req, vars)

// 	// create a fake session
// 	fakeSession := createFakeSession()

// 	// mock the service call, passing a fake session and device ID to match the deviceID used in the GET request
// 	mockServiceCall := mockDeviceService.On("Update", fakeSession, fakeDevice).Return(nil)

// 	// mock the JWT CreateSession call and return the fake session
// 	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

// 	// set the HTTP handler
// 	handler := http.HandlerFunc(deviceRestService.Update)

// 	// send the request and record the response
// 	handler.ServeHTTP(rr, req)

// 	// assert that the service expectations were met
// 	mockDeviceService.AssertExpectations(t)

// 	// remove the handler so another one can be added to take precedence
// 	mockServiceCall.Unset()

// 	// build the exepected web service response
// 	webServiceResponse := &response.WebServiceResponse{
// 		Code:    200,
// 		Success: true,
// 		Payload: nil}
// 	jsonWebServiceResponse, err := json.Marshal(webServiceResponse)
// 	assert.Nil(t, err)

// 	// assert the expected response
// 	assert.Equal(t, http.StatusOK, rr.Code)
// 	assert.Equal(t, string(jsonWebServiceResponse), rr.Body.String())
// }

// // Send a request to the real device device delete endpoint to
// // ensure the router is sending requests to the proper REST and
// // business logic service methods.
// func TestDeviceDeleteEndpoint(t *testing.T) {

// 	// Expected device passed to business logic service Delete method
// 	fakeDevice := &config.DeviceStruct{ID: uint64(1)}

// 	// create mock REST and business logic services that stub methods used by the devices endpoint
// 	deviceRestService := RestServiceRegistry.DeviceRestService()
// 	mockDeviceService := new(servicemocks.MockDeviceService)
// 	// create mock JWT middleware service to stub CreateSession
// 	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
// 	mockJsonWebTokenService.SetApp(App)

// 	// Replace the real device service with the mocked service
// 	deviceRestService.SetService(mockDeviceService)
// 	deviceRestService.SetMiddleware(mockJsonWebTokenService)

// 	// Build the request for the GetPage endpoint with a new httptest.Recorder
// 	req, err := http.NewRequest("DELETE", "/api/v1/farms/1/devices/1", nil)
// 	req.Header.Set("Authorization", JWT)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rr := httptest.NewRecorder()

// 	// set mock mux parameters
// 	vars := map[string]string{
// 		"farmID": "1",
// 		"id":     "1",
// 	}
// 	req = mux.SetURLVars(req, vars)

// 	// create a fake session
// 	fakeSession := createFakeSession()

// 	// mock the service call, passing a fake session and device ID to match the deviceID used in the GET request
// 	mockServiceCall := mockDeviceService.On("Delete", fakeSession, fakeDevice).Return(nil)

// 	// mock the JWT CreateSession call and return the fake session
// 	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

// 	// set the HTTP handler
// 	handler := http.HandlerFunc(deviceRestService.Delete)

// 	// send the request and record the response
// 	handler.ServeHTTP(rr, req)

// 	// assert that the service expectations were met
// 	mockDeviceService.AssertExpectations(t)

// 	// remove the handler so another one can be added to take precedence
// 	mockServiceCall.Unset()

// 	// build the exepected web service response
// 	webServiceResponse := &response.WebServiceResponse{
// 		Code:    200,
// 		Success: true,
// 		Payload: nil}
// 	jsonWebServiceResponse, err := json.Marshal(webServiceResponse)
// 	assert.Nil(t, err)

// 	// assert the expected response
// 	assert.Equal(t, http.StatusOK, rr.Code)
// 	assert.Equal(t, string(jsonWebServiceResponse), rr.Body.String())
// }
