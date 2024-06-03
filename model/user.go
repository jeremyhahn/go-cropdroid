package model

import "github.com/jeremyhahn/go-cropdroid/config"

type User interface {
	GetRoles() []Role
	SetRoles([]Role)
	AddRole(Role)
	config.CommonUser
}

type UserStruct struct {
	ID               uint64   `json:"id"`
	Email            string   `json:"email"`
	Password         string   `json:"password"`
	Roles            []Role   `json:"roles"`
	OrganizationRefs []uint64 `json:"-"`
	FarmRefs         []uint64 `json:"-"`
	User             `json:"-"`
}

func NewUser() User {
	return &UserStruct{
		Roles: make([]Role, 0)}
}

func (user *UserStruct) Identifier() uint64 {
	return user.ID
}

func (user *UserStruct) SetID(id uint64) {
	user.ID = id
}

func (user *UserStruct) GetEmail() string {
	return user.Email
}

func (user *UserStruct) SetEmail(email string) {
	user.Email = email
}

func (user *UserStruct) GetPassword() string {
	return user.Password
}

func (user *UserStruct) SetPassword(password string) {
	user.Password = password
}

func (user *UserStruct) GetRoles() []Role {
	return user.Roles
}

func (user *UserStruct) SetRoles(roles []Role) {
	user.Roles = roles
}

func (user *UserStruct) AddRole(role Role) {
	user.Roles = append(user.Roles, role)
}

func (user *UserStruct) HasRole(name string) bool {
	for _, role := range user.Roles {
		if role.GetName() == name {
			return true
		}
	}
	return false
}

func (user *UserStruct) SetOrganizationRefs(ids []uint64) {
	user.OrganizationRefs = ids
}

func (user *UserStruct) GetOrganizationRefs() []uint64 {
	return user.OrganizationRefs
}

func (user *UserStruct) SetFarmRefs(ids []uint64) {
	user.FarmRefs = ids
}

func (user *UserStruct) GetFarmRefs() []uint64 {
	return user.FarmRefs
}
