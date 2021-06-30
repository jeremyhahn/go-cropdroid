package device

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/device/test"
	"github.com/jeremyhahn/go-cropdroid/device/test/mocks"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHttpGetType(t *testing.T) {

	deviceType := "unittest"

	httpDeviceType := &SmartSwitch{
		app:        nil,
		baseURL:    "http://localtest",
		deviceType: deviceType}

	assert.Equal(t, deviceType, httpDeviceType.GetType())
}

func TestHttpState(t *testing.T) {

	app := test.NewUnitTestSession()

	deviceType := "unittest"
	mockURI := "http://localtest"
	//mockEndpoint := fmt.Sprintf("%s/state", mockURI)
	mockHttpClient := &mocks.MockHttpClient{}

	metrics := map[string]float64{
		"test1": 1.5,
		"test2": 2.6}
	channels := []int{0, 1}
	timestamp := time.Now().In(app.Location)

	expected := state.NewDeviceStateMap()
	expected.SetMetrics(metrics)
	expected.SetChannels(channels)
	expected.SetTimestamp(timestamp)

	bodyReader := strings.NewReader("{\"metrics\":{\"test1\":1.5,\"test2\":2.6},\"channels\":[0,1]}")
	bodyCloser := io.NopCloser(bodyReader)

	mockResponse := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       bodyCloser,
	}

	mockHttpClient.On("Get", mock.Anything).Return(mockResponse, nil)

	httpDeviceType := CreateSmartSwitch(
		mockHttpClient,
		test.NewUnitTestSession(),
		mockURI,
		deviceType)

	deviceStateMap, err := httpDeviceType.State()
	deviceStateMap.SetTimestamp(timestamp)

	assert.Nil(t, err)
	assert.Equal(t, expected, deviceStateMap)
}

func TestHttpSwitch(t *testing.T) {

	deviceType := "unittest"
	mockURI := "http://localtest"
	//mockEndpoint := fmt.Sprintf("%s/state", mockURI)
	mockHttpClient := &mocks.MockHttpClient{}

	//timestamp := time.Now().In(app.Location)

	expected := &common.Switch{
		Channel: 1,
		Pin:     1,
		State:   1}

	bodyReader := strings.NewReader("{\"channel\":1,\"pin\":1,\"position\":1}")
	bodyCloser := io.NopCloser(bodyReader)

	mockResponse := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       bodyCloser,
	}

	mockHttpClient.On("Get", mock.Anything).Return(mockResponse, nil)

	httpDeviceType := CreateSmartSwitch(
		mockHttpClient,
		test.NewUnitTestSession(),
		mockURI,
		deviceType)

	switchState, err := httpDeviceType.Switch(1, 1)

	assert.Nil(t, err)
	assert.Equal(t, expected, switchState)
}

func TestHttpTimerSwitch(t *testing.T) {

	deviceType := "unittest"
	mockURI := "http://localtest"
	mockHttpClient := &mocks.MockHttpClient{}

	channel := 1
	duration := 5

	bodyReader := strings.NewReader("{\"channel\":1,\"duration\":5}")
	bodyCloser := io.NopCloser(bodyReader)

	mockResponse := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       bodyCloser,
	}

	mockHttpClient.On("Get", mock.Anything).Return(mockResponse, nil)

	httpDeviceType := CreateSmartSwitch(
		mockHttpClient,
		test.NewUnitTestSession(),
		mockURI,
		deviceType)

	timerEvent, err := httpDeviceType.TimerSwitch(channel, duration)

	assert.Nil(t, err)
	assert.Equal(t, channel, timerEvent.GetChannel())
	assert.Equal(t, duration, timerEvent.GetDuration())
}

func TestHttpDeviceInfo(t *testing.T) {

	deviceType := "unittest"
	mockURI := "http://localtest"
	mockHttpClient := &mocks.MockHttpClient{}

	hardwareVersion := "v0.0.1-test"
	firmwareVersion := "v0.0.2-test"
	uptime := int64(123456)

	bodyReader := strings.NewReader("{\"hardware\":\"v0.0.1-test\",\"firmware\":\"v0.0.2-test\",\"uptime\":123456}")
	bodyCloser := io.NopCloser(bodyReader)

	mockResponse := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       bodyCloser,
	}

	mockHttpClient.On("Get", mock.Anything).Return(mockResponse, nil)

	httpDeviceType := CreateSmartSwitch(
		mockHttpClient,
		test.NewUnitTestSession(),
		mockURI,
		deviceType)

	deviceInfo, err := httpDeviceType.SystemInfo()

	assert.Nil(t, err)
	assert.Equal(t, hardwareVersion, deviceInfo.GetHardwareVersion())
	assert.Equal(t, firmwareVersion, deviceInfo.GetFirmwareVersion())
	assert.Equal(t, uptime, deviceInfo.GetUptime())
}
