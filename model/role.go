package model

import "github.com/jeremyhahn/go-cropdroid/common"

type Role struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	common.Role `json:"-"`
}

func NewRole() common.Role {
	return &Role{}
}

func (role *Role) GetID() uint64 {
	return role.ID
}

func (role *Role) GetName() string {
	return role.Name
}

func (role *Role) SetName(name string) {
	role.Name = name
}
