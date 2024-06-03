package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
	"github.com/stretchr/testify/assert"
)

func TestDeviceSettingCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})
	currentTest.gorm.AutoMigrate(&config.OrganizationStruct{})
	currentTest.gorm.AutoMigrate(&config.FarmStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceSettingStruct{})
	currentTest.gorm.AutoMigrate(&config.MetricStruct{})
	currentTest.gorm.AutoMigrate(&config.ChannelStruct{})
	currentTest.gorm.AutoMigrate(&config.ConditionStruct{})
	currentTest.gorm.AutoMigrate(&config.ScheduleStruct{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStruct{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStepStruct{})

	deviceConfigDAO := NewDeviceDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, deviceConfigDAO)

	deviceSettingsDAO := NewDeviceSettingDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, deviceSettingsDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)

	dstest.TestDeviceSettingCRUD(t, deviceConfigDAO, deviceSettingsDAO, org)
}
