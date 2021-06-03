package gorm

import (
	"testing"

	"github.com/jeremyhahn/cropdroid/config"
	"github.com/stretchr/testify/assert"
)

func TestUserDAO_CreateAndGetByEmail(t *testing.T) {
	currentTest := NewIntegrationTest()

	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)

	entity := &config.User{
		Email:    "root@localhost",
		Password: "test"}
	err := userDAO.Create(entity)
	assert.Nil(t, err)

	user, err := userDAO.GetByEmail("root@localhost")
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, entity.GetPassword(), user.GetPassword())
	assert.Equal(t, "root@localhost", user.GetEmail())
	assert.Equal(t, "test", user.GetPassword())

	currentTest.Cleanup()
}

func TestUserDAO_Update(t *testing.T) {
	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)

	err := userDAO.Create(&config.User{
		Email:    "root@localhost",
		Password: "test"})
	assert.Nil(t, err)

	persistedUser, err := userDAO.GetByEmail("root@localhost")
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

	persistedUser2, err := userDAO.GetByEmail("nologin@localhost")
	assert.Nil(t, err)
	assert.Equal(t, user.GetPassword(), persistedUser2.GetPassword())

	currentTest.Cleanup()
}

func TestUserDAO_WithRoles(t *testing.T) {
	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("test")

	err := userDAO.Create(user)
	assert.Nil(t, err)

	role := config.NewRole()
	role.SetName("test123")

	err = roleDAO.Save(role)
	assert.Nil(t, err)

	currentTest.gorm.Create(&config.Permission{
		UserID: user.GetID(),
		RoleID: role.GetID()})

	persisted, err := userDAO.GetByEmail("root@localhost")
	assert.Nil(t, err)
	assert.NotNil(t, persisted)
	assert.Equal(t, user.GetEmail(), persisted.GetEmail())
	assert.Equal(t, user.GetPassword(), persisted.GetPassword())
	assert.Equal(t, "root@localhost", persisted.GetEmail())
	assert.Equal(t, "test", persisted.GetPassword())
	assert.Equal(t, 1, len(persisted.GetRoles()))
	assert.Equal(t, "test123", persisted.GetRoles()[0].GetName())

	currentTest.Cleanup()
}
