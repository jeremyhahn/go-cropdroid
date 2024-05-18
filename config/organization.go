package config

import (
	"fmt"
)

// Organization groups users and devices
type Organization struct {
	ID   uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Name string `gorm:"size:255" yaml:"name" json:"name"`
	// Disabling gorm:"many2many:permissions on Farm so a permission
	// record doesn't get saved to the database without a role or user id
	//Farms []Farm `gorm:"many2many:permissions" yaml:"farms" json:"farms"`
	Farms []*Farm `yaml:"farms" json:"farms"`
	//Devices        []Device `yaml:"devices" json:"devices"`
	Users []*User `gorm:"foreignKey:ID" yaml:"users" json:"users"`
	//Users              []User   `yaml:"users" json:"users"`
	License        *License `yaml:"license" json:"license"`
	KeyValueEntity `gorm:"-" yaml:"-" json:"-"`
}

func NewOrganization() *Organization {
	return &Organization{
		Farms: make([]*Farm, 0),
		Users: make([]*User, 0)}
}

func CreateOrganization(farms []*Farm, users []*User) *Organization {
	return &Organization{
		Farms: farms,
		Users: users}
}

// SetID sets the unique identifier for the org
func (org *Organization) SetID(id uint64) {
	org.ID = id
}

// GetID returns the unique identifier for the org
func (org *Organization) Identifier() uint64 {
	return org.ID
}

// SetName sets the org name
func (org *Organization) SetName(name string) {
	org.Name = name
}

// GetName returns the org name
func (org *Organization) GetName() string {
	return org.Name
}

// SetFarms sets the farms that belong to the org
func (org *Organization) AddFarm(farm *Farm) {
	org.Farms = append(org.Farms, farm)
}

// SetFarms sets the farms that belong to the org
func (org *Organization) SetFarms(farms []*Farm) {
	org.Farms = farms
}

// GetFarm gets the farms that belong to the org
func (org *Organization) GetFarms() []*Farm {
	return org.Farms
}

// GetFarm returns the specified farm from the org
func (org *Organization) GetFarm(id uint64) (*Farm, error) {
	for _, farm := range org.Farms {
		if farm.ID == id {
			return farm, nil
		}
	}
	return nil, fmt.Errorf("[Organization.GetFarm] Farm not found with ID: %d", id)
}

func (org *Organization) AddUser(user *User) {
	org.Users = append(org.Users, user)
}

func (org *Organization) RemoveUser(user *User) {
	for i, u := range org.Users {
		if u.ID == user.ID {
			org.Users = append(org.Users[:i], org.Users[i+1:]...)
			break
		}
	}
}

func (org *Organization) SetUsers(users []*User) {
	org.Users = users
}

func (org *Organization) GetUsers() []*User {
	return org.Users
}

func (org *Organization) GetLicense() *License {
	return org.License
}

func (org *Organization) SetLicense(license *License) {
	org.License = license
}
