package device

import (
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/device/test"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/stretchr/testify/assert"
)

func TestGetType(t *testing.T) {

	app := test.NewUnitTestSession()

	deviceType := "unittest"
	farmStateMap := state.NewFarmStateMap(0)
	virtDeviceType := NewVirtualIOSwitch(app, farmStateMap,
		"http://localtest", deviceType)

	assert.Equal(t, deviceType, virtDeviceType.GetType())
}

func TestVirtualState(t *testing.T) {

	app := test.NewUnitTestSession()

	deviceType := "unittest"
	mockURI := "http://localtest"

	metrics := map[string]float64{
		"test1": 1.5,
		"test2": 2.6}
	channels := []int{0, 1}
	timestamp := time.Now().In(app.Location)
	stateFile := "/tmp/cropdroid-state-file"

	expected := state.NewDeviceStateMap()
	expected.SetMetrics(metrics)
	expected.SetChannels(channels)
	expected.SetTimestamp(timestamp)

	virtualHttpDeviceType := CreateVirtualIOSwitch(
		nil,
		app,
		state.NewFarmStateMap(0),
		mockURI,
		deviceType,
		stateFile)

	virtualHttpDeviceType.WriteState(expected)

	deviceStateMap, err := virtualHttpDeviceType.State()
	deviceStateMap.SetTimestamp(timestamp)

	assert.Nil(t, err)
	assert.Equal(t, expected, deviceStateMap)
}

func TestVirtualSwitch(t *testing.T) {

	app := test.NewUnitTestSession()

	deviceType := "unittest"
	mockURI := "http://localtest"

	metrics := map[string]float64{
		"test1": 1.5,
		"test2": 2.6}
	channels := []int{1, 1}
	timestamp := time.Now().In(app.Location)
	stateFile := "/tmp/cropdroid-state-file"

	deviceStateMap := state.NewDeviceStateMap()
	deviceStateMap.SetMetrics(metrics)
	deviceStateMap.SetChannels(channels)
	deviceStateMap.SetTimestamp(timestamp)

	expected := &common.Switch{
		Channel: 1,
		State:   1}

	virtualHttpDeviceType := CreateVirtualIOSwitch(
		nil,
		app,
		state.NewFarmStateMap(0),
		mockURI,
		deviceType,
		stateFile)

	virtualHttpDeviceType.WriteState(deviceStateMap)

	switchState, err := virtualHttpDeviceType.Switch(1, 1)

	assert.Nil(t, err)
	assert.Equal(t, expected, switchState)
}

func TestVirtualTimerSwitch(t *testing.T) {

	app := test.NewUnitTestSession()

	deviceType := "unittest"
	mockURI := "http://localtest"
	stateFile := "/tmp/cropdroid-state-file"

	channel := 1
	duration := 5

	virtualHttpDeviceType := CreateVirtualIOSwitch(
		nil,
		app,
		state.NewFarmStateMap(0),
		mockURI,
		deviceType,
		stateFile)

	timerEvent, err := virtualHttpDeviceType.TimerSwitch(channel, duration)

	assert.Nil(t, err)
	assert.Equal(t, channel, timerEvent.GetChannel())
	assert.Equal(t, duration, timerEvent.GetDuration())
}

func TestVirtualDeviceInfo(t *testing.T) {

	app := test.NewUnitTestSession()

	deviceType := "unittest"
	mockURI := "http://localtest"
	stateFile := "/tmp/cropdroid-state-file"

	hardwareVersion := "v0.0.1a"
	firmwareVersion := "v0.0.1a"
	uptime := int64(123456)

	virtualHttpDeviceType := CreateVirtualIOSwitch(
		nil,
		app,
		state.NewFarmStateMap(0),
		mockURI,
		deviceType,
		stateFile)

	deviceInfo, err := virtualHttpDeviceType.SystemInfo()

	assert.Nil(t, err)
	assert.Equal(t, hardwareVersion, deviceInfo.GetHardwareVersion())
	assert.Equal(t, firmwareVersion, deviceInfo.GetFirmwareVersion())
	assert.Equal(t, uptime, deviceInfo.GetUptime())
}
