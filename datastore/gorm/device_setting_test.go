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

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.DeviceSetting{})
	currentTest.gorm.AutoMigrate(&config.Metric{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})
	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

	deviceConfigDAO := NewDeviceDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, deviceConfigDAO)

	deviceSettingsDAO := NewDeviceSettingDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, deviceSettingsDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)

	dstest.TestDeviceSettingCRUD(t, deviceConfigDAO, deviceSettingsDAO, org)
}
