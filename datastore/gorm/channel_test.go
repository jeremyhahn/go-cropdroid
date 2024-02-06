package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"

	"github.com/stretchr/testify/assert"
)

func TestChannelCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.DeviceSetting{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})

	channelDAO := NewChannelDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, channelDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)

	dstest.TestChannelCRUD(t, channelDAO, org)
}

func TestChannelGetByDevice(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Metric{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.DeviceSetting{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})

	deviceDAO := NewDeviceDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, deviceDAO)

	channelDAO := NewChannelDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, channelDAO)

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm,
		currentTest.idGenerator)
	assert.NotNil(t, farmDAO)

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, userDAO)

	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, roleDAO)

	permissionDAO := NewPermissionDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, permissionDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)

	dstest.TestChannelGetByDevice(t, farmDAO, deviceDAO, channelDAO, permissionDAO, org)
}
