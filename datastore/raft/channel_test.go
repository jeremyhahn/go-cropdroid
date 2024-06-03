//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"

	"github.com/stretchr/testify/assert"
)

func TestChannelCRUD(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	systemRaftClusterID := raftNode1.GetParams().RaftOptions.SystemClusterID

	org, _, farmDAO, _ := createRaftTestOrganization(t, IntegrationTestCluster, systemRaftClusterID)

	deviceDAO := NewRaftDeviceConfigDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)
	assert.NotNil(t, deviceDAO)

	channelDAO := NewRaftChannelDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)
	assert.NotNil(t, channelDAO)

	dstest.TestChannelCRUD(t, channelDAO, org)
}

func TestChannelGetByDevice(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	systemRaftClusterID := raftNode1.GetParams().RaftOptions.SystemClusterID

	org, orgDAO, farmDAO, _ := createRaftTestOrganization(t, IntegrationTestCluster, systemRaftClusterID)

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

	channelDAO := NewRaftChannelDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)
	assert.NotNil(t, channelDAO)

	permissionDAO := NewRaftPermissionDAO(
		IntegrationTestCluster.app.Logger,
		orgDAO,
		farmDAO,
		userDAO)
	assert.NotNil(t, permissionDAO)

	dstest.TestChannelGetByDevice(t, farmDAO, deviceDAO, channelDAO, permissionDAO, org)
}
