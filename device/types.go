package device

import (
	"net/http"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/state"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
}

type SmartSwitcher interface {
	GetType() string
	State() (state.DeviceStateMap, error)
	Switch(channel, position int) (*common.Switch, error)
	TimerSwitch(channel, duration int) (common.TimerEvent, error)
}

type VirtualSmartSwitcher interface {
	SmartSwitcher
	WriteState(state state.DeviceStateMap) error
}
