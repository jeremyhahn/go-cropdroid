package model

import (
	"github.com/jeremyhahn/cropdroid/common"
)

type User struct {
	ID                 int           `json:"id"`
	Email              string        `json:"email"`
	Password           string        `json:"password"`
	Roles              []common.Role `json:"roles"`
	common.UserAccount `json:"-"`
}

func NewUser() common.UserAccount {
	return &User{
		Roles: make([]common.Role, 0)}
}

func (user *User) GetID() int {
	return user.ID
}

func (user *User) GetEmail() string {
	return user.Email
}

func (user *User) GetPassword() string {
	return user.Password
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
