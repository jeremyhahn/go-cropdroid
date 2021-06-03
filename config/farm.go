package config

import (
	"fmt"
	"strconv"
	"time"
)

type Farm struct {
	ID             int    `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	OrganizationID int    `yaml:"orgId" json:"orgId"`
	Replicas       int    `yaml:"replicas" json:"replicas"`
	Mode           string `gorm:"-" yaml:"mode" json:"mode"`
	Name           string `gorm:"-" yaml:"name" json:"name"`
	Interval       int    `gorm:"-" yaml:"interval" json:"interval"`
	Smtp           *Smtp  `gorm:"-" yaml:"smtp" json:"smtp"`
	//Timezone       *time.Location `gorm:"-" yaml:"timezone" json:"timezone"`
	Timezone    string       `gorm:"-" yaml:"timezone" json:"timezone"`
	Controllers []Controller `yaml:"controllers" json:"controllers"`
	Users       []User       `gorm:"many2many:permissions" yaml:"users" json:"users"`
	FarmConfig  `yaml:"-" json:"-"`
}

func NewFarm() *Farm {
	return &Farm{
		Controllers: make([]Controller, 0),
		Users:       make([]User, 0)}
}

func CreateFarm(name string, orgID, interval int, users []User, controllers []Controller) FarmConfig {
	return &Farm{
		OrganizationID: orgID,
		Controllers:    controllers}
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

func (farm *Farm) GetSmtp() SmtpConfig {
	return farm.Smtp
}

func (farm *Farm) SetID(id int) {
	farm.ID = id
}

func (farm *Farm) GetID() int {
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

func (farm *Farm) AddController(controller Controller) {
	farm.Controllers = append(farm.Controllers, controller)
}

func (farm *Farm) GetControllers() []Controller {
	return farm.Controllers
}

func (farm *Farm) SetControllers(controllers []Controller) {
	farm.Controllers = controllers
}

func (farm *Farm) SetController(controller ControllerConfig) {
	for i, c := range farm.Controllers {
		if c.GetID() == controller.GetID() {
			farm.Controllers[i] = *controller.(*Controller)
			return
		}
	}
	farm.Controllers = append(farm.Controllers, *controller.(*Controller))
}

func (farm *Farm) GetController(controllerType string) (*Controller, error) {
	for _, controller := range farm.Controllers {
		if controller.GetType() == controllerType {
			return &controller, nil
		}
	}
	return nil, fmt.Errorf("[config.Farm] Controller type not found: %s", controllerType)
}

func (farm *Farm) GetControllerById(id int) (*Controller, error) {
	farmSize := len(farm.Controllers)
	if farmSize < id {
		return nil, fmt.Errorf("[config.Farm] Controller ID out of bounds: %d. Farm size: %d", id, farmSize)
	}
	return &farm.Controllers[id], nil
}

func (farm *Farm) ParseConfigs() error {
	for i, controller := range farm.GetControllers() {
		if controller.GetType() == "server" {
			smtpConfig := NewSmtp()
			for _, item := range controller.GetConfigs() {
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
		if err := controller.ParseConfigs(); err != nil {
			return err
		}
		farm.Controllers[i] = controller
	}
	//return fmt.Errorf("[config.Farm] Server configuration not found for farm. farm.id=$%d, farm.name: %s", farm.ID, farm.Name)
	return nil
}

// HydrateConfigs populates the controller config items from the ConfigMap. This is
// used when unmarshalling from JSON or YAML since controller.Configs json:"-" and yaml:"-"
// is set so the results are returned as key/value pairs by the API. Probably best to refactor
// this so the API returns a dedicated view and controller.Configs doesn't get ignored.
func (farm *Farm) HydrateConfigs() error {
	for i, controller := range farm.GetControllers() {
		if controller.GetType() == "server" {
			smtpConfig := NewSmtp()
			for key, value := range controller.GetConfigMap() {
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
		if err := controller.HydrateConfigs(); err != nil {
			return err
		}
		farm.Controllers[i] = controller
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
