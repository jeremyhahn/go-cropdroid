// +build broken

package test

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/assert"
)

func TestIsTrue(t *testing.T) {

	_, conditionService := newConditionService()

	conditionConfig := &config.Condition{Comparator: ">", Threshold: 20}
	result, err := conditionService.IsTrue(conditionConfig, 21.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.Condition{Comparator: ">", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.Condition{Comparator: ">", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.Condition{Comparator: ">=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 21.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.Condition{Comparator: ">=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.Condition{Comparator: ">=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.Condition{Comparator: "<", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.Condition{Comparator: "<", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.Condition{Comparator: "<=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 21.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.Condition{Comparator: "<=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.Condition{Comparator: "<=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.Condition{Comparator: "=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 19.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	conditionConfig = &config.Condition{Comparator: "=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 20.0)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	conditionConfig = &config.Condition{Comparator: "=", Threshold: 20}
	result, err = conditionService.IsTrue(conditionConfig, 21.0)
	assert.Nil(t, err)
	assert.Equal(t, false, result)
}

func newConditionService() (*MockConditionDAO, service.ConditionService) {
	_, scope := NewUnitTestContext()
	dao := NewMockConditionDAO()
	mapper := mapper.NewConditionMapper()
	mockConfigService := NewMockConfigService()
	return dao, service.NewConditionService(scope, dao, mapper, mockConfigService)
}
