// +build cluster

package service

import (
	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/datastore"
	"github.com/jeremyhahn/cropdroid/mapper"
)

func CreateClusterServiceRegistry(_app *app.App, daos datastore.DatastoreRegistry, mappers mapper.MapperRegistry) ServiceRegistry {

	farmDAO := daos.GetFarmDAO()
	registry := CreateServiceRegistry(_app, daos, mappers)

	gas := NewGoogleAuthService(_app, daos.GetUserDAO(), daos.GetOrganizationDAO(), farmDAO, mappers.GetUserMapper())

	authServices := make(map[int]AuthService, 2)
	authServices[common.AUTH_TYPE_LOCAL] = registry.GetAuthService()
	authServices[common.AUTH_TYPE_GOOGLE] = gas
	userService := NewUserService(_app, daos.GetUserDAO(), daos.GetOrganizationDAO(), daos.GetRoleDAO(), farmDAO,
		mappers.GetUserMapper(), authServices, registry)

	registry.SetGoogleAuthService(gas)
	registry.SetUserService(userService)

	return registry
}
