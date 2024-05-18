//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
	"github.com/stretchr/testify/assert"
)

func TestConditionCRUD(t *testing.T) {

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

	conditionDAO := NewRaftConditionDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)

	dstest.TestConditionCRUD(t, conditionDAO, org)
}
