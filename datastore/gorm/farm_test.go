package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

var DEFAULT_CONSISTENCY_LEVEL = common.CONSISTENCY_LOCAL

func TestFarmAssociations(t *testing.T) {

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
	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm)

	role := config.NewRole()
	role.SetName("test")

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("$ecret")
	user.SetRoles([]config.Role{*role})

	farm := config.NewFarm()
	farm.SetName("Test Farm")
	farm.SetMode("test")
	farm.SetInterval(60)
	farm.SetUsers([]config.User{*user})

	err := farmDAO.Create(farm)
	assert.Nil(t, err)

	currentTest.gorm.Create(&config.Permission{
		OrganizationID: 0,
		FarmID:         farm.GetID(),
		UserID:         user.GetID(),
		RoleID:         role.GetID()})

	persisted, err := farmDAO.Get(farm.GetID(), DEFAULT_CONSISTENCY_LEVEL)
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
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})
	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

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
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.DeviceConfigItem{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})
	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm)

	farm1 := config.NewFarm()
	farm1.SetMode("test")
	farm1.SetDevices([]config.Device{
		{
			Type: "server",
			Configs: []config.DeviceConfigItem{
				{
					Key:   "name",
					Value: "Test Farm"},
				{
					Key:   "mode",
					Value: "test"},
				{
					Key:   "interval",
					Value: "59"}}}})

	farm2 := config.NewFarm()
	farm2.SetDevices([]config.Device{
		{
			Type: "server",
			Configs: []config.DeviceConfigItem{
				{
					Key:   "name",
					Value: "Test Farm 2"},
				{
					Key:   "mode",
					Value: "test2"},
				{
					Key:   "interval",
					Value: "60"}}}})

	err := farmDAO.Create(farm1)
	assert.Nil(t, err)

	err = farmDAO.Create(farm2)
	assert.Nil(t, err)

	persitedFarm1, err := farmDAO.Get(1, DEFAULT_CONSISTENCY_LEVEL)
	assert.Nil(t, err)
	assert.Equal(t, "Test Farm", persitedFarm1.GetName())
	assert.Equal(t, "test", persitedFarm1.GetMode())
	assert.Equal(t, 59, persitedFarm1.GetInterval())

	persitedFarm2, err := farmDAO.Get(2, DEFAULT_CONSISTENCY_LEVEL)
	assert.Nil(t, err)
	assert.Equal(t, "Test Farm 2", persitedFarm2.GetName())
	assert.Equal(t, "test2", persitedFarm2.GetMode())
	assert.Equal(t, 60, persitedFarm2.GetInterval())

	currentTest.Cleanup()
}

func TestCount(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Farm{})

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm)

	farm1 := config.NewFarm()
	err := farmDAO.Create(farm1)
	assert.Nil(t, err)

	farm2 := config.NewFarm()
	err = farmDAO.Create(farm2)
	assert.Nil(t, err)

	count, err := farmDAO.Count()
	assert.Nil(t, err)
	assert.Equal(t, int64(2), count)

	err = farmDAO.Delete(farm1)
	assert.Nil(t, err)
	count, err = farmDAO.Count()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)

	// err = farmDAO.DeleteById(farm2.GetID())
	// assert.Nil(t, err)
	// count, err = farmDAO.Count()
	// assert.Nil(t, err)
	// assert.Equal(t, int64(0), count)

	currentTest.Cleanup()
}
