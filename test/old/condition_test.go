package test

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/test/mocks/dao"
	"github.com/stretchr/testify/assert"
)

func TestIsTrue(t *testing.T) {

	_, conditionService := newConditionService()

	conditionConfig := &config.ConditionStruct{Comparator: ">", Threshold: 20}
	result, err := conditionService.IsTrue(conditionConfig, 21.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.ConditionStruct{Comparator: ">", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.ConditionStruct{Comparator: ">", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.ConditionStruct{Comparator: ">=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 21.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.ConditionStruct{Comparator: ">=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.ConditionStruct{Comparator: ">=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.ConditionStruct{Comparator: "<", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.ConditionStruct{Comparator: "<", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.ConditionStruct{Comparator: "<=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 21.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.ConditionStruct{Comparator: "<=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.ConditionStruct{Comparator: "<=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.ConditionStruct{Comparator: "=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.ConditionStruct{Comparator: "=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.ConditionStruct{Comparator: "=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 21.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)
}

func newConditionService() (*dao.MockConditionDAO, service.ConditionServicer) {
	app, _ := NewUnitTestSession()
	mockConditionDAO := dao.NewMockConditionDAO()
	mapper := mapper.NewConditionMapper()
	//mockConfigService := mocks.NewMockConfigService()
	return mockConditionDAO, service.NewConditionService(app.Logger, mockConditionDAO, mapper)
}
