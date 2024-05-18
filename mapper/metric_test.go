package mapper

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/stretchr/testify/assert"
)

func TestMetricMapperMapConfigToModel(t *testing.T) {

	mapper := NewMetricMapper()

	var metricConfig *config.Metric = &config.Metric{}
	metricConfig.SetID(1)
	metricConfig.SetDeviceID(2)
	metricConfig.SetName("Test Metric")
	metricConfig.SetKey("test")
	metricConfig.SetEnable(true)
	metricConfig.SetNotify(true)
	metricConfig.SetUnit("°")
	metricConfig.SetAlarmLow(12.34)
	metricConfig.SetAlarmHigh(56.0)

	// metric, ok := metricConfig.(common.Metric)
	// assert.True(t, ok)
	// assert.ObjectsAreEqual(metricConfig, metric)

	// assert.Equal(t, metricConfig.GetID(), metric.GetID())
	// assert.Equal(t, metricConfig.GetDeviceID(), metric.GetDeviceID())
	// assert.Equal(t, metricConfig.GetName(), metric.GetName())
	// assert.Equal(t, metricConfig.GetKey(), metric.GetKey())
	// assert.Equal(t, metricConfig.IsEnabled(), metric.IsEnabled())
	// assert.Equal(t, metricConfig.IsNotify(), metric.IsNotify())
	// assert.Equal(t, metricConfig.GetUnit(), metric.GetUnit())
	// assert.Equal(t, metricConfig.GetAlarmLow(), metric.GetAlarmLow())
	// assert.Equal(t, metricConfig.GetAlarmHigh(), metric.GetAlarmHigh())

	// Mapper must return new model object to prevent config objects from being updated by model pointer.
	metricConfig.SetID(2)
	assert.Equal(t, uint64(2), metricConfig.ID)

	model := mapper.MapConfigToModel(metricConfig)
	assert.ObjectsAreEqual(metricConfig, model)
	model.SetID(3)
	assert.Equal(t, uint64(2), metricConfig.ID) // This is the desired behavior
}

func TestMetricMapperMapModelToConfig(t *testing.T) {

	mapper := NewMetricMapper()

	var metric common.Metric = &model.Metric{}
	metric.SetID(1)
	metric.SetDeviceID(2)
	metric.SetName("Test Metric")
	metric.SetKey("test")
	metric.SetEnable(true)
	metric.SetNotify(true)
	metric.SetUnit("°")
	metric.SetAlarmLow(12.34)
	metric.SetAlarmHigh(56.0)

	config := mapper.MapModelToConfig(metric)
	assert.ObjectsAreEqual(metric, config)

	assert.Equal(t, metric.GetID(), config.ID)
	assert.Equal(t, metric.GetDeviceID(), config.GetDeviceID())
	assert.Equal(t, metric.GetName(), config.GetName())
	assert.Equal(t, metric.GetKey(), config.GetKey())
	assert.Equal(t, metric.IsEnabled(), config.IsEnabled())
	assert.Equal(t, metric.IsNotify(), config.IsNotify())
	assert.Equal(t, metric.GetUnit(), config.GetUnit())
	assert.Equal(t, metric.GetAlarmLow(), config.GetAlarmLow())
	assert.Equal(t, metric.GetAlarmHigh(), config.GetAlarmHigh())

	metric.SetDeviceID(20)
	assert.NotEqual(t, metric.GetDeviceID(), config.GetDeviceID())
}
