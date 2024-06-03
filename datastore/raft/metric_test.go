//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"

	"github.com/stretchr/testify/assert"
)

func TestMetricCRUD(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	org, _, farmDAO, _ := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)

	metricDAO := NewRaftMetricDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)
	assert.NotNil(t, metricDAO)

	dstest.TestMetricCRUD(t, metricDAO, org)
}

func TestMetricGetByDevice(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	org, orgDAO, farmDAO, _ := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)

	deviceDAO := NewRaftDeviceConfigDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)
	assert.NotNil(t, deviceDAO)

	userDAO := NewGenericRaftDAO[*config.UserStruct](
		IntegrationTestCluster.app.Logger,
		raftNode1,
		UserClusterID)
	assert.NotNil(t, userDAO)

	permissionDAO := NewRaftPermissionDAO(
		IntegrationTestCluster.app.Logger,
		orgDAO,
		farmDAO,
		userDAO)

	metricDAO := NewRaftMetricDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)
	assert.NotNil(t, metricDAO)

	dstest.TestMetricGetByDevice(t, farmDAO, deviceDAO,
		metricDAO, permissionDAO, org)
}
