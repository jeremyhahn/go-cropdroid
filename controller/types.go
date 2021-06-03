package controller

import (
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/state"
)

type Controller interface {
	GetType() string
	State() (state.ControllerStateMap, error)
	Switch(channel, position int) (*common.Switch, error)
	TimerSwitch(channel, duration int) (common.TimerEvent, error)
}

type VirtualController interface {
	Controller
	WriteState(state state.ControllerStateMap) error
}
