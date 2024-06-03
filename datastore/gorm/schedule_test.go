package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"

	"github.com/stretchr/testify/assert"
)

func TestScheduleCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.FarmStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceSettingStruct{})
	currentTest.gorm.AutoMigrate(&config.ChannelStruct{})
	currentTest.gorm.AutoMigrate(&config.ConditionStruct{})
	currentTest.gorm.AutoMigrate(&config.ScheduleStruct{})

	scheduleDAO := NewScheduleDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, scheduleDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)

	dstest.TestScheduleCRUD(t, scheduleDAO, org)
}
