//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func createTestOrganizationDAO(t *testing.T) dao.OrganizationDAO {
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

	orgDAO := NewRaftOrganizationDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		OrganizationClusterID,
		serverDAO)
	assert.NotNil(t, orgDAO)
	orgDAO.StartLocalCluster(IntegrationTestCluster, true)

	return orgDAO
}

func TestOrganizationCRUD(t *testing.T) {
	orgDAO := createTestOrganizationDAO(t)
	dstest.TestOrganizationCRUD(t, orgDAO)
}

func TestOrganizationGetPage(t *testing.T) {
	orgDAO := createTestOrganizationDAO(t)
	dstest.TestOrganizationGetPage(t, orgDAO)
}

func TestOrganizationDelete(t *testing.T) {
	orgDAO := createTestOrganizationDAO(t)
	dstest.TestOrganizationDelete(t, orgDAO)
}

// func TestOrganizationEnchilada(t *testing.T) {

// 	orgDAO := createTestOrganizationDAO(t)

// 	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
// 		Cluster.GetRaftNode1(), serverDAO, userDAO)
// 	assert.NotNil(t, farmDAO)

// 	permissionDAO := NewRaftPermissionDAO(Cluster.app.Logger,
// 		orgDAO, farmDAO, userDAO)

// 	roleDAO := NewRaftRoleDAO(Cluster.app.Logger,
// 		Cluster.GetRaftNode1(), RoleClusterID)
// 	assert.NotNil(t, roleDAO)

// 	org := createRaftTestOrganization(t, Cluster,
// 		ClusterID, serverDAO, userDAO, farmDAO)

// 		dstest.TestOrganizationEnchilada(t, orgDAO, roleDAO,
// 			userDAO, permissionDAO, org)
// 	}
// }

func createRaftTestOrganization(t *testing.T, cluster *LocalCluster,
	clusterID uint64) (*config.Organization, RaftOrganizationDAO, RaftFarmConfigDAO, RaftFarmConfigDAO) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	assert.NotNil(t, raftNode1)

	serverDAO := IntegrationTestCluster.serverDAO

	orgDAO := NewRaftOrganizationDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		OrganizationClusterID,
		serverDAO)
	assert.NotNil(t, orgDAO)
	orgDAO.StartLocalCluster(IntegrationTestCluster, false)

	userDAO := NewGenericRaftDAO[*config.User](
		IntegrationTestCluster.app.Logger,
		raftNode1,
		UserClusterID)
	assert.NotNil(t, userDAO)
	userDAO.StartLocalCluster(cluster, false)

	roleDAO := NewRaftRoleDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		RoleClusterID)
	assert.NotNil(t, roleDAO)
	roleDAO.StartLocalCluster(IntegrationTestCluster, false)

	orgDAO.WaitForClusterReady()
	userDAO.WaitForClusterReady()
	roleDAO.WaitForClusterReady()

	org := dstest.CreateTestOrganization(cluster.app.IdGenerator)
	assert.Nil(t, orgDAO.Save(org))

	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	// Save the role so it gets an ID
	assert.Nil(t, roleDAO.Save(farm1.GetUsers()[0].GetRoles()[0]))
	assert.Nil(t, roleDAO.Save(farm2.GetUsers()[0].GetRoles()[0]))

	// Save the user so it can be looked up by FarmConfigDAO.Save
	// when the user gets looked up to save it's FarmRefs.
	assert.Nil(t, userDAO.Save(farm1.GetUsers()[0]))
	assert.Nil(t, userDAO.Save(farm2.GetUsers()[0]))

	serverDAO.Save(&config.Server{
		ID: clusterID,
		// OrganizationRefs: []uint64{
		// 	org.GetID(),
		// },
		// FarmDAO.Save adds the farm refs
		// FarmRefs: []uint64{
		// 	farm1.GetID(),
		// 	farm2.GetID(),
		// },
	})

	farm1DAO := NewRaftFarmConfigDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		serverDAO,
		userDAO)
	farm1DAO.StartLocalCluster(IntegrationTestCluster, farm1.ID, true)

	farm2DAO := NewRaftFarmConfigDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		serverDAO,
		userDAO)
	farm1DAO.StartLocalCluster(IntegrationTestCluster, farm2.ID, true)

	err := farm1DAO.Save(farm1)
	assert.Nil(t, err)

	err = farm2DAO.Save(farm2)
	assert.Nil(t, err)

	return org, orgDAO, farm1DAO, farm2DAO
}
