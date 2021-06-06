package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestFarmAssociations(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Controller{})
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

	persisted, err := farmDAO.Get(farm.GetID())
	assert.Nil(t, err)

	assert.Equal(t, farm.GetID(), persisted.GetID())

	assert.Equal(t, 1, len(persisted.GetUsers()))
	assert.Equal(t, "root@localhost", persisted.GetUsers()[0].GetEmail())

	assert.Equal(t, 1, len(persisted.GetUsers()[0].GetRoles()))
	assert.Equal(t, "test", persisted.GetUsers()[0].GetRoles()[0].GetName())

	currentTest.Cleanup()
}

func TestGetAll(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Controller{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm)

	farm1 := config.NewFarm()
	farm1.SetName("Test Farm")
	farm1.SetMode("test")
	farm1.SetInterval(60)

	farm2 := config.NewFarm()
	farm2.SetName("Test Farm 2")
	farm2.SetMode("test2")
	farm2.SetInterval(59)

	err := farmDAO.Create(farm1)
	assert.Nil(t, err)

	err = farmDAO.Create(farm2)
	assert.Nil(t, err)

	farms, err := farmDAO.GetAll()
	assert.Nil(t, err)

	assert.Equal(t, 2, len(farms))

	currentTest.Cleanup()
}

func TestGet(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Controller{})
	currentTest.gorm.AutoMigrate(&config.ControllerConfigItem{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm)

	farm1 := config.NewFarm()
	farm1.SetMode("test")
	farm1.SetControllers([]config.Controller{
		config.Controller{
			Type: "server",
			Configs: []config.ControllerConfigItem{
				config.ControllerConfigItem{
					Key:   "name",
					Value: "Test Farm"},
				config.ControllerConfigItem{
					Key:   "mode",
					Value: "test"},
				config.ControllerConfigItem{
					Key:   "interval",
					Value: "59"}}}})

	farm2 := config.NewFarm()
	farm2.SetControllers([]config.Controller{
		config.Controller{
			Type: "server",
			Configs: []config.ControllerConfigItem{
				config.ControllerConfigItem{
					Key:   "name",
					Value: "Test Farm 2"},
				config.ControllerConfigItem{
					Key:   "mode",
					Value: "test2"},
				config.ControllerConfigItem{
					Key:   "interval",
					Value: "60"}}}})

	err := farmDAO.Create(farm1)
	assert.Nil(t, err)

	err = farmDAO.Create(farm2)
	assert.Nil(t, err)

	persitedFarm1, err := farmDAO.Get(1)
	assert.Nil(t, err)
	assert.Equal(t, "Test Farm", persitedFarm1.GetName())
	assert.Equal(t, "test", persitedFarm1.GetMode())
	assert.Equal(t, 59, persitedFarm1.GetInterval())

	persitedFarm2, err := farmDAO.Get(2)
	assert.Nil(t, err)
	assert.Equal(t, "Test Farm 2", persitedFarm2.GetName())
	assert.Equal(t, "test2", persitedFarm2.GetMode())
	assert.Equal(t, 60, persitedFarm2.GetInterval())

	currentTest.Cleanup()
}
