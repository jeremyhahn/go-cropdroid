package service

import (
	"errors"
	"net/http"

	"github.com/jeremyhahn/cropdroid/provisioner"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/datastore"
	"github.com/jeremyhahn/cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/cropdroid/state"
	"github.com/jeremyhahn/cropdroid/viewmodel"
)

var (
	ErrFarmConfigNotFound = errors.New("Farm config not found")
	ErrControllerNotFound = errors.New("Farm controller not found")
	ErrChannelNotFound    = errors.New("Farm channel not found")
	ErrMetricNotFound     = errors.New("Farm metric not found")
	ErrScheduleNotFound   = errors.New("Farm channel schedule not found")
	ErrConditionNotFound  = errors.New("Farm channel condition not found")
)

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
	SetControllerFactory(ControllerFactory)
	GetControllerFactory() ControllerFactory
	SetControllerServices(farmID int, controllerServices []common.ControllerService)
	GetControllerServices(farmID int) ([]common.ControllerService, error)
	SetEventLogService(EventLogService)
	GetEventLogService() EventLogService
	SetFarmFactory(FarmFactory *FarmFactory)
	GetFarmFactory() *FarmFactory
	AddFarmService(farmService FarmService) error
	SetFarmServices(map[int]FarmService)
	GetFarmServices() map[int]FarmService
	GetFarmService(int) (FarmService, bool)
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
}

type ConfigService interface {
	GetServerConfig() config.ServerConfig
	//SetValue(controllerID int, key, value string) error
	Sync()
	SetValue(session Session, farmID, controllerID int, key, value string) error
	SetControllerConfig(controllerConfig config.ControllerConfig)
	NotifyConfigChange(farmID int)
	OnMetricChange(metric config.MetricConfig)
	OnChannelChange(channel config.ChannelConfig)
	OnControllerConfigChange(controllerConfig config.ControllerConfigConfig)
	OnConditionChange(condition config.ConditionConfig)
	OnScheduleChange(schedule config.ScheduleConfig)
}

type ChangefeedService interface {
	Subscribe()
	FeedCount() int
	OnControllerConfigConfigChange(changefeed datastore.Changefeed)
	OnChannelConfigChange(changefeed datastore.Changefeed)
	OnConditionConfigChange(changefeed datastore.Changefeed)
	OnScheduleConfigChange(changefeed datastore.Changefeed)
	OnMetricConfigChange(changefeed datastore.Changefeed)
	OnControllerStateChange(changefeed datastore.Changefeed)
}

type FarmService interface {
	BuildControllerServices() ([]common.ControllerService, error) // only used by builder/farm.go
	GetFarmID() int
	GetConfig() config.FarmConfig
	GetConfigClusterID() uint64
	GetState() state.FarmStateMap
	//OnLeaderUpdated(info raftio.LeaderInfo)
	Poll()
	PollCluster()
	PublishConfig() error
	PublishState() error
	PublishControllerState(controllerState map[string]state.ControllerStateMap) error
	PublishControllerDelta(controllerState map[string]state.ControllerStateDeltaMap) error
	Run()
	RunCluster()
	SetConfig(farmConfig config.FarmConfig) error
	SetControllerConfig(controllerConfig config.ControllerConfig)
	SetControllerState(controllerType string, controllerState state.ControllerStateMap)
	SetConfigValue(session Session, farmID, controllerID int, key, value string) error
	SetMetricValue(controllerType string, key string, value float64) error
	SetSwitchValue(controllerType string, channelID int, value int) error
	//SetState(state state.FarmStateMap
	//Stop()
	WatchConfig() <-chan config.FarmConfig
	WatchState() <-chan state.FarmStateMap
	WatchControllerState() <-chan map[string]state.ControllerStateMap
	WatchControllerDeltas() <-chan map[string]state.ControllerStateDeltaMap
	WatchFarmStateChange()
}
