package config

import (
	"errors"
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

type KeyValueEntity interface {
	SetID(id uint64)
	Identifier() uint64
}

type TimeSeriesIndexeder interface {
	KeyValueEntity
	SetTimestamp(timestamp uint64)
	Timestamp() uint64
}

// type ServerConfig interface {
// 	SetID(id uint64)
// 	GetID() uint64
// 	SetOrganizationRefs(refs []uint64)
// 	GetOrganizationRefs() []uint64
// 	AddOrganizationRef(uint64)
// 	RemoveOrganizationRef(orgID uint64)
// 	SetFarmRefs(refs []uint64)
// 	GetFarmRefs() []uint64
// 	AddFarmRef(uint64)
// 	HasFarmRef(uint64) bool
// 	RemoveFarmRef(farmID uint64)
// }

// type OrganizationConfig interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	GetName() string
// 	SetName(string)
// 	AddFarm(farm FarmConfig)
// 	SetFarms(farms []FarmConfig)
// 	GetFarms() []FarmConfig
// 	GetFarm(id uint64) (FarmConfig, error)
// 	AddUser(user UserConfig)
// 	SetUsers(users []UserConfig)
// 	GetUsers() []UserConfig
// 	RemoveUser(user UserConfig)
// 	GetLicense() *License
// 	SetLicense(*License)
// }

// type FarmConfig interface {
// 	SetID(uint64)
// 	GetID() uint64
// 	SetOrganizationID(id uint64)
// 	GetOrganizationID() uint64
// 	GetReplicas() int
// 	SetReplicas(count int)
// 	SetConsistency(level int)
// 	GetConsistency() int
// 	SetStateStore(storeType int)
// 	GetStateStore() int
// 	SetConfigStore(storeType int)
// 	GetConfigStore() int
// 	SetDataStore(storeType int)
// 	GetDataStore() int
// 	SetName(string)
// 	GetName() string
// 	SetMode(string)
// 	GetMode() string
// 	GetInterval() int
// 	SetInterval(int)
// 	SetTimezone(tz string)
// 	GetTimezone() string
// 	SetPrivateKey(key string)
// 	GetPrivateKey() string
// 	SetPublicKey(key string)
// 	GetPublicKey() string
// 	GetSmtp() SmtpConfig
// 	AddUser(user UserConfig)
// 	SetUsers(users []UserConfig)
// 	GetUsers() []UserConfig
// 	RemoveUser(user UserConfig)
// 	AddDevice(DeviceConfig)
// 	GetDevices() []DeviceConfig
// 	SetDevices([]DeviceConfig)
// 	SetDevice(device DeviceConfig)
// 	GetDevice(deviceType string) (DeviceConfig, error)
// 	GetDeviceById(id uint64) (DeviceConfig, error)
// 	AddWorkflow(workflow WorkflowConfig)
// 	GetWorkflows() []WorkflowConfig
// 	RemoveWorkflow(workflow WorkflowConfig) error
// 	SetWorkflows(workflows []WorkflowConfig)
// 	SetWorkflow(workflow WorkflowConfig)
// 	ParseSettings() error
// 	HydrateSettings() error
// }

// type SmtpConfig interface {
// 	IsEnabled() bool
// 	SetEnable(enabled bool)
// 	SetHost(string)
// 	GetHost() string
// 	SetPort(int)
// 	GetPort() int
// 	SetUsername(string)
// 	GetUsername() string
// 	SetPassword(string)
// 	GetPassword() string
// 	SetRecipient(string)
// 	GetRecipient() string
// }

// type CommonDeviceConfig interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	GetFarmID() uint64
// 	SetFarmID(uint64)
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
// 	GetSettings() []DeviceSettingConfig
// 	GetSetting(key string) DeviceSettingConfig
// 	SetSettings([]DeviceSettingConfig)
// 	SetSetting(DeviceSettingConfig)
// 	GetConfigMap() map[string]string
// }

// type DeviceConfig interface {
// 	CommonDeviceConfig
// 	IsEnabled() bool
// 	SetEnabled(enabled bool)
// 	IsNotify() bool
// 	GetURI() string
// 	GetMetric(key string) (MetricConfig, error)
// 	GetMetrics() []MetricConfig
// 	SetMetrics([]MetricConfig)
// 	SetMetric(metric MetricConfig)
// 	GetChannel(id int) (ChannelConfig, error)
// 	GetChannels() []ChannelConfig
// 	SetChannel(channel ChannelConfig)
// 	SetChannels([]ChannelConfig)
// 	ParseSettings() error
// 	//HydrateConfigs() error
// }

// type DeviceSettingConfig interface {
// 	SetID(uint64)
// 	GetID() uint64
// 	SetUserID(uint64)
// 	GetUserID() uint64
// 	SetDeviceID(uint64)
// 	GetDeviceID() uint64
// 	SetKey(string)
// 	GetKey() string
// 	SetValue(string)
// 	GetValue() string
// }

// type MetricConfig interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	GetDeviceID() uint64
// 	SetDeviceID(uint64)
// 	GetDataType() int
// 	SetDataType(int)
// 	GetKey() string
// 	SetKey(string)
// 	GetName() string
// 	SetName(string)
// 	IsEnabled() bool
// 	SetEnable(bool)
// 	IsNotify() bool
// 	SetNotify(bool)
// 	GetUnit() string
// 	SetUnit(string)
// 	GetAlarmLow() float64
// 	SetAlarmLow(float64)
// 	GetAlarmHigh() float64
// 	SetAlarmHigh(float64)
// }

// type ChannelConfig interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	GetDeviceID() uint64
// 	SetDeviceID(uint64)
// 	GetChannelID() int
// 	SetChannelID(int)
// 	GetName() string
// 	SetName(name string)
// 	IsEnabled() bool
// 	SetEnable(bool)
// 	IsNotify() bool
// 	SetNotify(bool)
// 	AddCondition(condition ConditionConfig)
// 	GetConditions() []ConditionConfig
// 	SetConditions(conditions []ConditionConfig)
// 	SetCondition(condition ConditionConfig)
// 	GetSchedule() []ScheduleConfig
// 	SetSchedule(schedule []ScheduleConfig)
// 	SetScheduleItem(schedule ScheduleConfig)
// 	GetDuration() int
// 	SetDuration(int)
// 	GetDebounce() int
// 	SetDebounce(int)
// 	GetBackoff() int
// 	SetBackoff(int)
// 	GetAlgorithmID() uint64
// 	SetAlgorithmID(uint64)
// }

// type ConditionConfig interface {
// 	SetID(uint64)
// 	GetID() uint64
// 	SetWorkflowID(id uint64)
// 	GetWorkflowID() uint64
// 	SetChannelID(uint64)
// 	GetChannelID() uint64
// 	GetMetricID() uint64
// 	SetMetricID(uint64)
// 	SetComparator(string)
// 	GetComparator() string
// 	SetThreshold(float64)
// 	GetThreshold() float64
// 	Hash() uint64
// 	String() string
// }

// type ScheduleConfig interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	SetWorkflowID(id uint64)
// 	GetWorkflowID() uint64
// 	GetChannelID() uint64
// 	SetStartDate(time.Time)
// 	GetStartDate() time.Time
// 	SetEndDate(*time.Time)
// 	GetEndDate() *time.Time
// 	SetFrequency(int)
// 	GetFrequency() int
// 	SetInterval(int)
// 	GetInterval() int
// 	SetCount(int)
// 	GetCount() int
// 	SetDays(*string)
// 	GetDays() *string
// 	SetLastExecuted(time.Time)
// 	GetLastExecuted() time.Time
// 	SetExecutionCount(int)
// 	GetExecutionCount() int
// 	Hash() uint64
// 	String() string
// }

// type LicenseConfig interface {
// 	GetUserQuota() int
// 	GetFarmQuota() int
// 	GetDeviceQuota() int
// }

// type UserConfig interface {
// 	SetID(id uint64)
// 	GetID() uint64
// 	SetEmail(string)
// 	GetEmail() string
// 	SetPassword(string)
// 	GetPassword() string
// 	SetRoles(roles []RoleConfig)
// 	GetRoles() []RoleConfig
// 	AddRole(role RoleConfig)
// 	SetOrganizationRefs([]uint64)
// 	GetOrganizationRefs() []uint64
// 	AddOrganizationRef(id uint64)
// 	SetFarmRefs([]uint64)
// 	GetFarmRefs() []uint64
// 	AddFarmRef(id uint64)
// 	RedactPassword()
// }

// type WorkflowConfig interface {
// 	GetID() uint64
// 	SetID(id uint64)
// 	GetFarmID() uint64
// 	SetFarmID(id uint64)
// 	GetName() string
// 	SetName(name string)
// 	GetConditions() []Condition
// 	SetConditions(conditions []Condition)
// 	GetSchedules() []Schedule
// 	SetSchedules(schedules []Schedule)
// 	AddStep(step WorkflowStepConfig)
// 	GetSteps() []WorkflowStepConfig
// 	RemoveStep(step WorkflowStepConfig) error
// 	SetStep(step WorkflowStepConfig) error
// 	SetSteps(steps []WorkflowStepConfig)
// 	GetLastCompleted() *time.Time
// 	SetLastCompleted(t *time.Time)
// }

// type WorkflowStepConfig interface {
// 	GetID() uint64
// 	SetID(id uint64)
// 	// GetName() string
// 	// SetName(name string)
// 	GetWorkflowID() uint64
// 	SetWorkflowID(id uint64)
// 	GetDeviceID() uint64
// 	SetDeviceID(id uint64)
// 	GetChannelID() uint64
// 	SetChannelID(id uint64)
// 	GetWebhook() string
// 	SetWebhook(url string)
// 	GetDuration() int
// 	SetDuration(duration int)
// 	GetWait() int
// 	SetWait(seconds int)
// 	GetState() int
// 	SetState(state int)
// 	String() string
// }

// type RoleConfig interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	GetName() string
// }

// type RegistrationConfig interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	SetEmail(email string)
// 	GetEmail() string
// 	SetPassword(pw string)
// 	GetPassword() string
// 	RedactPassword()
// 	GetCreatedAt() int64
// 	SetOrganizationID(id uint64)
// 	GetOrganizationID() uint64
// 	SetOrganizationName(name string)
// 	GetOrganizationName() string
// }

// type PermissionConfig interface {
// 	// GetID() uint64
// 	// SetID(id uint64)
// 	GetOrgID() uint64
// 	SetOrgID(id uint64)
// 	GetFarmID() uint64
// 	SetFarmID(id uint64)
// 	GetUserID() uint64
// 	SetUserID(id uint64)
// 	GetRoleID() uint64
// 	SetRoleID(id uint64)
// }

// type AlgorithmConfig interface {
// 	GetID() uint64
// 	SetID(uint64)
// 	GetName() string
// 	SetName(string)
// }
