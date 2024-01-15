package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"

	"github.com/stretchr/testify/assert"
)

func TestConditionCRUD(t *testing.T, conditionDAO dao.ConditionDAO,
	org *config.Organization) {

	farm1 := org.GetFarms()[0]
	device1 := farm1.GetDevices()[1]
	channel1 := device1.GetChannels()[0]
	condition1 := channel1.GetConditions()[0]
	condition2 := channel1.GetConditions()[1]

	assert.Equal(t, condition1.GetComparator(), ">")
	assert.Equal(t, condition2.GetComparator(), "<")

	err := conditionDAO.Save(farm1.GetID(), device1.GetID(), condition1)
	assert.Nil(t, err)

	err = conditionDAO.Save(farm1.GetID(), device1.GetID(), condition2)
	assert.Nil(t, err)

	persistedConditions, err := conditionDAO.GetByChannelID(
		farm1.GetID(), device1.GetID(), channel1.GetID(),
		common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(persistedConditions))

	persistedCondition1 := persistedConditions[0]
	assert.Equal(t, condition1.GetID(), persistedCondition1.GetID())
	assert.Equal(t, condition1.GetChannelID(), persistedCondition1.GetChannelID())
	assert.Equal(t, condition1.GetMetricID(), persistedCondition1.GetMetricID())
	assert.Equal(t, condition1.GetComparator(), persistedCondition1.GetComparator())
	assert.Equal(t, condition1.GetThreshold(), persistedCondition1.GetThreshold())

	err = conditionDAO.Delete(farm1.GetID(),
		device1.GetID(), persistedCondition1)
	assert.Nil(t, err)

	persistedConditions, err = conditionDAO.GetByChannelID(
		farm1.GetID(), device1.GetID(), channel1.GetID(),
		common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedConditions))
}
