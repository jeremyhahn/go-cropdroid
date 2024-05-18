//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/util"

	"github.com/stretchr/testify/assert"
)

func TestUserRoleRelationship(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	consistencyLevel := common.CONSISTENCY_LOCAL

	err := Cluster.CreateOrganizationCluster()
	assert.Nil(t, err)

	err = Cluster.CreateUserCluster()
	assert.Nil(t, err)

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)
	assert.NotNil(t, userDAO)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)
	assert.NotNil(t, farmDAO)

	roleName := "test"
	role := config.NewRole()
	role.SetID(idGenerator.NewStringID(roleName))
	role.SetName(roleName)

	userEmail := "root@localhost"
	user := config.NewUser()
	user.SetID(UserClusterID)
	user.SetEmail(userEmail)
	user.SetPassword("$ecret")
	user.SetRoles([]*config.Role{role})

	err = userDAO.Save(user)
	assert.Nil(t, err)

	persisted, err := userDAO.Get(user.ID, consistencyLevel)
	assert.Nil(t, err)

	assert.Equal(t, persisted.ID, user.ID)
	assert.Equal(t, "root@localhost", persisted.GetEmail())
	assert.Equal(t, "$ecret", persisted.GetPassword())
	assert.Equal(t, 1, len(persisted.GetRoles()))
	assert.Equal(t, "test", persisted.GetRoles()[0].GetName())
}

func TestPermissions(t *testing.T) {

	//idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	consistencyLevel := common.CONSISTENCY_LOCAL

	err := Cluster.CreateOrganizationCluster()
	assert.Nil(t, err)

	err = Cluster.CreateUserCluster()
	assert.Nil(t, err)

	err = Cluster.CreateRoleCluster()
	assert.Nil(t, err)

	err = Cluster.CreateFarmConfigCluster(FarmConfigClusterID)
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

	server := config.NewServer()
	server.SetID(ClusterID)

	err = serverDAO.Save(server)
	assert.Nil(t, err)

	//roleName := "test"
	role := config.NewRole()
	role.SetID(RoleClusterID)
	role.SetName("test")

	err = roleDAO.Save(role)
	assert.Nil(t, err)

	userEmail := "root@localhost"
	user := config.NewUser()
	user.SetID(UserClusterID)
	user.SetEmail(userEmail)
	user.SetPassword("$ecret")
	user.SetRoles([]*config.Role{role})

	err = userDAO.Save(user)
	assert.Nil(t, err)

	farmName := "Test Farm"
	farm := config.NewFarm()
	farm.SetID(FarmConfigClusterID)
	farm.SetName(farmName)
	farm.SetMode("test")
	farm.SetInterval(60)
	farm.SetUsers([]*config.User{user})

	err = farmDAO.Save(farm)
	assert.Nil(t, err)

	permission := &config.Permission{
		OrganizationID: 0,
		FarmID:         farm.ID,
		UserID:         user.ID,
		RoleID:         role.ID}

	permissionDAO.Save(permission)

	// Verify Get(id)
	persisted, err := farmDAO.Get(farm.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, farm.ID, persisted.ID)

	// Verify GetByOrgAndUserID(orgID, userID)
	userFarms, err := farmDAO.GetByUserID(user.ID, consistencyLevel)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(userFarms))
	assert.Equal(t, farm.ID, userFarms[0].ID)
}

func TestGetOrganizations(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)

	err := Cluster.CreateOrganizationCluster()
	assert.Nil(t, err)

	err = Cluster.CreateUserCluster()
	assert.Nil(t, err)

	err = Cluster.CreateRoleCluster()
	assert.Nil(t, err)

	err = Cluster.CreateFarmConfigCluster(FarmConfigClusterID)
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

	server := config.NewServer()
	server.SetID(ClusterID)

	err = serverDAO.Save(server)
	assert.Nil(t, err)

	roleName := "test"
	role := config.NewRole()
	role.SetID(idGenerator.NewStringID(roleName))
	role.SetName("test")

	err = roleDAO.Save(role)
	assert.Nil(t, err)

	userEmail := "root@localhost"
	user := config.NewUser()
	user.SetID(idGenerator.NewStringID(userEmail))
	user.SetEmail(userEmail)
	user.SetPassword("$ecret")
	//user.SetRoles([]config.Role{*role})

	err = userDAO.Save(user)
	assert.Nil(t, err)

	// create first org
	testOrgName := "Test Org"
	testFarmName := "Test Farm"

	farmConfig := config.NewFarm()
	farmConfig.SetID(idGenerator.NewStringID(testFarmName))
	farmConfig.SetName(testFarmName)

	Cluster.CreateFarmConfigCluster(farmConfig.ID)

	orgConfig := &config.Organization{
		ID:    idGenerator.NewStringID(testOrgName),
		Name:  testOrgName,
		Farms: []*config.Farm{farmConfig}}

	err = orgDAO.Save(orgConfig)
	assert.Nil(t, err)

	// create second org
	testOrgName2 := "Test Org 2"
	testFarmName2 := "Test Org - Farm 1"

	farmConfig2 := config.NewFarm()
	farmConfig2.SetID(idGenerator.NewStringID(testFarmName))
	farmConfig2.SetName(testFarmName2)

	err = farmDAO.Save(farmConfig)
	assert.Nil(t, err)

	Cluster.CreateFarmConfigCluster(farmConfig2.ID)

	orgConfig2 := &config.Organization{
		ID:    idGenerator.NewStringID(testOrgName2),
		Name:  testOrgName2,
		Farms: []*config.Farm{farmConfig2}}

	err = orgDAO.Save(orgConfig2)
	assert.Nil(t, err)

	// make sure orgs are returned fully hydrated
	page1, err := orgDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(page1.Entities))
	assert.Equal(t, 1, len(page1.Entities[0].GetFarms()))
	assert.Equal(t, 1, len(page1.Entities[1].GetFarms()))

	permissionDAO.Save(&config.Permission{
		OrganizationID: orgConfig.ID,
		FarmID:         farmConfig.ID,
		UserID:         user.ID,
		RoleID:         role.ID})

	permissionDAO.Save(&config.Permission{
		OrganizationID: orgConfig2.ID,
		FarmID:         farmConfig2.ID,
		UserID:         user.ID,
		RoleID:         role.ID})

	persistedOrgs, err := permissionDAO.GetOrganizations(
		user.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(persistedOrgs))
	assert.Equal(t, persistedOrgs[0].ID, orgConfig.ID)
	assert.Equal(t, persistedOrgs[1].ID, orgConfig2.ID)
}
