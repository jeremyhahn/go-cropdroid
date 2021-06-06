// +build broken

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/stretchr/testify/assert"
)

func TestJsonWebTokenService(t *testing.T) {

	app := NewIntegrationTest().app

	user := &model.User{
		Email:    "root@localhost",
		Password: "dev"}

	farmDAO := gormstore.NewFarmDAO(app.Logger, app.GORM)
	provisioner := NewUserProvisioner(app, farmDAO)

	farm1, err := provisioner.CreateUserFarm(user)
	assert.Nil(t, err)

	assert.Nil(t, err)
	assert.NotNil(t, farm1)
	assert.Equal(t, 4, len(farm1.GetControllers()))

	farms, err := farmDAO.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(farms))

	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	controllerMapper := mapper.NewControllerMapper(metricMapper, channelMapper)

	datastoreRegistry := gormstore.NewGormRegistry(app.Logger, app.GORM)
	mapperRegistry := mapper.CreateRegistry()
	serviceRegistry := CreateServiceRegistry(app, datastoreRegistry, mapperRegistry)

	jwtService, err := NewJsonWebTokenService(app, farmDAO, controllerMapper, serviceRegistry, common.NewJsonWriter())
	assert.Nil(t, err)
	assert.NotNil(t, jwtService)

	// Assert GenerateToken works with proper credentials
	_bytes, err := json.Marshal(&UserCredentials{
		Email:    "root@localhost",
		Password: "dev"})
	assert.Nil(t, err)
	jsonCreds := bytes.NewBuffer(_bytes)
	request := httptest.NewRequest("GET", "/foo", jsonCreds)
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(jwtService.GenerateToken)
	handler.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, true, strings.Contains(response.Body.String(), "token"))

	fmt.Printf("HTTP %d: %s", response.Code, response.Body.String())

	// Assert GenerateToken fails with improper credentials
	_bytes, err = json.Marshal(&UserCredentials{
		Email:    "root@localhost",
		Password: "dev2"})
	assert.Nil(t, err)
	jsonCreds = bytes.NewBuffer(_bytes)
	request = httptest.NewRequest("GET", "/foo", jsonCreds)
	response = httptest.NewRecorder()
	handler = http.HandlerFunc(jwtService.GenerateToken)
	handler.ServeHTTP(response, request)
	assert.Equal(t, 403, response.Code)
	assert.Equal(t, true, strings.Contains(response.Body.String(), "Invalid credentials"))

	//fmt.Printf("HTTP %d: %s", response.Code, response.Body.String())

	CurrentTest.Cleanup()
}
