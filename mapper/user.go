package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type UserMapper interface {
	MapUserConfigToModel(config *config.UserStruct) model.User
	MapUserModelToConfig(model.User) *config.UserStruct
}

type DefaultUserMapper struct {
}

func NewUserMapper() UserMapper {
	return &DefaultUserMapper{}
}

func (mapper *DefaultUserMapper) MapUserConfigToModel(config *config.UserStruct) model.User {
	roles := make([]model.Role, len(config.GetRoles()))
	for i, role := range config.GetRoles() {
		roles[i] = &model.RoleStruct{
			ID:   role.ID,
			Name: role.GetName()}
	}
	return &model.UserStruct{
		ID:               config.Identifier(),
		Email:            config.GetEmail(),
		Password:         config.GetPassword(),
		OrganizationRefs: config.GetOrganizationRefs(),
		FarmRefs:         config.GetFarmRefs(),
		Roles:            roles}
}

func (mapper *DefaultUserMapper) MapUserModelToConfig(user model.User) *config.UserStruct {
	roles := make([]*config.RoleStruct, len(user.GetRoles()))
	for i, role := range user.GetRoles() {
		roles[i] = &config.RoleStruct{
			ID:   role.Identifier(),
			Name: role.GetName()}
	}
	return &config.UserStruct{
		ID:               user.Identifier(),
		Email:            user.GetEmail(),
		Password:         user.GetPassword(),
		OrganizationRefs: user.GetOrganizationRefs(),
		FarmRefs:         user.GetFarmRefs(),
		Roles:            roles}
}
