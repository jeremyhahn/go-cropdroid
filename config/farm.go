package config

import (
	"fmt"
	"strconv"
	"time"
)

type CommonFarm interface {
	SetOrganizationID(id uint64)
	GetOrganizationID() uint64
	GetReplicas() int
	SetReplicas(count int)
	SetConsistencyLevel(level int)
	GetConsistencyLevel() int
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
	GetInterval() int
	SetInterval(int)
	SetSmtp(smtp *SmtpStruct)
	GetSmtp() *SmtpStruct
	SetTimezone(tz string)
	GetTimezone() string
	SetPrivateKey(key string)
	GetPrivateKey() string
	SetPublicKey(key string)
	GetPublicKey() string
	AddDevice(*DeviceStruct)
	GetDevices() []*DeviceStruct
	SetDevices([]*DeviceStruct)
	SetDevice(device *DeviceStruct)
	GetDevice(deviceType string) (*DeviceStruct, error)
	GetDeviceById(id uint64) (*DeviceStruct, error)
	AddUser(user *UserStruct)
	SetUsers(users []*UserStruct)
	GetUsers() []*UserStruct
	RemoveUser(user *UserStruct)
	AddWorkflow(workflow *WorkflowStruct)
	GetWorkflows() []*WorkflowStruct
	RemoveWorkflow(workflow *WorkflowStruct) error
	SetWorkflows(workflows []*WorkflowStruct)
	SetWorkflow(workflow *WorkflowStruct)
	KeyValueEntity
}

type Farm interface {
	ParseSettings() error
	HydrateSettings() error
	CommonFarm
}

type FarmStruct struct {
	ID             uint64            `gorm:"primaryKey" yaml:"id" json:"id"`
	OrganizationID uint64            `yaml:"orgId" json:"orgId"`
	Replicas       int               `yaml:"replicas" json:"replicas"`
	Consistency    int               `gorm:"consistency" yaml:"consistency" json:"consistency"`
	StateStore     int               `gorm:"state_store" yaml:"state_store" json:"state_store"`
	ConfigStore    int               `gorm:"config_store" yaml:"config_store" json:"config_store"`
	DataStore      int               `gorm:"data_store" yaml:"data_store" json:"data_store"`
	Mode           string            `gorm:"-" yaml:"mode" json:"mode"`
	Name           string            `gorm:"-" yaml:"name" json:"name"`
	Interval       int               `gorm:"-" yaml:"interval" json:"interval"`
	Smtp           *SmtpStruct       `gorm:"-" yaml:"smtp" json:"smtp"`
	Timezone       string            `gorm:"-" yaml:"timezone" json:"timezone"`
	PrivateKey     string            `gorm:"private_key" yaml:"private_key" json:"private_key"`
	PublicKey      string            `gorm:"public_key" yaml:"public_key" json:"public_key"`
	Devices        []*DeviceStruct   `gorm:"foreignKey:FarmID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"devices" json:"devices"`
	Users          []*UserStruct     `gorm:"many2many:user_farm" yaml:"users" json:"users"`
	Workflows      []*WorkflowStruct `gorm:"name:workflow;foreignKey:FarmID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"workflows" json:"workflows"`
	Farm           `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewFarm() *FarmStruct {
	return &FarmStruct{
		//Interval: 60,
		Devices:   make([]*DeviceStruct, 0),
		Users:     make([]*UserStruct, 0),
		Workflows: make([]*WorkflowStruct, 0)}
}

func CreateFarm(name string, orgID uint64, interval int,
	users []*User, devices []*DeviceStruct) *FarmStruct {

	return &FarmStruct{
		//Interval:       60,
		OrganizationID: orgID,
		Devices:        devices,
		Users:          make([]*UserStruct, 0),
		Workflows:      make([]*WorkflowStruct, 0)}
}

func (farm *FarmStruct) TableName() string {
	return "farms"
}

func (farm *FarmStruct) SetID(id uint64) {
	farm.ID = id
}

func (farm *FarmStruct) Identifier() uint64 {
	return farm.ID
}

func (farm *FarmStruct) SetOrganizationID(id uint64) {
	farm.OrganizationID = id
}

func (farm *FarmStruct) GetOrganizationID() uint64 {
	return farm.OrganizationID
}

func (farm *FarmStruct) SetReplicas(count int) {
	farm.Replicas = count
}

func (farm *FarmStruct) GetReplicas() int {
	return farm.Replicas
}

func (farm *FarmStruct) SetConsistencyLevel(level int) {
	farm.Consistency = level
}

func (farm *FarmStruct) GetConsistencyLevel() int {
	return farm.Consistency
}

func (farm *FarmStruct) SetStateStore(storeType int) {
	farm.StateStore = storeType
}

func (farm *FarmStruct) GetStateStore() int {
	return farm.StateStore
}

func (farm *FarmStruct) SetConfigStore(storeType int) {
	farm.ConfigStore = storeType
}

func (farm *FarmStruct) GetConfigStore() int {
	return farm.ConfigStore
}

func (farm *FarmStruct) SetDataStore(storeType int) {
	farm.DataStore = storeType
}

func (farm *FarmStruct) GetDataStore() int {
	return farm.DataStore
}

func (farm *FarmStruct) SetName(name string) {
	farm.Name = name
}

func (farm *FarmStruct) GetName() string {
	return farm.Name
}

func (farm *FarmStruct) GetMode() string {
	return farm.Mode
}

func (farm *FarmStruct) SetMode(mode string) {
	farm.Mode = mode
}

func (farm *FarmStruct) GetInterval() int {
	return farm.Interval
}

func (farm *FarmStruct) SetInterval(interval int) {
	farm.Interval = interval
}

/*func (farm *FarmStruct) SetTimezone(tz *time.Location) {
	farm.Timezone = tz
}

func (farm *FarmStruct) GetTimezone() *time.Location {
	return farm.Timezone
}*/

func (farm *FarmStruct) SetTimezone(tz string) {
	farm.Timezone = tz
}

func (farm *FarmStruct) GetTimezone() string {
	return farm.Timezone
}

func (farm *FarmStruct) SetPrivateKey(key string) {
	farm.PrivateKey = key
}

func (farm *FarmStruct) GetPrivateKey() string {
	return farm.PrivateKey
}

func (farm *FarmStruct) SetPublicKey(key string) {
	farm.PublicKey = key
}

func (farm *FarmStruct) GetPublicKey() string {
	return farm.PublicKey
}

func (farm *FarmStruct) SetSmtp(smtp *SmtpStruct) {
	farm.Smtp = smtp
}

func (farm *FarmStruct) GetSmtp() *SmtpStruct {
	return farm.Smtp
}

func (farm *FarmStruct) AddUser(user *UserStruct) {
	farm.Users = append(farm.Users, user)
}

func (farm *FarmStruct) RemoveUser(user *UserStruct) {
	for i, u := range farm.Users {
		if u.ID == user.ID {
			farm.Users = append(farm.Users[:i], farm.Users[i+1:]...)
			break
		}
	}
}

func (farm *FarmStruct) SetUsers(users []*UserStruct) {
	farm.Users = users
}

func (farm *FarmStruct) GetUsers() []*UserStruct {
	return farm.Users
}

func (farm *FarmStruct) AddDevice(device *DeviceStruct) {
	farm.Devices = append(farm.Devices, device)
}

func (farm *FarmStruct) GetDevices() []*DeviceStruct {
	return farm.Devices
}

func (farm *FarmStruct) SetDevices(devices []*DeviceStruct) {
	farm.Devices = devices
}

func (farm *FarmStruct) SetDevice(device *DeviceStruct) {
	for i, c := range farm.Devices {
		if c.ID == device.ID {
			farm.Devices[i] = device
			return
		}
	}
	farm.Devices = append(farm.Devices, device)
}

func (farm *FarmStruct) GetDevice(deviceType string) (*DeviceStruct, error) {
	for _, device := range farm.Devices {
		if device.GetType() == deviceType {
			return device, nil
		}
	}
	return nil, fmt.Errorf("[config.Farm] Device type not found: %s",
		deviceType)
}

func (farm *FarmStruct) GetDeviceById(id uint64) (*DeviceStruct, error) {
	for _, device := range farm.Devices {
		if device.ID == id {
			return device, nil
		}
	}
	return nil, fmt.Errorf("device not found: %d", id)
}

// func (farm *FarmStruct) GetDeviceByType(t string) (*Device, error) {
// 	for _, device := range farm.Devices {
// 		if device.GetType() == t {
// 			return &device, nil
// 		}
// 	}
// 	return nil, ErrDeviceNotFound
// }

func (farm *FarmStruct) SetWorkflows(workflows []*WorkflowStruct) {
	farm.Workflows = workflows
}

func (farm *FarmStruct) GetWorkflows() []*WorkflowStruct {
	return farm.Workflows
}

func (farm *FarmStruct) AddWorkflow(workflow *WorkflowStruct) {
	farm.Workflows = append(farm.Workflows, workflow)
}

func (farm *FarmStruct) SetWorkflow(workflow *WorkflowStruct) {
	for i, w := range farm.Workflows {
		if w.ID == workflow.ID {
			farm.Workflows[i] = workflow
			return
		}
	}
	farm.Workflows = append(farm.Workflows, workflow)
}

func (farm *FarmStruct) RemoveWorkflow(workflow *WorkflowStruct) error {
	for i, w := range farm.Workflows {
		if w.ID == workflow.ID {
			farm.Workflows = append(farm.Workflows[:i], farm.Workflows[i+1:]...)
			return nil
		}
	}
	return ErrWorkflowNotFound
}

func (farm *FarmStruct) ParseSettings() error {
	for i, device := range farm.GetDevices() {
		if device.GetType() == "server" {
			smtp := NewSmtp()
			for _, item := range device.GetSettings() {
				key := item.GetKey()
				value := item.GetValue()
				switch key {
				case "name":
					farm.Name = value
				case "interval":
					interval, err := strconv.Atoi(value)
					if err != nil {
						return err
					}
					farm.Interval = int(interval)
				case "timezone":
					location, err := time.LoadLocation(value)
					if err != nil {
						return err
					}
					//farm.Timezone = location
					farm.Timezone = location.String()
				case "mode":
					farm.Mode = value
				case "smtp.enable":
					bEnable, err := strconv.ParseBool(value)
					if err != nil {
						return err
					}
					smtp.SetEnable(bEnable)
				case "smtp.host":
					smtp.SetHost(value)
				case "smtp.port":
					smtpPortInt, err := strconv.ParseInt(value, 10, 0)
					if err != nil {
						return err
					}
					smtp.SetPort(int(smtpPortInt))
				case "smtp.username":
					smtp.SetUsername(value)
				case "smtp.password":
					smtp.SetPassword(value)
				case "smtp.recipient":
					smtp.SetRecipient(value)
				}
			}
			farm.Smtp = smtp
		}
		if err := device.ParseSettings(); err != nil {
			return err
		}
		farm.Devices[i] = device
	}
	//return fmt.Errorf("[config.Farm] Server configuration not found for farm. farm.id=$%d, farm.name: %s", farm.ID, farm.Name)
	return nil
}

// HydrateConfigs populates the device config items from the SettingsMap. This is
// used when unmarshalling from JSON or YAML since device.Configs json:"-" and yaml:"-"
// is set so the results are returned as key/value pairs by the API. Probably best to refactor
// this so the API returns a dedicated view and device.Configs doesn't get ignored.
func (farm *FarmStruct) HydrateSettings() error {
	for i, device := range farm.GetDevices() {
		if device.GetType() == "server" {
			smtp := NewSmtp()
			for key, value := range device.GetSettingsMap() {
				switch key {
				case "name":
					farm.Name = value
				case "interval":
					interval, err := strconv.Atoi(value)
					if err != nil {
						return err
					}
					farm.Interval = int(interval)
				case "timezone":
					location, err := time.LoadLocation(value)
					if err != nil {
						return err
					}
					//farm.Timezone = location
					farm.Timezone = location.String()
				case "mode":
					farm.Mode = value
				case "smtp.enable":
					bEnable, err := strconv.ParseBool(value)
					if err != nil {
						return err
					}
					smtp.SetEnable(bEnable)
				case "smtp.host":
					smtp.SetHost(value)
				case "smtp.port":
					smtpPortInt, err := strconv.ParseInt(value, 10, 0)
					if err != nil {
						return err
					}
					smtp.SetPort(int(smtpPortInt))
				case "smtp.username":
					smtp.SetUsername(value)
				case "smtp.password":
					smtp.SetPassword(value)
				case "smtp.recipient":
					smtp.SetRecipient(value)
				}
			}
			farm.Smtp = smtp
		}
		// if err := device.HydrateConfigs(); err != nil {
		// 	return err
		// }
		farm.Devices[i] = device
	}
	return nil
}

/*
func (farm *FarmStruct) getStringConfig(key string) (string, error) {
	if farm.Configs != nil {
		for _, config := range farm.Configs {
			if config.GetKey() == key {
				return config.GetValue(), nil
			}
		}
		return "", fmt.Errorf("[config.Farm] Config not found: %s", key)
	}
	return "", errors.New("[config.Farm] Configuration undefined")
}

func (farm *FarmStruct) getIntConfig(key string) (int, error) {
	if farm.Configs != nil {
		for _, config := range farm.Configs {
			if config.GetKey() == key {
				intConfig, err := strconv.Atoi(config.GetValue())
				if err != nil {
					return 0, fmt.Errorf("Invalid farm config (integer): %s", key)
				}
				return intConfig, nil
			}
		}
		return 0, fmt.Errorf("[config.Farm] Config not found: %s", key)
	}
	return 0, errors.New("[config.Farm] Configuration undefined")
}

func (farm *FarmStruct) setConfig(key, value string) error {
	for _, config := range farm.Configs {
		if config.GetKey() == key {
			config.SetValue(value)
			return nil
		}
	}
	return fmt.Errorf("[config.Farm] Farm config key not found: %s", key)
}
*/
