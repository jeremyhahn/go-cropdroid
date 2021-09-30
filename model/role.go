package model

import "github.com/jeremyhahn/go-cropdroid/common"

type Role struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	common.Role `json:"-"`
}

func NewRole(name string) common.Role {
	return &Role{Name: name}
}

func (role *Role) GetID() uint64 {
	return role.ID
}

func (role *Role) GetName() string {
	return role.Name
}
