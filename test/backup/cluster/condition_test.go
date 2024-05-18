//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestConditionCRUD(t *testing.T) {

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)

	conditionDAO := NewRaftConditionDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmDAO)

	org := createRaftTestOrganization(t, Cluster,
		ClusterID, serverDAO, userDAO, farmDAO)

	dstest.TestConditionCRUD(t, conditionDAO, org)
}
