package mapper

import (
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/model"
)

type UserMapper interface {
	MapUserEntityToModel(config config.UserConfig) common.UserAccount
	MapUserModelToEntity(common.UserAccount) config.UserConfig
}

type DefaultUserMapper struct {
}

func NewUserMapper() UserMapper {
	return &DefaultUserMapper{}
}

func (mapper *DefaultUserMapper) MapUserEntityToModel(config config.UserConfig) common.UserAccount {
	roles := make([]common.Role, len(config.GetRoles()))
	for i, role := range config.GetRoles() {
		roles[i] = &model.Role{
			ID:   role.GetID(),
			Name: role.GetName()}
	}
	return &model.User{
		ID:       config.GetID(),
		Email:    config.GetEmail(),
		Password: config.GetPassword(),
		Roles:    roles}
}

func (mapper *DefaultUserMapper) MapUserModelToEntity(user common.UserAccount) config.UserConfig {
	roles := make([]config.Role, len(user.GetRoles()))
	for i, role := range user.GetRoles() {
		roles[i] = config.Role{
			ID:   role.GetID(),
			Name: role.GetName()}
	}
	return &config.User{
		ID:       user.GetID(),
		Email:    user.GetEmail(),
		Password: user.GetPassword(),
		Roles:    roles}
}
