package service

import (
	"errors"
	"net/http"

	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/shoppingcart"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

var (
	//ErrFarmNotFound             = errors.New("farm config not found")
	ErrChannelNotFound          = errors.New("channel not found")
	ErrMetricNotFound           = errors.New("metric not found")
	ErrScheduleNotFound         = errors.New("channel schedule not found")
	ErrConditionNotFound        = errors.New("channel condition not found")
	ErrNoDeviceState            = errors.New("no device state")
	ErrCreateService            = errors.New("failed to create service")
	ErrDeviceNotFound           = errors.New("device not found")
	ErrWorkflowNotFound         = errors.New("workflow not found")
	ErrWorkflowStepNotFound     = errors.New("workflow step not found")
	ErrPermissionDenied         = errors.New("permission denied")
	ErrDeleteAdminAccount       = errors.New("admin account can't be deleted")
	ErrChangeAdminRole          = errors.New("admin role can't be changed")
	ErrResetPasswordUnsupported = errors.New("reset password feature unsupported by auth store")
)

type AlgorithmHandler interface {
	Handle() (bool, error)
}

type ScheduleHandler interface {
	Handle() error
}

type ConditionHandler interface {
	Handle() (bool, error)
}

type FarmChannels struct {
	FarmConfigChan        chan config.Farm
	FarmConfigChangeChan  chan config.Farm
	FarmStateChan         chan state.FarmStateMap
	FarmStateChangeChan   chan state.FarmStateMap
	FarmErrorChan         chan common.FarmError
	FarmNotifyChan        chan common.FarmNotification
	DeviceChangeChan      chan config.Device
	DeviceStateChangeChan chan common.DeviceStateChange
	DeviceStateDeltaChan  chan map[string]state.DeviceStateDeltaMap
	// MetricChangedChan      chan common.MetricValueChanged
	//SwitchChangedChan chan common.SwitchValueChanged
}

type UserCredentials struct {
	OrgID    uint64 `json:"orgId"`
	OrgName  string `json:"orgName"`
	Email    string `json:"email"`
	Password string `json:"password"`
	AuthType int    `json:"authType"`
}

type JsonWebTokenService interface {
	ParseToken(r *http.Request, extractor request.Extractor) (*jwt.Token, *JsonWebTokenClaims, error)
	GenerateToken(w http.ResponseWriter, req *http.Request)
	RefreshToken(w http.ResponseWriter, req *http.Request)
	Middleware
}

type Middleware interface {
	Validate(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
	CreateSession(w http.ResponseWriter, r *http.Request) (Session, error)
}

type OrganizationService interface {
	Create(organization *config.Organization) error
	GetAll(session Session) ([]*config.Organization, error)
	GetUsers(session Session) ([]*config.User, error)
	Delete(session Session) error
}

type FarmService interface {
	BuildDeviceServices() ([]DeviceService, error) // only used by builder/farm.go
	GetFarmID() uint64
	GetChannels() *FarmChannels
	GetConfig() *config.Farm
	//GetConfigClusterID() uint64
	GetConsistencyLevel() int
	GetPublicKey() string
	GetState() state.FarmStateMap
	GetStateID() uint64
	//OnLeaderUpdated(info raftio.LeaderInfo)
	InitializeState(saveToStateStore bool) error
	IsRunning() bool
	Poll()
	//PollCluster(raftCluster cluster.RaftCluster)
	PublishConfig(farmConfig *config.Farm) error
	PublishState(farmState state.FarmStateMap) error
	PublishDeviceState(deviceState map[string]state.DeviceStateMap) error
	PublishDeviceDelta(deviceState map[string]state.DeviceStateDeltaMap) error
	RefreshHardwareVersions() error
	Run()
	RunCluster()
	RunWorkflow(workflow *config.Workflow)
	SaveConfig(farmConfig *config.Farm) error
	SetConfig(farmConfig *config.Farm) error
	SetDeviceConfig(deviceConfig *config.Device) error
	SetDeviceState(deviceType string, deviceState state.DeviceStateMap)
	SetConfigValue(session Session, farmID, deviceID uint64, key, value string) error
	SetMetricValue(deviceType string, key string, value float64) error
	SetSwitchValue(deviceType string, channelID int, value int) error
	Stop()
	WatchConfig() <-chan config.Farm
	WatchState() <-chan state.FarmStateMap
	WatchDeviceState() <-chan map[string]state.DeviceStateMap
	WatchDeviceDeltas() <-chan map[string]state.DeviceStateDeltaMap
	WatchFarmStateChange()
}

type AuthService interface {
	Login(userCredentials *UserCredentials) (common.UserAccount, []*config.Organization, []*config.Farm, error)
	Register(userCredentials *UserCredentials, baseURI string) (common.UserAccount, error)
	Activate(registrationID uint64) (common.UserAccount, error)
	ResetPassword(userCredentials *UserCredentials) error
}

type UserService interface {
	CreateUser(user common.UserAccount) error
	UpdateUser(user common.UserAccount) error // replaced with SetPermnission?
	Delete(session Session, userID uint64) error
	DeletePermission(session Session, userID uint64) error
	//Get(email string) (common.UserAccount, error)
	Get(userID uint64) (common.UserAccount, error)
	SetPermission(session Session, permission *config.Permission) error
	// probably needs to be moved to auth service; not implemented in google_auth yet
	Refresh(userID uint64) (common.UserAccount, []*config.Organization, []*config.Farm, error)
	AuthService
}

type NotificationService interface {
	Enqueue(notification common.Notification) error
	Dequeue() <-chan common.Notification
	QueueSize() int
}

type EventLogService interface {
	GetFarmID() uint64
	Create(deviceID uint64, deviceName, eventType, message string)
	GetAll() []*entity.EventLog
	GetPage(page int64) *viewmodel.EventsPage
}

type ServiceRegistry interface {
	SetAlgorithmService(AlgorithmService)
	GetAlgorithmService() AlgorithmService
	SetAuthService(AuthService)
	GetAuthService() AuthService
	SetChangefeedService(changefeedService ChangefeedService)
	GetChangefeedService() ChangefeedService
	SetChannelService(ChannelService)
	GetChannelService() ChannelService
	SetConditionService(ConditionService)
	GetConditionService() ConditionService
	SetConfigService(configService ConfigService)
	GetConfigService() ConfigService
	SetDeviceFactory(DeviceFactory)
	GetDeviceFactory() DeviceFactory
	SetDeviceServices(farmID uint64, deviceService []DeviceService)
	GetDeviceServices(farmID uint64) ([]DeviceService, error)
	GetDeviceService(farmID uint64, deviceType string) (DeviceService, error)
	GetDeviceServiceByID(farmID uint64, deviceID uint64) (DeviceService, error)
	SetDeviceService(farmID uint64, deviceService DeviceService) (DeviceService, error)
	AddEventLogService(eventLogService EventLogService) error
	SetEventLogService(eventLogServices map[uint64]EventLogService)
	GetEventLogServices() map[uint64]EventLogService
	GetEventLogService(farmID uint64) EventLogService
	RemoveEventLogService(farmID uint64)
	SetFarmFactory(FarmFactory FarmFactory)
	GetFarmFactory() FarmFactory
	AddFarmService(farmService FarmService) error
	SetFarmServices(map[uint64]FarmService)
	GetFarmServices() map[uint64]FarmService
	GetFarmService(uint64) FarmService
	RemoveFarmService(farmID uint64)
	SetFarmProvisioner(farmProvisioner provisioner.FarmProvisioner)
	GetFarmProvisioner() provisioner.FarmProvisioner
	SetGoogleAuthService(googleAuthService AuthService)
	GetGoogleAuthService() AuthService
	SetJsonWebTokenService(JsonWebTokenService)
	GetJsonWebTokenService() JsonWebTokenService
	SetMetricService(MetricService)
	GetMetricService() MetricService
	SetNotificationService(NotificationService)
	GetNotificationService() NotificationService
	SetScheduleService(ScheduleService)
	GetScheduleService() ScheduleService
	SetShoppingCartService(shoppingcart.ShoppingCartService)
	GetShoppingCartService() shoppingcart.ShoppingCartService
	SetOrganizationService(organizationService OrganizationService)
	GetOrganizationService() OrganizationService
	SetRoleService(roleService RoleService)
	GetRoleService() RoleService
	SetUserService(UserService)
	GetUserService() UserService
	SetWorkflowService(WorkflowService)
	GetWorkflowService() WorkflowService
	SetWorkflowStepService(WorkflowStepService)
	GetWorkflowStepService() WorkflowStepService
}

type ConfigService interface {
	GetServerConfig() config.Server
	//SetValue(deviceID int, key, value string) error
	Sync()
	SetValue(session Session, farmID, deviceID uint64, key, value string) error
	SetDevice(deviceConfig config.Device)
	NotifyConfigChange(farmID uint64)
	OnMetricChange(metric config.Metric)
	OnChannelChange(channel config.Channel)
	OnDeviceChange(deviceConfig config.DeviceSetting)
	OnConditionChange(condition config.Condition)
	OnScheduleChange(schedule config.Schedule)
}

type ChangefeedService interface {
	Subscribe()
	FeedCount() int
	OnDeviceConfigChange(changefeed datastore.Changefeed)
	OnChannelConfigChange(changefeed datastore.Changefeed)
	OnConditionConfigChange(changefeed datastore.Changefeed)
	OnScheduleConfigChange(changefeed datastore.Changefeed)
	OnMetricConfigChange(changefeed datastore.Changefeed)
	OnDeviceStateChange(changefeed datastore.Changefeed)
}

type DeviceService interface {
	//GetDevice() config.Device
	SetMetricValue(key string, value float64) error
	GetDeviceType() string
	GetConfig() (*config.Device, error)
	GetID() uint64
	GetState() (state.DeviceStateMap, error)
	GetView() (common.DeviceView, error)
	GetHistory(metric string) ([]float64, error)
	GetDevice() (common.Device, error)
	Manage(farmState state.FarmStateMap)
	Poll() error
	SetConfig(config *config.Device) error
	SetMode(mode string, device device.IOSwitcher)
	SetState(deviceStateMap state.DeviceStateMap) error
	Stop()
	Switch(channelID, position int, logMessage string) (*common.Switch, error)
	TimerSwitch(channelID, duration int, logMessage string) (common.TimerEvent, error)
	ManageMetrics(config config.Device, farmState state.FarmStateMap) []error
	ManageChannels(deviceConfig config.Device,
		farmState state.FarmStateMap, channels []config.Channel) []error
	GetChannelConfig(channelID int) (*config.Channel, error)
	RefreshSystemInfo() error
	//RegisterObserver(observer DeviceObserver)
}

type RoleService interface {
	GetAll() ([]*config.Role, error)
}
