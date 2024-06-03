package config

import (
	"fmt"
)

type CommonOrganization interface {
	GetName() string
	SetName(string)
	KeyValueEntity
}

type Organization interface {
	AddFarm(farm *FarmStruct)
	SetFarms(farms []*FarmStruct)
	GetFarms() []*FarmStruct
	GetFarm(id uint64) (*FarmStruct, error)
	AddUser(user *UserStruct)
	SetUsers(users []*UserStruct)
	GetUsers() []*UserStruct
	RemoveUser(user *UserStruct)
	GetLicense() *OrganizationLicenseStruct
	SetLicense(*OrganizationLicenseStruct)
	CommonOrganization
}

// *OrganizationStruct groups users and devices
type OrganizationStruct struct {
	ID   uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Name string `gorm:"size:255" yaml:"name" json:"name"`
	// Disabling gorm:"many2many:permissions on Farm so a permission
	// record doesn't get saved to the database without a role or user id
	//Farms []Farm `gorm:"many2many:permissions" yaml:"farms" json:"farms"`
	Farms []*FarmStruct `gorm:"foreignKey:OrganizationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"farms" json:"farms"`
	//Devices        []Device `yaml:"devices" json:"devices"`
	Users []*UserStruct `gorm:"many2many:organization_user" yaml:"users" json:"users"`
	//Users              []User   `yaml:"users" json:"users"`
	//License      *OrganizationLicenseStruct `yaml:"license" json:"license"`
	Organization `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewOrganization() *OrganizationStruct {
	return &OrganizationStruct{
		Farms: make([]*FarmStruct, 0),
		Users: make([]*UserStruct, 0)}
}

func CreateOrganization(farms []*FarmStruct, users []*UserStruct) *OrganizationStruct {
	return &OrganizationStruct{
		Farms: farms,
		Users: users}
}

func (org *OrganizationStruct) TableName() string {
	return "organizations"
}

// SetID sets the unique identifier for the org
func (org *OrganizationStruct) SetID(id uint64) {
	org.ID = id
}

// GetID returns the unique identifier for the org
func (org *OrganizationStruct) Identifier() uint64 {
	return org.ID
}

// SetName sets the org name
func (org *OrganizationStruct) SetName(name string) {
	org.Name = name
}

// GetName returns the org name
func (org *OrganizationStruct) GetName() string {
	return org.Name
}

// SetFarms sets the farms that belong to the org
func (org *OrganizationStruct) AddFarm(farm *FarmStruct) {
	org.Farms = append(org.Farms, farm)
}

// SetFarms sets the farms that belong to the org
func (org *OrganizationStruct) SetFarms(farms []*FarmStruct) {
	org.Farms = farms
}

// GetFarm gets the farms that belong to the org
func (org *OrganizationStruct) GetFarms() []*FarmStruct {
	return org.Farms
}

// GetFarm returns the specified farm from the org
func (org *OrganizationStruct) GetFarm(id uint64) (*FarmStruct, error) {
	for _, farm := range org.Farms {
		if farm.ID == id {
			return farm, nil
		}
	}
	return nil, fmt.Errorf("[*OrganizationStruct.GetFarm] Farm not found with ID: %d", id)
}

func (org *OrganizationStruct) AddUser(user *UserStruct) {
	org.Users = append(org.Users, user)
}

func (org *OrganizationStruct) RemoveUser(user *UserStruct) {
	for i, u := range org.Users {
		if u.ID == user.ID {
			org.Users = append(org.Users[:i], org.Users[i+1:]...)
			break
		}
	}
}

func (org *OrganizationStruct) SetUsers(users []*UserStruct) {
	org.Users = users
}

func (org *OrganizationStruct) GetUsers() []*UserStruct {
	return org.Users
}

// func (org *OrganizationStruct) GetLicense() *OrganizationLicenseStruct {
// 	return org.License
// }

// func (org *OrganizationStruct) SetLicense(license *OrganizationLicenseStruct) {
// 	org.License = license
// }
