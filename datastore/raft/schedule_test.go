//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestScheduleCRUD(t *testing.T) {

	org, _, farmDAO, _ := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)

	scheduleDAO := NewRaftScheduleDAO(
		IntegrationTestCluster.app.Logger,
		IntegrationTestCluster.GetRaftNode1(),
		farmDAO)

	dstest.TestScheduleCRUD(t, scheduleDAO, org)
}
