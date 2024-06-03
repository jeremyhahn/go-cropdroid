package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/stretchr/testify/assert"

	restmocks "github.com/jeremyhahn/go-cropdroid/test/mocks/restservice"
	servicemocks "github.com/jeremyhahn/go-cropdroid/test/mocks/service"
)

// Send a request to the real device channel's condition list
// endpoint to ensure the router is sending requests to the proper
// REST and business logic service methods.
func TestConditionListViewEndpoint(t *testing.T) {

	// mocked / expected response
	models := []*viewmodel.Condition{
		{
			ID:         uint64(1),
			DeviceType: "Test Device",
			MetricID:   uint64(1),
			MetricName: "Test Metric",
			ChannelID:  uint64(1),
			Comparator: ">",
			Threshold:  5,
			Text:       "Test metric is greater than 5",
		},
	}

	// create mock REST and business logic services with stubs used by the conditions endpoint
	conditionRestService := RestServiceRegistry.ConditionRestService()
	mockConditionService := new(servicemocks.MockConditionService)
	// create mock JWT middleware service with stub for CreateSession
	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
	mockJsonWebTokenService.SetApp(App)

	// Replace the real condition service with the mocked service
	conditionRestService.SetService(mockConditionService)
	conditionRestService.SetMiddleware(mockJsonWebTokenService)

	// Build the request for the conditions endpoint with a new httptest.Recorder
	req, err := http.NewRequest("GET", "/api/v1/farms/1/conditions/1", nil)
	req.Header.Set("Authorization", JWT)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// set mock mux parameters
	vars := map[string]string{
		"farmID":    "1",
		"channelID": "1",
	}
	req = mux.SetURLVars(req, vars)

	// create a fake session
	fakeSession := createFakeSession()

	// mock the service call, passing a fake session and channel ID to match the channelID used in the GET request
	mockServiceCall := mockConditionService.On("ListView", fakeSession, uint64(1)).Return(models, nil)

	// mock the JWT CreateSession call and return the fake session
	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

	// set the HTTP handler
	handler := http.HandlerFunc(conditionRestService.ListView)

	// send the request and record the response
	handler.ServeHTTP(rr, req)

	// assert that the service expectations were met
	mockConditionService.AssertExpectations(t)

	// remove the handler so another one can be added to take precedence
	mockServiceCall.Unset()

	// build the exepected web service response
	webServiceResponse := &response.WebServiceResponse{
		Code:    200,
		Success: true,
		Payload: models}
	jsonWebServiceResponse, err := json.Marshal(webServiceResponse)
	assert.Nil(t, err)

	// assert the expected response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, string(jsonWebServiceResponse), rr.Body.String())
}

// Send a request to the real device condition create endpoint to
// ensure the router is sending requests to the proper REST and
// business logic service methods.
func TestConditionCreateEndpoint(t *testing.T) {

	// Fake condition Update
	fakeCondition := &config.ConditionStruct{
		ID:         uint64(1),
		WorkflowID: uint64(1),
		MetricID:   uint64(1),
		ChannelID:  uint64(1),
		Comparator: ">",
		Threshold:  5,
	}

	// create mock REST and business logic services that stub methods used by the conditions endpoint
	conditionRestService := RestServiceRegistry.ConditionRestService()
	mockConditionService := new(servicemocks.MockConditionService)
	// create mock JWT middleware service to stub CreateSession
	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
	mockJsonWebTokenService.SetApp(App)

	// Replace the real condition service with the mocked service
	conditionRestService.SetService(mockConditionService)
	conditionRestService.SetMiddleware(mockJsonWebTokenService)

	jsonFakeCondition, err := json.Marshal(fakeCondition)
	assert.Nil(t, err)

	// Build the request for the GetPage endpoint with a new httptest.Recorder
	req, err := http.NewRequest("POST", "/api/v1/farms/1/conditions/1", strings.NewReader(string(jsonFakeCondition)))
	req.Header.Set("Authorization", JWT)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// set mock mux parameters
	vars := map[string]string{
		"farmID": "1",
	}
	req = mux.SetURLVars(req, vars)

	// create a fake session
	fakeSession := createFakeSession()

	// mock the service call, passing a fake session and device ID to match the deviceID used in the GET request
	mockServiceCall := mockConditionService.On("Create", fakeSession, fakeCondition).Return(nil)

	// mock the JWT CreateSession call and return the fake session
	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

	// set the HTTP handler
	handler := http.HandlerFunc(conditionRestService.Create)

	// send the request and record the response
	handler.ServeHTTP(rr, req)

	// assert that the service expectations were met
	mockConditionService.AssertExpectations(t)

	// remove the handler so another one can be added to take precedence
	mockServiceCall.Unset()

	// build the exepected web service response
	webServiceResponse := &response.WebServiceResponse{
		Code:    200,
		Success: true,
		Payload: fakeCondition}
	jsonWebServiceResponse, err := json.Marshal(webServiceResponse)
	assert.Nil(t, err)

	// assert the expected response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, string(jsonWebServiceResponse), rr.Body.String())
}

// Send a request to the real device condition update endpoint to
// ensure the router is sending requests to the proper REST and
// business logic service methods.
func TestConditionUpdateEndpoint(t *testing.T) {

	// Fake Update
	fakeCondition := &config.ConditionStruct{
		ID:         uint64(1),
		WorkflowID: uint64(1),
		MetricID:   uint64(1),
		ChannelID:  uint64(1),
		Comparator: ">",
		Threshold:  5,
	}

	// create mock REST and business logic services that stub methods used by the conditions endpoint
	conditionRestService := RestServiceRegistry.ConditionRestService()
	mockConditionService := new(servicemocks.MockConditionService)
	// create mock JWT middleware service to stub CreateSession
	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
	mockJsonWebTokenService.SetApp(App)

	// Replace the real condition service with the mocked service
	conditionRestService.SetService(mockConditionService)
	conditionRestService.SetMiddleware(mockJsonWebTokenService)

	jsonFakeCondition, err := json.Marshal(fakeCondition)
	assert.Nil(t, err)

	// Build the request for the GetPage endpoint with a new httptest.Recorder
	req, err := http.NewRequest("PUT", "/api/v1/farms/1/conditions/1", strings.NewReader(string(jsonFakeCondition)))
	req.Header.Set("Authorization", JWT)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// set mock mux parameters
	vars := map[string]string{
		"farmID": "1",
	}
	req = mux.SetURLVars(req, vars)

	// create a fake session
	fakeSession := createFakeSession()

	// mock the service call, passing a fake session and device ID to match the deviceID used in the GET request
	mockServiceCall := mockConditionService.On("Update", fakeSession, fakeCondition).Return(nil)

	// mock the JWT CreateSession call and return the fake session
	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

	// set the HTTP handler
	handler := http.HandlerFunc(conditionRestService.Update)

	// send the request and record the response
	handler.ServeHTTP(rr, req)

	// assert that the service expectations were met
	mockConditionService.AssertExpectations(t)

	// remove the handler so another one can be added to take precedence
	mockServiceCall.Unset()

	// build the exepected web service response
	webServiceResponse := &response.WebServiceResponse{
		Code:    200,
		Success: true,
		Payload: nil}
	jsonWebServiceResponse, err := json.Marshal(webServiceResponse)
	assert.Nil(t, err)

	// assert the expected response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, string(jsonWebServiceResponse), rr.Body.String())
}

// Send a request to the real device condition delete endpoint to
// ensure the router is sending requests to the proper REST and
// business logic service methods.
func TestConditionDeleteEndpoint(t *testing.T) {

	// Expected condition passed to business logic service Delete method
	fakeCondition := &config.ConditionStruct{ID: uint64(1)}

	// create mock REST and business logic services that stub methods used by the conditions endpoint
	conditionRestService := RestServiceRegistry.ConditionRestService()
	mockConditionService := new(servicemocks.MockConditionService)
	// create mock JWT middleware service to stub CreateSession
	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
	mockJsonWebTokenService.SetApp(App)

	// Replace the real condition service with the mocked service
	conditionRestService.SetService(mockConditionService)
	conditionRestService.SetMiddleware(mockJsonWebTokenService)

	// Build the request for the GetPage endpoint with a new httptest.Recorder
	req, err := http.NewRequest("DELETE", "/api/v1/farms/1/conditions/1", nil)
	req.Header.Set("Authorization", JWT)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// set mock mux parameters
	vars := map[string]string{
		"farmID": "1",
		"id":     "1",
	}
	req = mux.SetURLVars(req, vars)

	// create a fake session
	fakeSession := createFakeSession()

	// mock the service call, passing a fake session and device ID to match the deviceID used in the GET request
	mockServiceCall := mockConditionService.On("Delete", fakeSession, fakeCondition).Return(nil)

	// mock the JWT CreateSession call and return the fake session
	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

	// set the HTTP handler
	handler := http.HandlerFunc(conditionRestService.Delete)

	// send the request and record the response
	handler.ServeHTTP(rr, req)

	// assert that the service expectations were met
	mockConditionService.AssertExpectations(t)

	// remove the handler so another one can be added to take precedence
	mockServiceCall.Unset()

	// build the exepected web service response
	webServiceResponse := &response.WebServiceResponse{
		Code:    200,
		Success: true,
		Payload: nil}
	jsonWebServiceResponse, err := json.Marshal(webServiceResponse)
	assert.Nil(t, err)

	// assert the expected response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, string(jsonWebServiceResponse), rr.Body.String())
}
