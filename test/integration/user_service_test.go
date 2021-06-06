// +build integration

package integration

import (
	"fmt"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/test"
	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {
	_, userService := createUserService()
	userByID, err := userService.GetUserByID(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, userByID.GetID())
	assert.Equal(t, "admin", userByID.GetEmail())

	role, err := userService.GetRole(1, 0)

	print(fmt.Sprintf("%+v", role))

	assert.Nil(t, err)
	assert.Equal(t, common.DEFAULT_ROLE, role.GetName())

	test.CleanupIntegrationTest()
}

func createUserService() (common.Context, service.UserService) {
	ctx := test.NewIntegrationTestContext()
	userDAO := dao.NewUserDAO(ctx)
	orgDAO := dao.NewOrganizationDAO(ctx)
	roleDAO := dao.NewRoleDAO(ctx)
	userMapper := mapper.NewUserMapper()
	authService := service.NewLocalAuthService(ctx, userDAO, userMapper)
	return ctx, service.NewUserService(ctx, userDAO, orgDAO, roleDAO, userMapper, authService)
}
