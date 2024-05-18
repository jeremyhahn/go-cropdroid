package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/stretchr/testify/assert"
)

func TestMetricCRUD(t *testing.T, metricDAO dao.MetricDAO,
	org *config.Organization) {

	farm1 := org.GetFarms()[0]
	device1 := farm1.GetDevices()[1]
	metric1 := device1.GetMetrics()[0]

	err := metricDAO.Save(farm1.ID, metric1)
	assert.Nil(t, err)

	persistedMetric, err := metricDAO.Get(farm1.ID, device1.ID,
		metric1.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)

	assert.Equal(t, metric1.ID, persistedMetric.ID)
	assert.Equal(t, metric1.GetDeviceID(), persistedMetric.GetDeviceID())
	assert.Equal(t, metric1.GetName(), persistedMetric.GetName())
	assert.Equal(t, metric1.IsEnabled(), persistedMetric.IsEnabled())
	assert.Equal(t, metric1.IsNotify(), persistedMetric.IsNotify())
}

func TestMetricGetByDevice(t *testing.T, farmDAO dao.FarmDAO,
	deviceDAO dao.DeviceDAO, metricDAO dao.MetricDAO,
	permissionDAO dao.PermissionDAO, org *config.Organization) {

	farm1 := org.GetFarms()[0]
	device1 := farm1.GetDevices()[1]
	metric1 := device1.GetMetrics()[0]

	err := farmDAO.Save(farm1)
	assert.Nil(t, err)

	permissionDAO.Save(&config.Permission{
		OrganizationID: 0,
		FarmID:         farm1.ID,
		UserID:         farm1.GetUsers()[0].ID,
		RoleID:         farm1.GetUsers()[0].GetRoles()[0].ID})

	newMetricName := "newtest"
	metric1.SetName(newMetricName)
	err = metricDAO.Save(farm1.ID, metric1)
	assert.Nil(t, err)

	metric, err := metricDAO.Get(farm1.ID, device1.ID,
		metric1.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, metric1.ID, metric.ID)

	persistedMetrics, err := metricDAO.GetByDevice(farm1.ID,
		device1.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(persistedMetrics))

	// Gorm and Raft items are stored in different order.
	// Raft returns records in the same order they were saved.
	// GORM returns records ordered by id.
	// This loop performs assertions regardless of order
	found := false
	for _, persistedMetric := range persistedMetrics {
		if metric1.ID == persistedMetric.ID {
			assert.Equal(t, metric1.GetDeviceID(), persistedMetric.GetDeviceID())
			assert.Equal(t, newMetricName, persistedMetric.GetName())
			assert.Equal(t, metric1.IsEnabled(), persistedMetric.IsEnabled())
			assert.Equal(t, metric1.IsNotify(), persistedMetric.IsNotify())
			found = true
		}
	}
	assert.True(t, found)
}
