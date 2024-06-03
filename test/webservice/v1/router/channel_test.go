package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/stretchr/testify/assert"

	restmocks "github.com/jeremyhahn/go-cropdroid/test/mocks/restservice"
	servicemocks "github.com/jeremyhahn/go-cropdroid/test/mocks/service"
)

// Send a request to the real device channel list endpoint to
// ensure the router is sending requests to the proper REST and
// business logic service methods.
func TestChannelListEndpoint(t *testing.T) {

	// Fake response for channel service GetByDeviceID
	pageQuery := query.NewPageQuery()
	models := make([]model.Channel, pageQuery.PageSize)
	for i := 0; i < pageQuery.PageSize; i++ {
		models[i] = &model.ChannelStruct{
			ID:   uint64(i),
			Name: fmt.Sprintf("Test Channel %d", i)}
	}

	// create mock REST and business logic services that stub methods used by the channels endpoint
	channelRestService := RestServiceRegistry.ChannelRestService()
	mockChannelService := new(servicemocks.MockChannelService)
	// create mock JWT middleware service to stub CreateSession
	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
	mockJsonWebTokenService.SetApp(App)

	// Replace the real channel service with the mocked service
	channelRestService.SetService(mockChannelService)
	channelRestService.SetMiddleware(mockJsonWebTokenService)

	// Build the request for the GetPage endpoint with a new httptest.Recorder
	req, err := http.NewRequest("GET", "/api/v1/farms/1/channels/1", nil)
	req.Header.Set("Authorization", JWT)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// set mock mux parameters
	vars := map[string]string{
		"farmID":   "1",
		"deviceID": "1",
	}
	req = mux.SetURLVars(req, vars)

	// create a fake session
	fakeSession := createFakeSession()

	// mock the service call, passing a fake session and device ID to match the deviceID used in the GET request
	mockServiceCall := mockChannelService.On("GetByDeviceID", fakeSession, uint64(1)).Return(models, nil)

	// mock the JWT CreateSession call and return the fake session
	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

	// set the HTTP handler
	handler := http.HandlerFunc(channelRestService.List)

	// send the request and record the response
	handler.ServeHTTP(rr, req)

	// assert that the service expectations were met
	mockChannelService.AssertExpectations(t)

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

// Send a request to the real device channel update endpoint to
// ensure the router is sending requests to the proper REST and
// business logic service methods.
func TestChannelUpdateEndpoint(t *testing.T) {

	// Fake channel Update
	fakeChannel := &model.ChannelStruct{
		ID:   uint64(1),
		Name: "Test Channel 1"}

	// create mock REST and business logic services that stub methods used by the channels endpoint
	channelRestService := RestServiceRegistry.ChannelRestService()
	mockChannelService := new(servicemocks.MockChannelService)
	// create mock JWT middleware service to stub CreateSession
	mockJsonWebTokenService := new(restmocks.MockJsonWebTokenService)
	mockJsonWebTokenService.SetApp(App)

	// Replace the real channel service with the mocked service
	channelRestService.SetService(mockChannelService)
	channelRestService.SetMiddleware(mockJsonWebTokenService)

	jsonFakeChannel, err := json.Marshal(fakeChannel)
	assert.Nil(t, err)

	// Build the request for the GetPage endpoint with a new httptest.Recorder
	req, err := http.NewRequest("PUT", "/api/v1/farms/1/channels/1", strings.NewReader(string(jsonFakeChannel)))
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
	mockServiceCall := mockChannelService.On("Update", fakeSession, fakeChannel).Return(nil)

	// mock the JWT CreateSession call and return the fake session
	mockJsonWebTokenService.On("CreateSession", rr, req).Return(fakeSession, nil)

	// set the HTTP handler
	handler := http.HandlerFunc(channelRestService.Update)

	// send the request and record the response
	handler.ServeHTTP(rr, req)

	// assert that the service expectations were met
	mockChannelService.AssertExpectations(t)

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
