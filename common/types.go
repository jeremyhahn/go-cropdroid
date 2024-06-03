package common

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
)

const (
	CONSISTENCY_LOCAL = iota
	CONSISTENCY_QUORUM
)

var (
	ErrClusterNotFound = errors.New("cluster not found")
)

// type UserAccount interface {
// 	GetID() uint64
// 	SetID(id uint64)
// 	GetEmail() string
// 	SetEmail(string)
// 	GetPassword() string
// 	SetPassword(string)
// 	GetRoles() []Role
// 	SetRoles([]Role)
// 	AddRole(Role)
// 	HasRole(name string) bool
// 	SetOrganizationRefs(ids []uint64)
// 	GetOrganizationRefs() []uint64
// 	SetFarmRefs(ids []uint64)
// 	GetFarmRefs() []uint64
// }

// type Role interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	GetName() string
// 	SetName(name string)
// }

// type HttpWriter interface {
// 	Write(w http.ResponseWriter, status int, response interface{})
// 	Success200(w http.ResponseWriter, response interface{})
// 	Success404(w http.ResponseWriter, payload interface{}, err error)
// 	Error200(w http.ResponseWriter, err error)
// 	Error400(w http.ResponseWriter, err error)
// 	Error500(w http.ResponseWriter, err error)
// }

// type Notification interface {
// 	GetDevice() string
// 	GetPriority() int
// 	GetType() string
// 	GetTitle() string
// 	GetMessage() string
// 	GetTimestamp() string
// 	GetTimestampAsObject() time.Time
// }

type Mailer interface {
	SetRecipient(recipient string)
	Send(subject, message string) error
	SendHtml(template, subject string, data interface{}) (bool, error)
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

// type CommonDevice interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	GetOrgID() int
// 	SetOrgID(int)
// 	GetType() string
// 	SetType(string)
// 	GetInterval() int
// 	SetInterval(int)
// 	GetDescription() string
// 	SetDescription(string)
// 	GetHardwareVersion() string
// 	SetHardwareVersion(string)
// 	GetFirmwareVersion() string
// 	SetFirmwareVersion(string)
// 	GetConfigs() map[string]string
// 	SetConfigs(map[string]string)
// }

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
	SetSmtp(smtp config.Smtp)
	GetSmtp() config.Smtp
	SetFarms(farms []Farm)
	GetFarms() []Farm
}

type Organization interface {
	ID() int
	SetID(int)
	Name() string
	SetName(string)
	Farms() []Farm
	SetFarms(farms []Farm)
	GetFarm(id int) (Farm, error)
}

type Farm interface {
	ID() int
	SetID(int)
	OrgID() int
	SetOrgID(int)
	Interval() int
	SetInterval(int)
	Mode() string
	SetMode(string)
	Name() string
	SetName(string)
	Devices() []config.Device
	SetDevices([]config.Device)
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

// type DeviceStore interface {
// 	//datastore.DeviceDataStore
// 	state.DeviceStorer
// }

type ProvisionerParams struct {
	UserID           uint64
	RoleID           uint64
	OrganizationID   uint64
	FarmName         string
	ConfigStoreType  int
	StateStoreType   int
	DataStoreType    int
	ConsistencyLevel int
}
