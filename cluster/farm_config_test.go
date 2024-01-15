//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestFarmAssociations(t *testing.T) {

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)

	org := createRaftTestOrganization(t, Cluster,
		ClusterID, serverDAO, userDAO, farmDAO)
	farm1 := org.GetFarms()[0]

	dstest.TestFarmAssociations(t, Cluster.app.IdGenerator,
		farmDAO, farm1)
}

func TestFarmGetByIds(t *testing.T) {

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)

	org := createRaftTestOrganization(t, Cluster, ClusterID,
		serverDAO, userDAO, farmDAO)
	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	dstest.TestFarmGetByIds(t, farmDAO, farm1, farm2)
}

func TestFarmGetAll(t *testing.T) {

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)

	org := createRaftTestOrganization(t, Cluster, ClusterID,
		serverDAO, userDAO, farmDAO)
	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	dstest.TestFarmGetAll(t, farmDAO, farm1, farm2)
}

func TestFarmGet(t *testing.T) {

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)

	org := createRaftTestOrganization(t, Cluster, ClusterID,
		serverDAO, userDAO, farmDAO)
	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	dstest.TestFarmGet(t, farmDAO, farm1, farm2)
}
