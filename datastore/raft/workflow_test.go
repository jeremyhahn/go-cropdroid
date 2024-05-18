//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestWorkflowCRUD(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	org, _, farmDAO, _ := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)

	workflowDAO := NewRaftWorkflowDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)

	dstest.TestWorkflowCRUD(t, workflowDAO, org)
}
