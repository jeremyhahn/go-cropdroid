package config

import (
	"fmt"
	"strconv"
	"time"
)

type Farm struct {
	ID             uint64     `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	OrganizationID int        `yaml:"orgId" json:"orgId"`
	Replicas       int        `yaml:"replicas" json:"replicas"`
	Consistency    int        `gorm:"consistency" yaml:"consistency" json:"consistency"`
	StateStore     int        `gorm:"state_store" yaml:"state_store" json:"state_store"`
	ConfigStore    int        `gorm:"config_store" yaml:"config_store" json:"config_store"`
	Mode           string     `gorm:"-" yaml:"mode" json:"mode"`
	Name           string     `gorm:"-" yaml:"name" json:"name"`
	Interval       int        `gorm:"-" yaml:"interval" json:"interval"`
	Smtp           *Smtp      `gorm:"-" yaml:"smtp" json:"smtp"`
	Timezone       string     `gorm:"-" yaml:"timezone" json:"timezone"`
	PrivateKey     string     `gorm:"private_key" yaml:"private_key" json:"private_key"`
	PublicKey      string     `gorm:"public_key" yaml:"public_key" json:"public_key"`
	Devices        []Device   `yaml:"devices" json:"devices"`
	Users          []User     `gorm:"many2many:permissions" yaml:"users" json:"users"`
	Workflows      []Workflow `gorm:"workflow" yaml:"workflows" json:"workflows"`
	FarmConfig     `yaml:"-" json:"-"`
}

func NewFarm() *Farm {
	return &Farm{
		//Interval: 60,
		Devices: make([]Device, 0),
		Users:   make([]User, 0)}
}

func CreateFarm(name string, orgID, interval int, users []User,
	devices []Device) FarmConfig {

	return &Farm{
		//Interval:       60,
		OrganizationID: orgID,
		Devices:        devices}
}

func (farm *Farm) SetOrganizationID(id int) {
	farm.OrganizationID = id
}

func (farm *Farm) GetOrganizationID() int {
	return farm.OrganizationID
}

func (farm *Farm) SetReplicas(count int) {
	farm.Replicas = count
}

func (farm *Farm) GetReplicas() int {
	return farm.Replicas
}

func (farm *Farm) SetConsistency(level int) {
	farm.Consistency = level
}

func (farm *Farm) GetConsistency() int {
	return farm.Consistency
}

func (farm *Farm) SetStateStore(storeType int) {
	farm.StateStore = storeType
}

func (farm *Farm) GetStateStore() int {
	return farm.StateStore
}

func (farm *Farm) SetConfigStore(storeType int) {
	farm.ConfigStore = storeType
}

func (farm *Farm) GetConfigStore() int {
	return farm.ConfigStore
}

func (farm *Farm) SetName(name string) {
	farm.Name = name
}

func (farm *Farm) GetName() string {
	return farm.Name
}

func (farm *Farm) GetMode() string {
	return farm.Mode
}

func (farm *Farm) SetMode(mode string) {
	farm.Mode = mode
}

func (farm *Farm) GetInterval() int {
	return farm.Interval
}

func (farm *Farm) SetInterval(interval int) {
	farm.Interval = interval
}

/*func (farm *Farm) SetTimezone(tz *time.Location) {
	farm.Timezone = tz
}

func (farm *Farm) GetTimezone() *time.Location {
	return farm.Timezone
}*/

func (farm *Farm) SetTimezone(tz string) {
	farm.Timezone = tz
}

func (farm *Farm) GetTimezone() string {
	return farm.Timezone
}

func (farm *Farm) SetPrivateKey(key string) {
	farm.PrivateKey = key
}

func (farm *Farm) GetPrivateKey() string {
	return farm.PrivateKey
}

func (farm *Farm) SetPublicKey(key string) {
	farm.PublicKey = key
}

func (farm *Farm) GetPublicKey() string {
	return farm.PublicKey
}

func (farm *Farm) GetSmtp() SmtpConfig {
	return farm.Smtp
}

func (farm *Farm) SetID(id uint64) {
	farm.ID = id
}

func (farm *Farm) GetID() uint64 {
	return farm.ID
}

func (farm *Farm) SetOrgID(id int) {
	farm.OrganizationID = id
}

func (farm *Farm) GetOrgID() int {
	return farm.OrganizationID
}

func (farm *Farm) SetUsers(users []User) {
	farm.Users = users
}

func (farm *Farm) GetUsers() []User {
	return farm.Users
}

func (farm *Farm) AddDevice(device Device) {
	farm.Devices = append(farm.Devices, device)
}

func (farm *Farm) GetDevices() []Device {
	return farm.Devices
}

func (farm *Farm) SetDevices(devices []Device) {
	farm.Devices = devices
}

func (farm *Farm) SetDevice(device DeviceConfig) {
	for i, c := range farm.Devices {
		if c.GetID() == device.GetID() {
			farm.Devices[i] = *device.(*Device)
			return
		}
	}
	farm.Devices = append(farm.Devices, *device.(*Device))
}

func (farm *Farm) GetDevice(deviceType string) (*Device, error) {
	for _, device := range farm.Devices {
		if device.GetType() == deviceType {
			return &device, nil
		}
	}
	return nil, fmt.Errorf("[config.Farm] Device type not found: %s", deviceType)
}

func (farm *Farm) GetDeviceById(id uint64) (*Device, error) {
	for _, device := range farm.Devices {
		if device.GetID() == id {
			return &device, nil
		}
	}
	return nil, fmt.Errorf("device not found: %d", id)
}

// func (farm *Farm) GetDeviceByType(t string) (*Device, error) {
// 	for _, device := range farm.Devices {
// 		if device.GetType() == t {
// 			return &device, nil
// 		}
// 	}
// 	return nil, ErrDeviceNotFound
// }

func (farm *Farm) SetWorkflows(workflows []Workflow) {
	farm.Workflows = workflows
}

func (farm *Farm) GetWorkflows() []Workflow {
	return farm.Workflows
}

func (farm *Farm) AddWorkflow(workflow WorkflowConfig) {
	farm.Workflows = append(farm.Workflows, *workflow.(*Workflow))
}

func (farm *Farm) SetWorkflow(workflow WorkflowConfig) {
	for i, w := range farm.Workflows {
		if w.GetID() == workflow.GetID() {
			farm.Workflows[i] = *workflow.(*Workflow)
			return
		}
	}
	farm.Workflows = append(farm.Workflows, *workflow.(*Workflow))
}

func (farm *Farm) RemoveWorkflow(workflow WorkflowConfig) error {
	for i, w := range farm.Workflows {
		if w.GetID() == workflow.GetID() {
			farm.Workflows = append(farm.Workflows[:i], farm.Workflows[i+1:]...)
			return nil
		}
	}
	return ErrWorkflowNotFound
}

func (farm *Farm) ParseConfigs() error {
	for i, device := range farm.GetDevices() {
		if device.GetType() == "server" {
			smtpConfig := NewSmtp()
			for _, item := range device.GetConfigs() {
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
					smtpConfig.SetEnable(bEnable)
				case "smtp.host":
					smtpConfig.SetHost(value)
				case "smtp.port":
					smtpPortInt, err := strconv.ParseInt(value, 10, 0)
					if err != nil {
						return err
					}
					smtpConfig.SetPort(int(smtpPortInt))
				case "smtp.username":
					smtpConfig.SetUsername(value)
				case "smtp.password":
					smtpConfig.SetPassword(value)
				case "smtp.recipient":
					smtpConfig.SetRecipient(value)
				}
			}
			farm.Smtp = smtpConfig.(*Smtp)
		}
		if err := device.ParseConfigs(); err != nil {
			return err
		}
		farm.Devices[i] = device
	}
	//return fmt.Errorf("[config.Farm] Server configuration not found for farm. farm.id=$%d, farm.name: %s", farm.ID, farm.Name)
	return nil
}

// HydrateConfigs populates the device config items from the ConfigMap. This is
// used when unmarshalling from JSON or YAML since device.Configs json:"-" and yaml:"-"
// is set so the results are returned as key/value pairs by the API. Probably best to refactor
// this so the API returns a dedicated view and device.Configs doesn't get ignored.
func (farm *Farm) HydrateConfigs() error {
	for i, device := range farm.GetDevices() {
		if device.GetType() == "server" {
			smtpConfig := NewSmtp()
			for key, value := range device.GetConfigMap() {
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
					smtpConfig.SetEnable(bEnable)
				case "smtp.host":
					smtpConfig.SetHost(value)
				case "smtp.port":
					smtpPortInt, err := strconv.ParseInt(value, 10, 0)
					if err != nil {
						return err
					}
					smtpConfig.SetPort(int(smtpPortInt))
				case "smtp.username":
					smtpConfig.SetUsername(value)
				case "smtp.password":
					smtpConfig.SetPassword(value)
				case "smtp.recipient":
					smtpConfig.SetRecipient(value)
				}
			}
			farm.Smtp = smtpConfig.(*Smtp)
		}
		if err := device.HydrateConfigs(); err != nil {
			return err
		}
		farm.Devices[i] = device
	}
	return nil
}

/*
func (farm *Farm) getStringConfig(key string) (string, error) {
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

func (farm *Farm) getIntConfig(key string) (int, error) {
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

func (farm *Farm) setConfig(key, value string) error {
	for _, config := range farm.Configs {
		if config.GetKey() == key {
			config.SetValue(value)
			return nil
		}
	}
	return fmt.Errorf("[config.Farm] Farm config key not found: %s", key)
}
*/
