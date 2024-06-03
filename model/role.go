package model

import "github.com/jeremyhahn/go-cropdroid/config"

type Role interface {
	config.Role
}

type RoleStruct struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Role `json:"-"`
}

func NewRole() Role {
	return &RoleStruct{}
}

func (role *RoleStruct) Identifier() uint64 {
	return role.ID
}

func (role *RoleStruct) SetID(id uint64) {
	role.ID = id
}

func (role *RoleStruct) GetName() string {
	return role.Name
}

func (role *RoleStruct) SetName(name string) {
	role.Name = name
}
