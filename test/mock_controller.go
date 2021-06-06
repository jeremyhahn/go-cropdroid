// +build broken

package test

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/controller"
	"github.com/stretchr/testify/mock"
)

type MockController struct {
	controller.Controller
	mock.Mock
}

func NewMockController() *MockController {
	return &MockController{}
}

func (client *MockController) GetType() string {
	args := client.Called()
	fmt.Print("[MockController.GetType] called")
	return args.Get(0).(string)
}

func (client *MockController) Switch(channel, state int) (*common.Switch, error) {
	args := client.Called(channel, state)
	fmt.Printf("[MockController.Switch] channel=%d, state=%d\n", channel, state)
	return args.Get(0).(*common.Switch), args.Error(1)
}

func (client *MockController) TimerSwitch(channel, duration int) (common.TimerEvent, error) {
	args := client.Called(channel, duration)
	fmt.Printf("[MockController.TimedSwitch] channel=%d, duration=%d\n", channel, duration)
	return args.Get(0).(common.TimerEvent), args.Error(1)
}

func (client *MockController) State() (common.ControllerStateMap, error) {
	args := client.Called()
	fmt.Print("[MockController.State] called")
	return args.Get(0).(common.ControllerStateMap), args.Error(1)
}
