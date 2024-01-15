//go:build ignore
// +build ignore

package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/config"
)

type ConfigMapper interface {
	MapFromFileConfig(serverConfig config.Server) (config.ServerConfig, error)
}

type DefaultConfigMapper struct {
	deviceMapper DeviceMapper
}

func NewConfigMapper() ConfigMapper {
	return &DefaultConfigMapper{}
}

func (mapper *DefaultConfigMapper) MapFromFileConfig(yamlConfig config.Server) (config.ServerConfig, error) {

	license := &config.License{
		UserQuota:   1,
		FarmQuota:   1,
		DeviceQuota: 3}

	_orgs := make([]config.Organization, len(yamlConfig.Organizations))
	for i, org := range yamlConfig.Organizations {

		_farms := make([]config.Farm, len(org.Farms))
		for j, farm := range org.Farms {

			farmUsers := make([]config.User, len(farm.Users))
			for k, user := range farm.Users {

				_roles := make([]config.Role, len(user.Roles))
				for l, role := range user.Roles {
					_roles[l] = config.Role{
						ID:   uint64(l),
						Name: role.Name}
				}
				farmUsers[k] = config.User{
					ID:       user.ID,
					Email:    user.Email,
					Password: user.Password,
					Roles:    _roles}
			}
			_farms[j] = config.Farm{
				ID: farm.ID,
				//OrganizationID: farm.OrganizationID,
				Users: farmUsers}
		}

		orgUsers := make([]config.User, len(org.Users))
		for k, user := range org.Users {

			_roles := make([]config.Role, len(user.Roles))
			for k, role := range user.Roles {
				_roles[k] = config.Role{
					ID:   uint64(k),
					Name: role.Name}
			}
			orgUsers[k] = config.User{
				ID:       user.ID,
				Email:    user.Email,
				Password: user.Password,
				Roles:    _roles}
		}

		_orgs[i] = config.Organization{
			ID:      org.ID,
			Name:    org.Name,
			Farms:   _farms,
			Users:   orgUsers,
			License: license}
	}

	return &config.Server{
		Interval: yamlConfig.Interval,
		Timezone: yamlConfig.Timezone,
		Mode:     yamlConfig.Mode,
		Smtp: &config.Smtp{
			Enable:    yamlConfig.Smtp.Enable,
			Host:      yamlConfig.Smtp.Host,
			Port:      yamlConfig.Smtp.Port,
			Username:  yamlConfig.Smtp.Username,
			Password:  yamlConfig.Smtp.Password,
			Recipient: yamlConfig.Smtp.Recipient},
		Organizations: _orgs}, nil
}
