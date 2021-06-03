package model

import "github.com/jeremyhahn/cropdroid/common"

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	common.Role `json:"-"`
}

func NewRole(name string) common.Role {
	return &Role{Name: name}
}

func (role *Role) GetID() int {
	return role.ID
}

func (role *Role) GetName() string {
	return role.Name
}
