package model

import (
	"github.com/jeremyhahn/go-cropdroid/common"
)

type User struct {
	ID                 uint64        `json:"id"`
	Email              string        `json:"email"`
	Password           string        `json:"password"`
	Roles              []common.Role `json:"roles"`
	common.UserAccount `json:"-"`
}

func NewUser() common.UserAccount {
	return &User{
		Roles: make([]common.Role, 0)}
}

func (user *User) GetID() uint64 {
	return user.ID
}

func (user *User) GetEmail() string {
	return user.Email
}

func (user *User) SetEmail(email string) {
	user.Email = email
}

func (user *User) GetPassword() string {
	return user.Password
}

func (user *User) SetPassword(password string) {
	user.Password = password
}

func (user *User) GetRoles() []common.Role {
	return user.Roles
}

func (user *User) SetRoles(roles []common.Role) {
	user.Roles = roles
}

func (user *User) AddRole(role common.Role) {
	user.Roles = append(user.Roles, role)
}

func (user *User) HasRole(name string) bool {
	for _, role := range user.Roles {
		if role.GetName() == name {
			return true
		}
	}
	return false
}
