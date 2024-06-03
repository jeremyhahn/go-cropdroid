package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"
)

func TestUserDAO_CreateAndGetByOrgID(t *testing.T) {
	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.OrganizationStruct{})
	currentTest.gorm.AutoMigrate(&config.FarmStruct{})
	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})
	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_32BIT)
	orgDAO := NewOrganizationDAO(currentTest.logger, currentTest.gorm, idGenerator)
	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	permissionDAO := NewPermissionDAO(currentTest.logger, currentTest.gorm)
	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm,
		currentTest.idGenerator)

	role := config.NewRole()
	role.SetName("test123")
	err := roleDAO.Save(role)
	assert.Nil(t, err)

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("test")
	user.SetRoles([]*config.RoleStruct{role})
	err = userDAO.Save(user)
	assert.Nil(t, err)

	farm := config.NewFarm()
	farm.SetName("test farm")
	farm.SetOrganizationID(1)
	err = farmDAO.Save(farm)
	assert.Nil(t, err)

	org := config.NewOrganization()
	org.SetName("test org")
	//org.SetFarms([]config.FarmConfig{farm})
	org.SetUsers([]*config.UserStruct{user})
	err = orgDAO.Save(org)
	assert.Nil(t, err)

	permission := config.NewPermission()
	permission.SetOrgID(org.ID)
	permission.SetUserID(user.ID)
	permission.SetRoleID(role.ID)
	permission.SetFarmID(farm.ID)
	err = permissionDAO.Save(permission)
	assert.Nil(t, err)

	users, err := orgDAO.GetUsers(org.ID)
	assert.Nil(t, err)
	assert.NotNil(t, users)
	assert.Equal(t, len(users), 1)

	persisted := users[0]
	assert.Equal(t, user.GetEmail(), persisted.GetEmail())
	assert.Equal(t, user.GetPassword(), persisted.GetPassword())
	assert.Equal(t, "root@localhost", persisted.GetEmail())
	assert.Equal(t, "test", persisted.GetPassword())
	assert.Equal(t, 1, len(persisted.GetRoles()))
	assert.Equal(t, "test123", persisted.GetRoles()[0].GetName())
}

func TestUserDAO_CreateAndGetByEmail(t *testing.T) {
	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_SQLITE)

	email := "root@localhost"
	userID := idGenerator.NewStringID(email)

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)

	entity := &config.UserStruct{
		ID:       userID,
		Email:    email,
		Password: "test"}
	err := userDAO.Save(entity)
	assert.Nil(t, err)

	user, err := userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, entity.GetPassword(), user.GetPassword())
	assert.Equal(t, "root@localhost", user.GetEmail())
	assert.Equal(t, "test", user.GetPassword())
}

func TestUserDAO_Update(t *testing.T) {
	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_SQLITE)

	email := "root@localhost"
	userID := idGenerator.NewStringID(email)

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)

	entity := &config.UserStruct{
		ID:       userID,
		Email:    email,
		Password: "test"}
	err := userDAO.Save(entity)
	assert.Nil(t, err)

	persistedUser, err := userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, persistedUser)
	assert.Equal(t, "root@localhost", persistedUser.GetEmail())
	assert.Equal(t, "test", persistedUser.GetPassword())

	user := &config.UserStruct{
		ID:       persistedUser.ID,
		Email:    "nologin@localhost",
		Password: "test123"}

	err = userDAO.Save(user)
	assert.Nil(t, err)

	persistedUser2, err := userDAO.Get(persistedUser.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, user.GetPassword(), persistedUser2.GetPassword())
}

func TestUserDAO_WithRoles(t *testing.T) {
	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_SQLITE)

	email := "root@localhost"
	userID := idGenerator.NewStringID(email)

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	permissionDAO := NewPermissionDAO(currentTest.logger, currentTest.gorm)

	role := config.NewRole()
	role.SetName("test123")

	user := config.NewUser()
	user.SetID(userID)
	user.SetEmail(email)
	user.SetRoles([]*config.RoleStruct{role})
	user.SetPassword("test")

	err := userDAO.Save(user)
	assert.Nil(t, err)

	permissionDAO.Save(&config.PermissionStruct{
		UserID: user.ID,
		RoleID: role.ID})

	persisted, err := userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, persisted)
	assert.Equal(t, user.GetEmail(), persisted.GetEmail())
	assert.Equal(t, user.GetPassword(), persisted.GetPassword())
	assert.Equal(t, "root@localhost", persisted.GetEmail())
	assert.Equal(t, "test", persisted.GetPassword())
	assert.Equal(t, 1, len(persisted.GetRoles()))
	assert.Equal(t, "test123", persisted.GetRoles()[0].GetName())
}
