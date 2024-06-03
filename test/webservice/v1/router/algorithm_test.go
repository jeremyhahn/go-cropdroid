package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/stretchr/testify/assert"

	servicemocks "github.com/jeremyhahn/go-cropdroid/test/mocks/service"
)

func TestAlgorithmPageRequest(t *testing.T) {

	// t.Parallel()

	// Build a fake page result for the test
	pageQuery := query.NewPageQuery()
	pageResult := dao.NewPageResultFromQuery[*config.AlgorithmStruct](pageQuery)
	for i := 0; i < pageQuery.PageSize; i++ {
		pageResult.Entities[i] = &config.AlgorithmStruct{
			ID:   uint64(i),
			Name: fmt.Sprintf("Test Algorithm %d", i)}
	}

	// algorithmRestService- := new(restmocks.MockAlgorithmRestService)
	algorithmRestService := RestServiceRegistry.AlgorithmRestService()
	algorithmService := new(servicemocks.MockAlgorithmService)

	// Mock the service call
	mockServiceCall := algorithmService.On("Page", pageQuery, common.CONSISTENCY_LOCAL).Return(pageResult, nil)

	// Replace the real algorithm service with the mocked service
	algorithmRestService.SetService(algorithmService)

	// Build the request for the GetPage endpoint with a new httptest.Recorder
	req, err := http.NewRequest("GET", "/api/v1/algorithms/1", nil)
	req.Header.Set("Authorization", JWT)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// Set mock mux parameters
	vars := map[string]string{
		"page": "1",
	}
	req = mux.SetURLVars(req, vars)

	// Set the handler function an serve the request
	handler := http.HandlerFunc(algorithmRestService.Page)

	// Send the request and record the response
	handler.ServeHTTP(rr, req)

	// assert that the service expectations were met
	algorithmService.AssertExpectations(t)

	// remove the handler so another one can be added to take precedence
	mockServiceCall.Unset()

	webServiceResponse := &response.WebServiceResponse{
		Code:    200,
		Success: true,
		Payload: pageResult}

	jsonWebServiceResponse, err := json.Marshal(webServiceResponse)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, string(jsonWebServiceResponse), rr.Body.String())
}
