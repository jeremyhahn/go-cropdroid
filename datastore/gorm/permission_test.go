package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestUserRoleRelationship(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, farmDAO)

	role := config.NewRole()
	role.SetName("test")

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("$ecret")
	user.SetRoles([]config.Role{*role})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	err := userDAO.Create(user)
	assert.Nil(t, err)

	persisted, err := userDAO.GetByEmail(user.GetEmail())
	assert.Nil(t, err)

	assert.Equal(t, persisted.GetID(), user.GetID())
	assert.Equal(t, "root@localhost", persisted.GetEmail())
	assert.Equal(t, "$ecret", persisted.GetPassword())
	assert.Equal(t, 1, len(persisted.GetRoles()))
	assert.Equal(t, "test", persisted.GetRoles()[0].GetName())

	currentTest.Cleanup()
}

func TestPermissions(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm)

	role := config.NewRole()
	role.SetName("test")

	err := roleDAO.Create(role)
	assert.Nil(t, err)

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("$ecret")
	//user.SetRoles([]config.Role{*role})

	err = userDAO.Create(user)
	assert.Nil(t, err)

	farm := config.NewFarm()
	farm.SetName("Test Farm")
	farm.SetMode("test")
	farm.SetInterval(60)
	//farm.SetUsers([]config.User{*user})

	err = farmDAO.Create(farm)
	assert.Nil(t, err)

	currentTest.gorm.Create(&config.Permission{
		OrganizationID: 0,
		FarmID:         farm.GetID(),
		UserID:         user.GetID(),
		RoleID:         role.GetID()})

	// Verify Get(id)
	persisted, err := farmDAO.Get(farm.GetID())
	assert.Nil(t, err)
	assert.Equal(t, farm.GetID(), persisted.GetID())

	// Verify GetByOrgAndUserID(orgID, userID)
	userFarms, err := farmDAO.GetByOrgAndUserID(farm.GetOrgID(), user.GetID())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(userFarms))
	assert.Equal(t, farm.GetID(), userFarms[0].GetID())

	currentTest.Cleanup()
}
