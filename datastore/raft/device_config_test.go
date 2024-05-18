//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCRUD(t *testing.T) {

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

	org, _, farmDAO, _ := createRaftTestOrganization(
		t,
		IntegrationTestCluster,
		ClusterID)
	farm := org.GetFarms()[0]

	deviceDAO := NewRaftDeviceConfigDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)
	assert.NotNil(t, deviceDAO)

	dstest.TestDeviceCRUD(t, deviceDAO, farm)
}
