package config

import "time"

const CONFIG_KEY_NAME = "name"
const CONFIG_KEY_MODE = "mode"
const CONFIG_KEY_INTERVAL = "interval"
const CONFIG_KEY_TIMEZONE = "timezone"
const CONFIG_KEY_ENABLE = "enable"
const CONFIG_KEY_NOTIFY = "notify"
const CONFIG_KEY_URI = "uri"

type ServerConfig interface {
	GetID() int
	SetID(int)
	GetOrganizations() []Organization
	SetOrganizations([]Organization)
	GetOrganization(id int) (*Organization, error)
	SetInterval(int)
	GetInterval() int
	SetTimezone(string)
	GetTimezone() string
	SetMode(string)
	GetMode() string
	SetSmtp(*Smtp)
	GetSmtp() *Smtp
	GetLicense() *License
	SetLicense(license *License)
	GetFarms() []Farm
	SetFarms(farms []Farm)
	SetFarm(id int, farm FarmConfig)
}

type OrganizationConfig interface {
	GetID() int
	SetID() int
	GetName() string
	SetName(string)
	//SetFarms(farms []Farm)
	//GetFarms() []Farm
	//GetFarm(id int) (Farm, error)

	AddFarm(farm Farm)
	SetFarms(farms []Farm)
	GetFarms() []Farm
	GetFarm(id int) (*Farm, error)

	SetUsers(users []User)
	GetUsers() []User
	GetLicense() *License
	SetLicense(*License)
}

type FarmConfig interface {
	SetID(int)
	GetID() int
	SetOrganizationID(id int)
	GetOrganizationID() int
	GetReplicas() int
	SetReplicas(count int)
	SetName(string)
	GetName() string
	SetMode(string)
	GetMode() string
	GetOrgID() int
	SetOrgID(id int)
	GetInterval() int
	SetInterval(int)
	//SetTimezone(tz *time.Location)
	//GetTimezone() *time.Location
	SetTimezone(tz string)
	GetTimezone() string
	GetSmtp() SmtpConfig
	SetUsers(users []User)
	GetUsers() []User
	AddController(Controller)
	GetControllers() []Controller
	SetControllers([]Controller)
	SetController(controller ControllerConfig)
	GetController(controllerType string) (*Controller, error)
	GetControllerById(id int) (*Controller, error)
	ParseConfigs() error
	HydrateConfigs() error
}

type SmtpConfig interface {
	IsEnabled() bool
	SetEnable(enabled bool)
	SetHost(string)
	GetHost() string
	SetPort(int)
	GetPort() int
	SetUsername(string)
	GetUsername() string
	SetPassword(string)
	GetPassword() string
	SetRecipient(string)
	GetRecipient() string
}

type CommonControllerConfig interface {
	GetID() int
	SetID(int)
	GetFarmID() int
	SetFarmID(int)
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
	GetConfigs() []ControllerConfigItem
	SetConfigs([]ControllerConfigItem)
	SetConfig(controllerConfig ControllerConfigConfig)
	GetConfigMap() map[string]string
}

type ControllerConfig interface {
	CommonControllerConfig
	IsEnabled() bool
	IsNotify() bool
	GetURI() string
	GetMetric(key string) (*Metric, error)
	GetMetrics() []Metric
	SetMetrics([]Metric)
	SetMetric(metric MetricConfig)
	GetChannel(id int) (*Channel, error)
	GetChannels() []Channel
	SetChannel(channel ChannelConfig)
	SetChannels([]Channel)
	ParseConfigs() error
	HydrateConfigs() error
}

type ControllerConfigConfig interface {
	SetID(int)
	GetID() int
	SetUserID(int)
	GetUserID() int
	SetControllerID(int)
	GetControllerID() int
	SetKey(string)
	GetKey() string
	SetValue(string)
	GetValue() string
}

type MetricConfig interface {
	GetID() int
	SetID(int)
	GetControllerID() int
	SetControllerID(int)
	GetDataType() int
	SetDataType(int)
	GetKey() string
	SetKey(string)
	GetName() string
	SetName(string)
	IsEnabled() bool
	SetEnable(bool)
	IsNotify() bool
	SetNotify(bool)
	GetUnit() string
	SetUnit(string)
	GetAlarmLow() float64
	SetAlarmLow(float64)
	GetAlarmHigh() float64
	SetAlarmHigh(float64)
}

type ChannelConfig interface {
	GetID() int
	SetID(int)
	GetControllerID() int
	SetControllerID(int)
	GetChannelID() int
	SetChannelID(int)
	GetName() string
	SetName(name string)
	IsEnabled() bool
	SetEnable(bool)
	IsNotify() bool
	SetNotify(bool)
	AddCondition(condition ConditionConfig)
	GetConditions() []Condition
	SetConditions(conditions []Condition)
	SetCondition(condition ConditionConfig)
	GetSchedule() []Schedule
	SetSchedule(schedule []Schedule)
	SetScheduleItem(schedule ScheduleConfig)
	GetDuration() int
	SetDuration(int)
	GetDebounce() int
	SetDebounce(int)
	GetBackoff() int
	SetBackoff(int)
	GetAlgorithmID() int
	SetAlgorithmID(int)
}

type ConditionConfig interface {
	SetID(uint64)
	GetID() uint64
	SetChannelID(int)
	GetChannelID() int
	GetMetricID() int
	SetMetricID(int)
	SetComparator(string)
	GetComparator() string
	SetThreshold(float64)
	GetThreshold() float64
	Hash() uint64
}

type ScheduleConfig interface {
	GetID() uint64
	SetID(uint64)
	GetChannelID() int
	SetStartDate(time.Time)
	GetStartDate() time.Time
	SetEndDate(*time.Time)
	GetEndDate() *time.Time
	// recurring options
	SetFrequency(int)
	GetFrequency() int
	SetInterval(int)
	GetInterval() int
	SetCount(int)
	GetCount() int
	SetDays(*string)
	GetDays() *string
	SetLastExecuted(time.Time)
	GetLastExecuted() time.Time
	SetExecutionCount(int)
	GetExecutionCount() int
	Hash() uint64
}

type TriggerConfig interface {
	GetChannel() int
	GetState() int
	GetTimer() string
	IsAsync() bool
}

type LicenseConfig interface {
	GetUserQuota() int
	GetFarmQuota() int
	GetControllerQuota() int
}

type UserConfig interface {
	GetID() int
	SetEmail(string)
	GetEmail() string
	SetPassword(string)
	GetPassword() string
	SetRoles(roles []Role)
	GetRoles() []Role
	AddRole(role Role)
	RedactPassword()
}

type RoleConfig interface {
	GetID() int
	GetName() string
}

/*
type FarmConfigChange struct {
	Type       string      `json:"type"`
	FarmConfig FarmConfig  `json:"farmConfig"`
	Payload    interface{} `json:"payload"`
}*/

type FarmConfigStorer interface {
	Len() int
	Put(farmID int, v FarmConfig) error
	Get(farmID int) (FarmConfig, error)
	GetAll() []FarmConfig
}
