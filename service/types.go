package service

import (
	"errors"
	"net/http"

	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/jeremyhahn/go-cropdroid/provisioner"

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
	ErrFarmConfigNotFound   = errors.New("farm config not found")
	ErrChannelNotFound      = errors.New("channel not found")
	ErrMetricNotFound       = errors.New("metric not found")
	ErrScheduleNotFound     = errors.New("channel schedule not found")
	ErrConditionNotFound    = errors.New("channel condition not found")
	ErrNoDeviceState        = errors.New("no device state")
	ErrCreateService        = errors.New("failed to create service")
	ErrDeviceNotFound       = errors.New("device not found")
	ErrWorkflowNotFound     = errors.New("workflow not found")
	ErrWorkflowStepNotFound = errors.New("workflow step not found")
	//ErrDeviceDisabled     = errors.New("Device disabled")
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
	FarmConfigChan         chan config.FarmConfig
	FarmConfigChangeChan   chan config.FarmConfig
	FarmStateChan          chan state.FarmStateMap
	FarmStateChangeChan    chan state.FarmStateMap
	FarmErrorChan          chan common.FarmError
	FarmNotifyChan         chan common.FarmNotification
	DeviceConfigChangeChan chan config.DeviceConfig
	DeviceStateChangeChan  chan common.DeviceStateChange
	DeviceStateDeltaChan   chan map[string]state.DeviceStateDeltaMap
	// MetricChangedChan      chan common.MetricValueChanged
	//SwitchChangedChan chan common.SwitchValueChanged
}

type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	AuthType int    `json:"authType"`
}

type JsonWebTokenService interface {
	ParseToken(r *http.Request, extractor request.Extractor) (*jwt.Token, *JsonWebTokenClaims, error)
	GenerateToken(w http.ResponseWriter, req *http.Request)
	Middleware
}

type Middleware interface {
	Validate(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
	CreateSession(w http.ResponseWriter, r *http.Request) (Session, error)
}

type AuthService interface {
	Get(email string) (common.UserAccount, error)
	Login(userCredentials *UserCredentials, farmProvisioner provisioner.FarmProvisioner) (common.UserAccount, []config.OrganizationConfig, error)
	Register(userCredentials *UserCredentials) (common.UserAccount, error)
}

type UserService interface {
	CreateUser(user common.UserAccount)
	GetCurrentUser() (common.UserAccount, error)
	//GetUserByID(userId int) (common.UserAccount, error)
	GetUserByEmail(email string) (common.UserAccount, error)
	GetRole(userID, orgID int) (config.RoleConfig, error)
	//GetOrganizations(userID int) ([]entity.Organization, error)
	AuthService
}

type NotificationService interface {
	Enqueue(notification common.Notification) error
	Dequeue() <-chan common.Notification
	QueueSize() int
}

type EventLogService interface {
	Create(event, message string)
	GetAll() []entity.EventLog
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
	SetDeviceService(farmID uint64, deviceService DeviceService) (DeviceService, error)
	SetEventLogService(EventLogService)
	GetEventLogService() EventLogService
	SetFarmFactory(FarmFactory *FarmFactory)
	GetFarmFactory() *FarmFactory
	AddFarmService(farmService FarmService) error
	SetFarmServices(map[uint64]FarmService)
	GetFarmServices() map[uint64]FarmService
	GetFarmService(uint64) FarmService
	SetGoogleAuthService(googleAuthService AuthService)
	GetGoogleAuthService() AuthService
	SetJsonWebTokenService(JsonWebTokenService)
	GetJsonWebTokenService() JsonWebTokenService
	SetMailer(common.Mailer)
	GetMailer() common.Mailer
	SetMetricService(MetricService)
	GetMetricService() MetricService
	SetNotificationService(NotificationService)
	GetNotificationService() NotificationService
	SetScheduleService(ScheduleService)
	GetScheduleService() ScheduleService
	SetUserService(UserService)
	GetUserService() UserService
	SetWorkflowService(WorkflowService)
	GetWorkflowService() WorkflowService
	SetWorkflowStepService(WorkflowStepService)
	GetWorkflowStepService() WorkflowStepService
}

type ConfigService interface {
	GetServerConfig() config.ServerConfig
	//SetValue(deviceID int, key, value string) error
	Sync()
	SetValue(session Session, farmID, deviceID uint64, key, value string) error
	SetDeviceConfig(deviceConfig config.DeviceConfig)
	NotifyConfigChange(farmID uint64)
	OnMetricChange(metric config.MetricConfig)
	OnChannelChange(channel config.ChannelConfig)
	OnDeviceConfigChange(deviceConfig config.DeviceConfigConfig)
	OnConditionChange(condition config.ConditionConfig)
	OnScheduleChange(schedule config.ScheduleConfig)
}

type ChangefeedService interface {
	Subscribe()
	FeedCount() int
	OnDeviceConfigConfigChange(changefeed datastore.Changefeed)
	OnChannelConfigChange(changefeed datastore.Changefeed)
	OnConditionConfigChange(changefeed datastore.Changefeed)
	OnScheduleConfigChange(changefeed datastore.Changefeed)
	OnMetricConfigChange(changefeed datastore.Changefeed)
	OnDeviceStateChange(changefeed datastore.Changefeed)
}

type FarmService interface {
	BuildDeviceServices() ([]DeviceService, error) // only used by builder/farm.go
	GetFarmID() uint64
	GetChannels() *FarmChannels
	GetConfig() config.FarmConfig
	GetConfigClusterID() uint64
	GetConsistencyLevel() int
	GetPublicKey() string
	GetState() state.FarmStateMap
	//OnLeaderUpdated(info raftio.LeaderInfo)
	Poll()
	PollCluster()
	PublishConfig(farmConfig config.FarmConfig) error
	PublishState(farmState state.FarmStateMap) error
	PublishDeviceState(deviceState map[string]state.DeviceStateMap) error
	PublishDeviceDelta(deviceState map[string]state.DeviceStateDeltaMap) error
	Run()
	RunCluster()
	SetConfig(farmConfig config.FarmConfig) error
	SetDeviceConfig(deviceConfig config.DeviceConfig) error
	SetDeviceState(deviceType string, deviceState state.DeviceStateMap)
	SetConfigValue(session Session, farmID, deviceID uint64, key, value string) error
	SetMetricValue(deviceType string, key string, value float64) error
	SetSwitchValue(deviceType string, channelID int, value int) error
	//SetState(state state.FarmStateMap
	//Stop()
	WatchConfig() <-chan config.FarmConfig
	WatchState() <-chan state.FarmStateMap
	WatchDeviceState() <-chan map[string]state.DeviceStateMap
	WatchDeviceDeltas() <-chan map[string]state.DeviceStateDeltaMap
	WatchFarmStateChange()
}

type DeviceService interface {
	GetDeviceConfig() config.DeviceConfig
	SetMetricValue(key string, value float64) error
	GetDeviceType() string
	GetConfig() (config.DeviceConfig, error)
	GetState() (state.DeviceStateMap, error)
	GetView() (common.DeviceView, error)
	GetHistory(metric string) ([]float64, error)
	GetDevice() (common.Device, error)
	Manage(farmState state.FarmStateMap)
	Poll() error
	SetConfig(config config.DeviceConfig) error
	SetMode(mode string, device device.IOSwitcher)
	Switch(channelID, position int, logMessage string) (*common.Switch, error)
	TimerSwitch(channelID, duration int, logMessage string) (common.TimerEvent, error)
	ManageMetrics(config config.DeviceConfig, farmState state.FarmStateMap) []error
	ManageChannels(deviceConfig config.DeviceConfig,
		farmState state.FarmStateMap, channels []config.ChannelConfig) []error
	//RegisterObserver(observer DeviceObserver)
}
