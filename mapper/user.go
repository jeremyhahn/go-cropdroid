package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type UserMapper interface {
	MapUserConfigToModel(config *config.User) common.UserAccount
	MapUserModelToConfig(common.UserAccount) *config.User
}

type DefaultUserMapper struct {
}

func NewUserMapper() UserMapper {
	return &DefaultUserMapper{}
}

func (mapper *DefaultUserMapper) MapUserConfigToModel(config *config.User) common.UserAccount {
	roles := make([]common.Role, len(config.GetRoles()))
	for i, role := range config.GetRoles() {
		roles[i] = &model.Role{
			ID:   role.ID,
			Name: role.GetName()}
	}
	return &model.User{
		ID:               config.Identifier(),
		Email:            config.GetEmail(),
		Password:         config.GetPassword(),
		OrganizationRefs: config.GetOrganizationRefs(),
		FarmRefs:         config.GetFarmRefs(),
		Roles:            roles}
}

func (mapper *DefaultUserMapper) MapUserModelToConfig(user common.UserAccount) *config.User {
	roles := make([]*config.Role, len(user.GetRoles()))
	for i, role := range user.GetRoles() {
		roles[i] = &config.Role{
			ID:   role.GetID(),
			Name: role.GetName()}
	}
	return &config.User{
		ID:               user.GetID(),
		Email:            user.GetEmail(),
		Password:         user.GetPassword(),
		OrganizationRefs: user.GetOrganizationRefs(),
		FarmRefs:         user.GetFarmRefs(),
		Roles:            roles}
}
