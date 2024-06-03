package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
	"github.com/stretchr/testify/assert"
)

func TestMetricCRUD(t *testing.T) {

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
	currentTest.gorm.AutoMigrate(&config.MetricStruct{})
	currentTest.gorm.AutoMigrate(&config.ConditionStruct{})
	currentTest.gorm.AutoMigrate(&config.ScheduleStruct{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStruct{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStepStruct{})

	metricDAO := NewMetricDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, metricDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)

	dstest.TestMetricCRUD(t, metricDAO, org)
}

func TestMetricGetByDevice(t *testing.T) {

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

	metricDAO := NewMetricDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, metricDAO)

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

	dstest.TestMetricGetByDevice(t, farmDAO, deviceDAO,
		metricDAO, permissionDAO, org)
}
