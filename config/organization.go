package config

import "fmt"

// Organization groups users and devices
type Organization struct {
	ID    uint64 `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Name  string `gorm:"size:255" yaml:"name" json:"name"`
	Farms []Farm `yaml:"farms" json:"farms"`
	//Devices        []Device `yaml:"devices" json:"devices"`
	Users []User `gorm:"many2many:permissions" yaml:"users" json:"users"`
	//Users              []User   `yaml:"users" json:"users"`
	License            *License `yaml:"license" json:"license"`
	OrganizationConfig `yaml:"-" json:"-"`
}

func NewOrganization() *Organization {
	return &Organization{
		Farms: make([]Farm, 0),
		Users: make([]User, 0)}
}

func CreateOrganization(farms []Farm, users []User) *Organization {
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
func (o *Organization) AddFarm(farm Farm) {
	o.Farms = append(o.Farms, farm)
}

// SetFarms sets the farms that belong to the org
func (o *Organization) SetFarms(farms []Farm) {
	o.Farms = farms
}

// GetFarm gets the farms that belong to the org
func (o *Organization) GetFarms() []Farm {
	return o.Farms
}

// GetFarm returns the specified farm from the org
func (o *Organization) GetFarm(id uint64) (*Farm, error) {
	for _, farm := range o.Farms {
		if farm.GetID() == id {
			return &farm, nil
		}
	}
	return nil, fmt.Errorf("[config.Organization.GetFarm] Farm not found with ID: %d", id)
}

func (o *Organization) SetUsers(users []User) {
	o.Users = users
}
func (o *Organization) GetUsers() []User {
	return o.Users
}

func (o *Organization) GetLicense() *License {
	return o.License
}

func (o *Organization) SetLicense(license *License) {
	o.License = license
}
