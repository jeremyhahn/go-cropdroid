//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/stretchr/testify/assert"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestOrganizationCRUD(t *testing.T) {

	err := Cluster.CreateOrganizationCluster()
	assert.Nil(t, err)

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	orgDAO := NewRaftOrganizationDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), OrganizationClusterID, serverDAO)
	assert.NotNil(t, orgDAO)

	dstest.TestOrganizationCRUD(t, orgDAO)
}

func TestOrganizationGetPage(t *testing.T) {

	err := Cluster.CreateOrganizationCluster()
	assert.Nil(t, err)

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	orgDAO := NewRaftOrganizationDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), OrganizationClusterID, serverDAO)
	assert.NotNil(t, orgDAO)

	dstest.TestOrganizationGetPage(t, orgDAO)
}

func TestOrganizationDelete(t *testing.T) {

	err := Cluster.CreateOrganizationCluster()
	assert.Nil(t, err)

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	orgDAO := NewRaftOrganizationDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), OrganizationClusterID, serverDAO)
	assert.NotNil(t, orgDAO)

	dstest.TestOrganizationDelete(t, orgDAO)
}

func TestOrganizationEnchilada(t *testing.T) {

	err := Cluster.CreateRoleCluster()
	assert.Nil(t, err)

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

	roleDAO := NewRaftRoleDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), RoleClusterID)
	assert.NotNil(t, roleDAO)

	org := createRaftTestOrganization(t, Cluster,
		ClusterID, serverDAO, userDAO, farmDAO)

	dstest.TestOrganizationEnchilada(t, orgDAO, roleDAO,
		userDAO, permissionDAO, org)
}

func createRaftTestOrganization(t *testing.T, cluster *TestCluster,
	clusterID uint64, serverDAO ServerDAO, userDAO dao.UserDAO,
	farmDAO dao.FarmDAO) *config.Organization {

	assert.Nil(t, Cluster.CreateUserCluster())

	org := dstest.CreateTestOrganization(cluster.app.IdGenerator)

	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	// Save the user so it can be looked up by FarmConfigDAO.Save
	// when the user gets looked up to save it's FarmRefs.
	assert.Nil(t, userDAO.Save(farm1.GetUsers()[0]))

	serverDAO.Save(&config.Server{
		ID: clusterID,
		// OrganizationRefs: []uint64{
		// 	org.ID,
		// },
		// FarmDAO.Save adds the farm refs
		// FarmRefs: []uint64{
		// 	farm1.ID,
		// 	farm2.ID,
		// },
	})

	if err := cluster.CreateOrganizationCluster(); err != nil {
		assert.Fail(t, err.Error())
	}

	if err := cluster.CreateFarmConfigCluster(farm1.ID); err != nil {
		assert.Fail(t, err.Error())
	}
	if err := farmDAO.Save(farm1); err != nil {
		assert.Fail(t, err.Error())
	}

	if err := cluster.CreateFarmConfigCluster(farm2.ID); err != nil {
		assert.Fail(t, err.Error())
	}
	if err := farmDAO.Save(farm2); err != nil {
		assert.Fail(t, err.Error())
	}

	return org
}
