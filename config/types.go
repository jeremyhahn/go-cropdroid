package config

import (
	"errors"
	"time"
)

const (
	MEMORY_STORE = iota
	GORM_STORE
	RAFT_MEMORY_STORE
	RAFT_DISK_STORE
)

const (
	CONFIG_KEY_NAME     = "name"
	CONFIG_KEY_MODE     = "mode"
	CONFIG_KEY_INTERVAL = "interval"
	CONFIG_KEY_TIMEZONE = "timezone"
	CONFIG_KEY_ENABLE   = "enable"
	CONFIG_KEY_NOTIFY   = "notify"
	CONFIG_KEY_URI      = "uri"
)

var (
	ErrDeviceNotFound       = errors.New("device not found")
	ErrWorkflowNotFound     = errors.New("workflow not found")
	ErrWorkflowStepNotFound = errors.New("workflow step not found")
)

type ServerConfig interface {
	GetID() int
	SetID(int)
	GetOrganizations() []Organization
	SetOrganizations([]Organization)
	GetOrganization(id uint64) (*Organization, error)
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
	AddFarm(farm FarmConfig)
}

type OrganizationConfig interface {
	GetID() uint64
	SetID(uint64)
	GetName() string
	SetName(string)
	AddFarm(farm Farm)
	SetFarms(farms []Farm)
	GetFarms() []Farm
	GetFarm(id uint64) (*Farm, error)
	SetUsers(users []User)
	GetUsers() []User
	GetLicense() *License
	SetLicense(*License)
}

type FarmConfig interface {
	SetID(uint64)
	GetID() uint64
	SetOrganizationID(id uint64)
	GetOrganizationID() uint64
	GetReplicas() int
	SetReplicas(count int)
	SetConsistency(level int)
	GetConsistency() int
	SetStateStore(storeType int)
	GetStateStore() int
	SetConfigStore(storeType int)
	GetConfigStore() int
	SetDataStore(storeType int)
	GetDataStore() int
	SetName(string)
	GetName() string
	SetMode(string)
	GetMode() string
	// GetOrgID() int
	// SetOrgID(id int)
	GetInterval() int
	SetInterval(int)
	SetTimezone(tz string)
	GetTimezone() string
	SetPrivateKey(key string)
	GetPrivateKey() string
	SetPublicKey(key string)
	GetPublicKey() string
	GetSmtp() SmtpConfig
	SetUsers(users []User)
	GetUsers() []User
	AddDevice(Device)
	GetDevices() []Device
	SetDevices([]Device)
	SetDevice(device DeviceConfig)
	GetDevice(deviceType string) (*Device, error)
	GetDeviceById(id uint64) (*Device, error)
	AddWorkflow(workflow WorkflowConfig)
	GetWorkflows() []Workflow
	RemoveWorkflow(workflow WorkflowConfig) error
	SetWorkflows(workflows []Workflow)
	SetWorkflow(workflow WorkflowConfig)
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

type CommonDeviceConfig interface {
	GetID() uint64
	SetID(uint64)
	GetFarmID() uint64
	SetFarmID(uint64)
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
	GetConfigs() []DeviceConfigItem
	SetConfigs([]DeviceConfigItem)
	SetConfig(deviceConfig DeviceConfigConfig)
	GetConfigMap() map[string]string
}

type DeviceConfig interface {
	CommonDeviceConfig
	IsEnabled() bool
	SetEnabled(enabled bool)
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

type DeviceConfigConfig interface {
	SetID(uint64)
	GetID() uint64
	SetUserID(uint64)
	GetUserID() uint64
	SetDeviceID(uint64)
	GetDeviceID() uint64
	SetKey(string)
	GetKey() string
	SetValue(string)
	GetValue() string
}

type MetricConfig interface {
	GetID() int
	SetID(int)
	GetDeviceID() uint64
	SetDeviceID(uint64)
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
	GetID() uint64
	SetID(uint64)
	GetDeviceID() uint64
	SetDeviceID(uint64)
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
	SetWorkflowID(id uint64)
	GetWorkflowID() uint64
	SetChannelID(uint64)
	GetChannelID() uint64
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
	SetWorkflowID(id uint64)
	GetWorkflowID() uint64
	GetChannelID() uint64
	SetStartDate(time.Time)
	GetStartDate() time.Time
	SetEndDate(*time.Time)
	GetEndDate() *time.Time
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

type LicenseConfig interface {
	GetUserQuota() int
	GetFarmQuota() int
	GetDeviceQuota() int
}

type UserConfig interface {
	GetID() uint64
	SetEmail(string)
	GetEmail() string
	SetPassword(string)
	GetPassword() string
	SetRoles(roles []Role)
	GetRoles() []Role
	AddRole(role Role)
	RedactPassword()
}

type WorkflowConfig interface {
	GetID() uint64
	SetID(id uint64)
	GetFarmID() uint64
	SetFarmID(id uint64)
	GetName() string
	SetName(name string)
	GetConditions() []Condition
	SetConditions(conditions []Condition)
	GetSchedules() []Schedule
	SetSchedules(schedules []Schedule)
	AddStep(step WorkflowStepConfig)
	GetSteps() []WorkflowStep
	RemoveStep(step WorkflowStepConfig) error
	SetStep(step WorkflowStepConfig) error
	SetSteps(steps []WorkflowStep)
	GetLastCompleted() *time.Time
	SetLastCompleted(t *time.Time)
}

type WorkflowStepConfig interface {
	GetID() uint64
	SetID(id uint64)
	// GetName() string
	// SetName(name string)
	GetWorkflowID() uint64
	SetWorkflowID(id uint64)
	GetDeviceID() uint64
	SetDeviceID(id uint64)
	GetChannelID() uint64
	SetChannelID(id uint64)
	GetWebhook() string
	SetWebhook(url string)
	GetDuration() int
	SetDuration(duration int)
	GetWait() int
	SetWait(seconds int)
	GetState() int
	SetState(state int)
}

type RoleConfig interface {
	GetID() uint64
	GetName() string
}

type RegistrationConfig interface {
	GetID() uint64
	SetEmail(email string)
	GetEmail() string
	SetPassword(pw string)
	GetPassword() string
	RedactPassword()
	GetCreatedAt() int64
	SetOrganizationID(id uint64)
	GetOrganizationID() uint64
	SetOrganizationName(name string)
	GetOrganizationName() string
}

type PermissionConfig interface {
	GetID() uint64
	SetID(id uint64)
	GetOrgID() uint64
	SetOrgID(id uint64)
	GetFarmID() uint64
	SetFarmID(id uint64)
	GetUserID() uint64
	SetUserID(id uint64)
	GetRoleID() uint64
	SetRoleID(id uint64)
}
