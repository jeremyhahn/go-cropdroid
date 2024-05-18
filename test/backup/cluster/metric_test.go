//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"

	"github.com/stretchr/testify/assert"
)

func TestMetricCRUD(t *testing.T) {

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	orgDAO := NewRaftOrganizationDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), OrganizationClusterID, serverDAO)
	assert.NotNil(t, orgDAO)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)
	assert.NotNil(t, userDAO)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)
	assert.NotNil(t, farmDAO)

	permissionDAO := NewRaftPermissionDAO(Cluster.app.Logger,
		orgDAO, farmDAO, userDAO)
	assert.NotNil(t, permissionDAO)

	deviceDAO := NewRaftDeviceConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmDAO)
	assert.NotNil(t, deviceDAO)

	metricDAO := NewRaftMetricDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmDAO)
	assert.NotNil(t, metricDAO)

	org := createRaftTestOrganization(t, Cluster, ClusterID,
		serverDAO, userDAO, farmDAO)

	dstest.TestMetricCRUD(t, metricDAO, org)
}

func TestMetricGetByDevice(t *testing.T) {

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	orgDAO := NewRaftOrganizationDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), OrganizationClusterID, serverDAO)
	assert.NotNil(t, orgDAO)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)
	assert.NotNil(t, userDAO)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)
	assert.NotNil(t, farmDAO)

	permissionDAO := NewRaftPermissionDAO(Cluster.app.Logger,
		orgDAO, farmDAO, userDAO)
	assert.NotNil(t, permissionDAO)

	deviceDAO := NewRaftDeviceConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmDAO)
	assert.NotNil(t, deviceDAO)

	metricDAO := NewRaftMetricDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmDAO)
	assert.NotNil(t, metricDAO)

	org := createRaftTestOrganization(t, Cluster,
		ClusterID, serverDAO, userDAO, farmDAO)

	dstest.TestMetricGetByDevice(t, farmDAO, deviceDAO,
		metricDAO, permissionDAO, org)
}
