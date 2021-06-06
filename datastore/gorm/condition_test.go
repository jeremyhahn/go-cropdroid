package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestConditionCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Condition{})

	conditionDAO := NewConditionDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, conditionDAO)

	condition1 := &config.Condition{
		ID:         1,
		ChannelID:  1,
		MetricID:   2,
		Comparator: ">",
		Threshold:  10}

	condition2 := &config.Condition{
		//ID:         2,
		ChannelID:  2,
		MetricID:   3,
		Comparator: ">=",
		Threshold:  5001}

	err := conditionDAO.Create(condition1)
	assert.Nil(t, err)

	err = conditionDAO.Save(condition2)
	assert.Nil(t, err)

	persistedConditions, err := conditionDAO.GetByChannelID(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedConditions))

	persistedCondition1 := persistedConditions[0]
	assert.Equal(t, condition1.ID, persistedCondition1.GetID())
	assert.Equal(t, condition1.ChannelID, persistedCondition1.GetChannelID())
	assert.Equal(t, condition1.MetricID, persistedCondition1.GetMetricID())
	assert.Equal(t, condition1.Comparator, persistedCondition1.GetComparator())
	assert.Equal(t, condition1.Threshold, persistedCondition1.GetThreshold())

	currentTest.Cleanup()
}

func TestConditionGetByUserOrgAndChannelID(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Controller{})
	currentTest.gorm.AutoMigrate(&config.ControllerConfigItem{})
	currentTest.gorm.AutoMigrate(&config.Farm{})

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, farmDAO)

	conditionDAO := NewConditionDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, conditionDAO)

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, userDAO)

	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, roleDAO)

	role := config.NewRole()
	role.SetName("test")

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("$ecret")

	channel1 := config.NewChannel()
	channel1.SetID(1)
	channel1.SetControllerID(1)
	channel1.SetChannelID(3)
	channel1.SetName("Test Channel 1")
	channel1.SetEnable(true)
	channel1.SetNotify(true)
	channel1.SetDuration(2)
	channel1.SetDebounce(3)
	channel1.SetBackoff(4)

	controller1 := config.NewController()
	controller1.SetType("fake")
	controller1.SetDescription("This is a fake controller used for integration testing")
	controller1.SetInterval(30)
	controller1.SetChannels([]config.Channel{*channel1})

	farm := config.NewFarm()
	farm.SetName("Test Farm")
	farm.SetMode("test")
	farm.SetInterval(60)
	farm.SetControllers([]config.Controller{*controller1})

	err := farmDAO.Save(farm)
	assert.Nil(t, err)

	err = userDAO.Create(user)
	assert.Nil(t, err)

	err = roleDAO.Create(role)
	assert.Nil(t, err)

	permission := &config.Permission{
		OrganizationID: 0,
		FarmID:         farm.GetID(),
		UserID:         user.GetID(),
		RoleID:         role.GetID()}
	currentTest.gorm.Create(permission)

	condition1 := &config.Condition{
		ID:         1,
		ChannelID:  1,
		MetricID:   2,
		Comparator: ">",
		Threshold:  10}
	err = conditionDAO.Create(condition1)
	assert.Nil(t, err)

	persistedConditions, err := conditionDAO.GetByOrgUserAndChannelID(0, user.GetID(), channel1.GetID())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedConditions))

	persistedCondition1 := persistedConditions[0]
	assert.Equal(t, condition1.ID, persistedCondition1.GetID())
	assert.Equal(t, condition1.ChannelID, persistedCondition1.GetChannelID())
	assert.Equal(t, condition1.MetricID, persistedCondition1.GetMetricID())
	assert.Equal(t, condition1.Comparator, persistedCondition1.GetComparator())
	assert.Equal(t, condition1.Threshold, persistedCondition1.GetThreshold())

	currentTest.gorm.Delete(permission)
	conditions, err := conditionDAO.GetByOrgUserAndChannelID(0, user.GetID(), channel1.GetID())
	assert.Nil(t, err)
	assert.Equal(t, 0, len(conditions))

	currentTest.Cleanup()
}
