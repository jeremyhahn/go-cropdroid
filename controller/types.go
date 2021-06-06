package controller

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/state"
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
