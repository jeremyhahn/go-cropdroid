package mapper

import (
	"testing"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/stretchr/testify/assert"
)

func TestMetrMapperMapEntityToModel(t *testing.T) {
	mapper := NewMetricMapper()
	entity := &config.Metric{
		ID:           1,
		ControllerID: 2,
		Name:         "Test Metric",
		Key:          "test",
		Enable:       true,
		Notify:       true,
		Unit:         "°",
		AlarmLow:     12.34,
		AlarmHigh:    56}
	model := mapper.MapEntityToModel(entity)
	assert.Equal(t, model.GetID(), entity.GetID())
	assert.Equal(t, model.GetControllerID(), entity.GetControllerID())
	assert.Equal(t, model.GetName(), entity.GetName())
	assert.Equal(t, model.GetKey(), entity.GetKey())
	assert.Equal(t, model.IsEnabled(), entity.IsEnabled())
	assert.Equal(t, model.IsNotify(), entity.IsNotify())
	assert.Equal(t, model.GetUnit(), entity.GetUnit())
	assert.Equal(t, model.GetAlarmLow(), entity.GetAlarmLow())
	assert.Equal(t, model.GetAlarmHigh(), entity.GetAlarmHigh())
}

func TestMetricMapperMapConfigToModel(t *testing.T) {

	mapper := NewMetricMapper()

	var metricConfig config.MetricConfig = &model.Metric{}
	metricConfig.SetID(1)
	metricConfig.SetControllerID(2)
	metricConfig.SetName("Test Metric")
	metricConfig.SetKey("test")
	metricConfig.SetEnable(true)
	metricConfig.SetNotify(true)
	metricConfig.SetUnit("°")
	metricConfig.SetAlarmLow(12.34)
	metricConfig.SetAlarmHigh(56.0)

	metric, ok := metricConfig.(common.Metric)
	assert.True(t, ok)
	assert.ObjectsAreEqual(metricConfig, metric)

	assert.Equal(t, metricConfig.GetID(), metric.GetID())
	assert.Equal(t, metricConfig.GetControllerID(), metric.GetControllerID())
	assert.Equal(t, metricConfig.GetName(), metric.GetName())
	assert.Equal(t, metricConfig.GetKey(), metric.GetKey())
	assert.Equal(t, metricConfig.IsEnabled(), metric.IsEnabled())
	assert.Equal(t, metricConfig.IsNotify(), metric.IsNotify())
	assert.Equal(t, metricConfig.GetUnit(), metric.GetUnit())
	assert.Equal(t, metricConfig.GetAlarmLow(), metric.GetAlarmLow())
	assert.Equal(t, metricConfig.GetAlarmHigh(), metric.GetAlarmHigh())

	// Mapper must return new model object to prevent config objects from being updated by model pointer.
	metric.SetID(2)
	assert.Equal(t, 2, metricConfig.GetID())

	model := mapper.MapConfigToModel(metricConfig)
	assert.ObjectsAreEqual(metricConfig, model)
	model.SetID(3)
	assert.Equal(t, 2, metricConfig.GetID()) // This is the desired behavior
}

func TestMetricMapperMapModelToEntity(t *testing.T) {

	mapper := NewMetricMapper()

	var metric common.Metric = &model.Metric{}
	metric.SetID(1)
	metric.SetControllerID(2)
	metric.SetName("Test Metric")
	metric.SetKey("test")
	metric.SetEnable(true)
	metric.SetNotify(true)
	metric.SetUnit("°")
	metric.SetAlarmLow(12.34)
	metric.SetAlarmHigh(56.0)

	config := mapper.MapModelToConfig(metric)
	assert.ObjectsAreEqual(metric, config)

	assert.Equal(t, metric.GetID(), config.GetID())
	assert.Equal(t, metric.GetControllerID(), config.GetControllerID())
	assert.Equal(t, metric.GetName(), config.GetName())
	assert.Equal(t, metric.GetKey(), config.GetKey())
	assert.Equal(t, metric.IsEnabled(), config.IsEnabled())
	assert.Equal(t, metric.IsNotify(), config.IsNotify())
	assert.Equal(t, metric.GetUnit(), config.GetUnit())
	assert.Equal(t, metric.GetAlarmLow(), config.GetAlarmLow())
	assert.Equal(t, metric.GetAlarmHigh(), config.GetAlarmHigh())

	metric.SetControllerID(20)
	assert.NotEqual(t, metric.GetControllerID(), config.GetControllerID())
}

func TestMetricMapperMapEntityToConfig(t *testing.T) {

	mapper := NewMetricMapper()
	entity := &config.Metric{
		ID:           1,
		ControllerID: 2,
		Name:         "Test Metric",
		Key:          "test",
		Enable:       true,
		Notify:       true,
		Unit:         "°",
		AlarmLow:     12.34,
		AlarmHigh:    56}

	config := mapper.MapConfigToModel(entity)
	assert.Equal(t, entity.GetID(), config.GetID())
	assert.Equal(t, entity.GetControllerID(), config.GetControllerID())
	assert.Equal(t, entity.GetName(), config.GetName())
	assert.Equal(t, entity.GetKey(), config.GetKey())
	assert.Equal(t, entity.IsEnabled(), config.IsEnabled())
	assert.Equal(t, entity.IsNotify(), config.IsNotify())
	assert.Equal(t, entity.GetUnit(), config.GetUnit())
	assert.Equal(t, entity.GetAlarmLow(), config.GetAlarmLow())
	assert.Equal(t, entity.GetAlarmHigh(), config.GetAlarmHigh())
}
