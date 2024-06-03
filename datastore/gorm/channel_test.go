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

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.FarmStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceSettingStruct{})
	currentTest.gorm.AutoMigrate(&config.ChannelStruct{})
	currentTest.gorm.AutoMigrate(&config.ConditionStruct{})
	currentTest.gorm.AutoMigrate(&config.ScheduleStruct{})

	channelDAO := NewChannelDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, channelDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)

	dstest.TestChannelCRUD(t, channelDAO, org)
}

func TestChannelGetByDevice(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})
	currentTest.gorm.AutoMigrate(&config.MetricStruct{})
	currentTest.gorm.AutoMigrate(&config.ChannelStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceSettingStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceStruct{})
	currentTest.gorm.AutoMigrate(&config.FarmStruct{})
	currentTest.gorm.AutoMigrate(&config.OrganizationStruct{})
	currentTest.gorm.AutoMigrate(&config.ConditionStruct{})
	currentTest.gorm.AutoMigrate(&config.ScheduleStruct{})

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
