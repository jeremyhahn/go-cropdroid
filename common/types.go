package common

import (
	"net/http"
	"time"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
)

const (
	CONSISTENCY_CACHED = iota
	CONSISTENCY_LOCAL
	CONSISTENCY_QUORUM
)

type UserAccount interface {
	GetID() int
	GetEmail() string
	SetEmail(string)
	GetPassword() string
	SetPassword(string)
	GetToken() string
	GetRoles() []Role
	SetRoles([]Role)
	AddRole(Role)
}

type Role interface {
	GetID() int
	GetName() string
}

type HttpWriter interface {
	Write(w http.ResponseWriter, status int, response interface{})
	Error200(w http.ResponseWriter, err error)
	Error400(w http.ResponseWriter, err error)
	Error500(w http.ResponseWriter, err error)
}

type Notification interface {
	GetDevice() string
	GetPriority() int
	GetType() string
	GetTitle() string
	GetMessage() string
	GetTimestamp() string
	GetTimestampAsObject() time.Time
}

type Mailer interface {
	Send(farmName, subject, message string) error
}

// type DeviceService interface {
// 	GetDeviceConfig() config.DeviceConfig
// 	SetMetricValue(key string, value float64) error
// 	GetDeviceType() string
// 	GetConfig() (config.DeviceConfig, error)
// 	GetState() (state.DeviceStateMap, error)
// 	GetView() (DeviceView, error)
// 	GetHistory(metric string) ([]float64, error)
// 	GetDevice() (Device, error)
// 	Manage(farmState state.FarmStateMap)
// 	Poll(deviceStateChangeChan chan<- DeviceStateChange) error
// 	SetMode(mode string, device device.SmartSwitcher)
// 	Switch(channelID, position int, logMessage string) (*Switch, error)
// 	TimerSwitch(channelID, duration int, logMessage string) (TimerEvent, error)
// 	ManageMetrics(config config.DeviceConfig, farmState state.FarmStateMap) []error
// 	ManageChannels(deviceConfig config.DeviceConfig,
// 		farmState state.FarmStateMap, channels []config.ChannelConfig) []error
// 	//RegisterObserver(observer DeviceObserver)
// }

type DeviceView interface {
	GetMetrics() []Metric
	GetChannels() []Channel
	GetTimestamp() time.Time
}

type CommonDevice interface {
	GetID() uint64
	SetID(uint64)
	GetOrgID() int
	SetOrgID(int)
	GetType() string
	SetType(string)
	GetInterval() int
	SetInterval(int)
	GetDescription() string
	SetDescription(string)
	GetHardwareVersion() string
	SetHardwareVersion(string)
	GetFirmwareVersion() string
	SetFirmwareVersion(string)
	GetConfigs() map[string]string
	SetConfigs(map[string]string)
}

type Server interface {
	SetID(id int)
	GetID() int
	SetOrgID(id int)
	GetOrgID() int
	SetInterval(interval int)
	GetInterval() int
	SetTimezone(timezone string)
	GetTimezone() string
	SetMode(mode string)
	GetMode() string
	SetSmtp(smtp config.SmtpConfig)
	GetSmtp() config.SmtpConfig
	SetFarms(farms []Farm)
	GetFarms() []Farm
}

type Organization interface {
	GetID() int
	SetID(int)
	GetName() string
	SetName(string)
	SetFarms(farms []Farm)
	GetFarms() []Farm
	GetFarm(id int) (Farm, error)
}

type Farm interface {
	GetID() int
	SetID(int)
	GetOrgID() int
	SetOrgID(int)
	GetInterval() int
	SetInterval(int)
	GetMode() string
	SetMode(string)
	GetName() string
	SetName(string)
	GetDevices() []Device
	SetDevices([]Device)
}

type Device interface {
	CommonDevice
	IsEnabled() bool
	SetEnabled(enabled bool)
	IsNotify() bool
	SetNotify(notify bool)
	GetURI() string
	SetURI(uri string)
	GetMetric(key string) (Metric, error)
	GetMetrics() []Metric
	SetMetrics([]Metric)
	GetChannel(id int) (Channel, error)
	GetChannels() []Channel
	SetChannels([]Channel)
}

type Metric interface {
	config.MetricConfig
	SetValue(value float64)
	GetValue() float64
}

type Channel interface {
	SetValue(value int)
	GetValue() int
	config.ChannelConfig
}

type InAppPurchase interface {
	//GetOrderID() string
	//SetOrderID(string)
	GetProductID() string
	GetPurchaseToken() string
	GetPurchaseTimeMillis() int64
}

type DeviceStateChange struct {
	DeviceID    uint64
	DeviceType  string
	StateMap    state.DeviceStateMap
	IsPollEvent bool
}

// type MetricValueChanged struct {
// 	DeviceType string
// 	Key        string
// 	Value      float64
// }

// type SwitchValueChanged struct {
// 	DeviceType string
// 	ChannelID  int
// 	Value      int
// }

type FarmNotification struct {
	EventType string
	Message   string
}

type FarmError struct {
	Method    string
	EventType string
	Error     error
}

/*
type DeviceObserver interface {
	OnDeviceStateChange(diff DeviceState)
}*/
