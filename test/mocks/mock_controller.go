// +build broken

package mocks

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/stretchr/testify/mock"
)

type MockDevice struct {
	device.Device
	mock.Mock
}

func NewMockDevice() *MockDevice {
	return &MockDevice{}
}

func (client *MockDevice) GetType() string {
	args := client.Called()
	fmt.Print("[MockDevice.GetType] called")
	return args.Get(0).(string)
}

func (client *MockDevice) Switch(channel, state int) (*common.Switch, error) {
	args := client.Called(channel, state)
	fmt.Printf("[MockDevice.Switch] channel=%d, state=%d\n", channel, state)
	return args.Get(0).(*common.Switch), args.Error(1)
}

func (client *MockDevice) TimerSwitch(channel, duration int) (common.TimerEvent, error) {
	args := client.Called(channel, duration)
	fmt.Printf("[MockDevice.TimedSwitch] channel=%d, duration=%d\n", channel, duration)
	return args.Get(0).(common.TimerEvent), args.Error(1)
}

func (client *MockDevice) State() (common.DeviceStateMap, error) {
	args := client.Called()
	fmt.Print("[MockDevice.State] called")
	return args.Get(0).(common.DeviceStateMap), args.Error(1)
}
