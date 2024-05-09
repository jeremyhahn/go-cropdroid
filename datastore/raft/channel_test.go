//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestChannelCRUD(t *testing.T) {

	ClusterID = 1

	raftNode1 := IntegrationTestCluster.GetRaftNode1()

	serverDAO := NewGenericRaftDAO[*config.Server](
		IntegrationTestCluster.app.Logger,
		raftNode1,
		ClusterID)
	assert.NotNil(t, serverDAO)

	userDAO := NewGenericRaftDAO[*config.User](
		IntegrationTestCluster.app.Logger,
		raftNode1,
		UserClusterID)
	assert.NotNil(t, userDAO)

	// farmDAO := NewGenericRaftDAO[*config.User](
	// 	IntegrationTestCluster.app.Logger,
	// 	raftNode1,
	// 	)
	// assert.NotNil(t, userDAO)

	// farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
	// 	Cluster.GetRaftNode1(), serverDAO, userDAO)
	// assert.NotNil(t, farmDAO)

	// deviceDAO := NewRaftDeviceConfigDAO(Cluster.app.Logger,
	// 	Cluster.GetRaftNode1(), farmDAO)
	// assert.NotNil(t, deviceDAO)

	// channelDAO := NewRaftChannelDAO(Cluster.app.Logger,
	// 	Cluster.GetRaftNode1(), farmDAO)
	// assert.NotNil(t, channelDAO)

	// org := createRaftTestOrganization(t, Cluster,
	// 	ClusterID, serverDAO, userDAO, farmDAO)

	//dstest.TestChannelCRUD(t, channelDAO, org)
}

// func TestChannelGetByDevice(t *testing.T) {

// 	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
// 		Cluster.GetRaftNode1(), ClusterID)

// 	orgDAO := NewRaftOrganizationDAO(Cluster.app.Logger,
// 		Cluster.GetRaftNode1(), OrganizationClusterID, serverDAO)
// 	assert.NotNil(t, orgDAO)

// 	userDAO := NewRaftUserDAO(Cluster.app.Logger,
// 		Cluster.GetRaftNode1(), UserClusterID)
// 	assert.NotNil(t, userDAO)

// 	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
// 		Cluster.GetRaftNode1(), serverDAO, userDAO)
// 	assert.NotNil(t, farmDAO)

// 	permissionDAO := NewRaftPermissionDAO(Cluster.app.Logger,
// 		orgDAO, farmDAO, userDAO)
// 	assert.NotNil(t, permissionDAO)

// 	deviceDAO := NewRaftDeviceConfigDAO(Cluster.app.Logger,
// 		Cluster.GetRaftNode1(), farmDAO)
// 	assert.NotNil(t, deviceDAO)

// 	channelDAO := NewRaftChannelDAO(Cluster.app.Logger,
// 		Cluster.GetRaftNode1(), farmDAO)
// 	assert.NotNil(t, channelDAO)

// 	org := createRaftTestOrganization(t, Cluster,
// 		ClusterID, serverDAO, userDAO, farmDAO)

// 	dstest.TestChannelGetByDevice(t, farmDAO, deviceDAO, channelDAO, permissionDAO, org)
// }
