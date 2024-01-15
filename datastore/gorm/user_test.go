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

	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Permission{})

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
	user.SetRoles([]*config.Role{role})
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
	//org.SetUsers([]config.UserConfig{user})
	err = orgDAO.Save(org)
	assert.Nil(t, err)

	permission := config.NewPermission()
	permission.SetOrgID(org.GetID())
	permission.SetUserID(user.GetID())
	permission.SetRoleID(role.GetID())
	permission.SetFarmID(farm.GetID())
	err = permissionDAO.Save(permission)
	assert.Nil(t, err)

	users, err := orgDAO.GetUsers(org.GetID())
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

	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_SQLITE)

	email := "root@localhost"
	userID := idGenerator.NewID(email)

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)

	entity := &config.User{
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

	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_SQLITE)

	email := "root@localhost"
	userID := idGenerator.NewID(email)

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)

	entity := &config.User{
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

	user := &config.User{
		ID:       persistedUser.GetID(),
		Email:    "nologin@localhost",
		Password: "test123"}

	err = userDAO.Save(user)
	assert.Nil(t, err)

	persistedUser2, err := userDAO.Get(persistedUser.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, user.GetPassword(), persistedUser2.GetPassword())
}

func TestUserDAO_WithRoles(t *testing.T) {
	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_SQLITE)

	email := "root@localhost"
	userID := idGenerator.NewID(email)

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	permissionDAO := NewPermissionDAO(currentTest.logger, currentTest.gorm)

	role := config.NewRole()
	role.SetName("test123")

	user := config.NewUser()
	user.SetID(userID)
	user.SetEmail(email)
	user.SetRoles([]*config.Role{role})
	user.SetPassword("test")

	err := userDAO.Save(user)
	assert.Nil(t, err)

	permissionDAO.Save(&config.Permission{
		UserID: user.GetID(),
		RoleID: role.GetID()})

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
