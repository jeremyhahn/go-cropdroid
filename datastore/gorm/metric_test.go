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

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.DeviceSetting{})
	currentTest.gorm.AutoMigrate(&config.Metric{})
	currentTest.gorm.AutoMigrate(&config.Metric{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})
	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

	metricDAO := NewMetricDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, metricDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)

	dstest.TestMetricCRUD(t, metricDAO, org)
}

func TestMetricGetByDevice(t *testing.T) {

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
