package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/stretchr/testify/assert"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

var DEFAULT_CONSISTENCY_LEVEL = common.CONSISTENCY_LOCAL

func TestFarmAssociations(t *testing.T) {

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

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm,
		currentTest.idGenerator)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)
	farm1 := org.GetFarms()[0]

	dstest.TestFarmAssociations(t, currentTest.idGenerator,
		farmDAO, farm1)
}

func TestFarmGetByIds(t *testing.T) {

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

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm,
		currentTest.idGenerator)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)
	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	dstest.TestFarmGetByIds(t, farmDAO, farm1, farm2)
}

func TestFarmGetPage(t *testing.T) {

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

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm,
		currentTest.idGenerator)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)
	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	err := farmDAO.Save(farm1)
	assert.Nil(t, err)

	err = farmDAO.Save(farm2)
	assert.Nil(t, err)

	dstest.TestFarmGetPage(t, farmDAO, farm1, farm2)
}

func TestFarmGet(t *testing.T) {

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

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm,
		currentTest.idGenerator)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)
	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	err := farmDAO.Save(farm1)
	assert.Nil(t, err)

	err = farmDAO.Save(farm2)
	assert.Nil(t, err)

	dstest.TestFarmGet(t, farmDAO, farm1, farm2)
}
