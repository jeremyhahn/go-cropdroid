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

type IOSwitcher interface {
	GetType() string
	State() (state.DeviceStateMap, error)
	Switch(channel, position int) (*common.Switch, error)
	TimerSwitch(channel, duration int) (common.TimerEvent, error)
	SystemInfo() (DeviceInfo, error)
}

type VirtualIOSwitcher interface {
	IOSwitcher
	WriteState(state state.DeviceStateMap) error
}

type DeviceInfo interface {
	GetHardwareVersion() string
	GetFirmwareVersion() string
	GetUptime() int64
}
