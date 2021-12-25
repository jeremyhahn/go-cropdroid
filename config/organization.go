package config

import (
	"fmt"
)

// Organization groups users and devices
type Organization struct {
	ID    uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Name  string `gorm:"size:255" yaml:"name" json:"name"`
	Farms []Farm `gorm:"many2many:permissions" yaml:"farms" json:"farms"`
	//Devices        []Device `yaml:"devices" json:"devices"`
	Users []User `gorm:"many2many:permissions" yaml:"users" json:"users"`
	//Users              []User   `yaml:"users" json:"users"`
	License            *License `yaml:"license" json:"license"`
	OrganizationConfig `yaml:"-" json:"-"`
}

func NewOrganization() OrganizationConfig {
	return &Organization{
		Farms: make([]Farm, 0),
		Users: make([]User, 0)}
}

func CreateOrganization(farms []Farm, users []User) OrganizationConfig {
	return &Organization{
		Farms: farms,
		Users: users}
}

// SetID sets the unique identifier for the org
func (o *Organization) SetID(id uint64) {
	o.ID = id
}

// GetID returns the unique identifier for the org
func (o *Organization) GetID() uint64 {
	return o.ID
}

// SetName sets the org name
func (o *Organization) SetName(name string) {
	o.Name = name
}

// GetName returns the org name
func (o *Organization) GetName() string {
	return o.Name
}

// SetFarms sets the farms that belong to the org
func (o *Organization) AddFarm(farm FarmConfig) {
	o.Farms = append(o.Farms, *farm.(*Farm))
}

// SetFarms sets the farms that belong to the org
func (o *Organization) SetFarms(farms []FarmConfig) {
	farmStructs := make([]Farm, len(farms))
	for i, farm := range farms {
		farmStructs[i] = *farm.(*Farm)
	}
	o.Farms = farmStructs
}

// GetFarm gets the farms that belong to the org
func (o *Organization) GetFarms() []FarmConfig {
	farmConfigs := make([]FarmConfig, len(o.Farms))
	for i, farm := range o.Farms {
		farmConfigs[i] = &farm
	}
	return farmConfigs
}

// GetFarm returns the specified farm from the org
func (o *Organization) GetFarm(id uint64) (FarmConfig, error) {
	for _, farm := range o.Farms {
		if farm.GetID() == id {
			return &farm, nil
		}
	}
	return nil, fmt.Errorf("[Organization.GetFarm] Farm not found with ID: %d", id)
}

func (o *Organization) SetUsers(users []UserConfig) {
	userStructs := make([]User, len(users))
	for i, user := range users {
		userStructs[i] = *user.(*User)
	}
	o.Users = userStructs
}

func (o *Organization) GetUsers() []UserConfig {
	userConfigs := make([]UserConfig, len(o.Users))
	for i, user := range o.Users {
		userConfigs[i] = &user
	}
	return userConfigs
}

func (o *Organization) GetLicense() *License {
	return o.License
}

func (o *Organization) SetLicense(license *License) {
	o.License = license
}
