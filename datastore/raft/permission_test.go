//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/util"

	"github.com/stretchr/testify/assert"
)

func TestUserRoleRelationship(t *testing.T) {

	consistencyLevel := common.CONSISTENCY_LOCAL
	raftNode1 := IntegrationTestCluster.GetRaftNode1()

	userDAO := NewGenericRaftDAO[*config.UserStruct](
		IntegrationTestCluster.app.Logger,
		raftNode1,
		UserClusterID)
	assert.NotNil(t, userDAO)
	userDAO.StartLocalCluster(IntegrationTestCluster, true)

	roleName := "test"
	role := config.NewRole()
	role.SetName(roleName)

	userEmail := "root@localhost"
	user := config.NewUser()
	user.SetEmail(userEmail)
	user.SetPassword("$ecret")
	user.SetRoles([]*config.RoleStruct{role})

	err := userDAO.Save(user)
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

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	serverDAO := IntegrationTestCluster.serverDAO

	org, orgDAO, _, _ := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)

	userDAO := NewGenericRaftDAO[*config.UserStruct](
		IntegrationTestCluster.app.Logger,
		raftNode1,
		UserClusterID)
	assert.NotNil(t, userDAO)

	roleDAO := NewRaftRoleDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		RoleClusterID)
	assert.NotNil(t, roleDAO)

	role := config.NewRole()
	role.SetName("test")
	assert.Nil(t, roleDAO.Save(role))
	err := roleDAO.Save(role)
	assert.Nil(t, err)

	userEmail := "root@localhost"
	user := config.NewUser()
	user.SetEmail(userEmail)
	user.SetPassword("$ecret")
	user.SetRoles([]*config.RoleStruct{role})
	err = userDAO.Save(user)
	assert.Nil(t, err)

	farmName := "Test Farm"
	farm := config.NewFarm()
	farm.SetID(raftNode1.GetParams().IdGenerator.NewFarmID(org.ID, farmName))
	farm.SetName(farmName)
	farm.SetMode("test")
	farm.SetInterval(60)
	farm.SetUsers([]*config.UserStruct{user})

	// Create the new farm cluster so it can be saved
	farmDAO := NewRaftFarmConfigDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		serverDAO,
		userDAO)
	farmDAO.StartLocalCluster(IntegrationTestCluster, farm.ID, true)

	err = farmDAO.Save(farm)
	assert.Nil(t, err)

	permissionDAO := NewRaftPermissionDAO(
		IntegrationTestCluster.app.Logger,
		orgDAO,
		farmDAO,
		userDAO)

	permission := &config.PermissionStruct{
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
	userFarms, err := farmDAO.GetByUserID(user.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(userFarms))
	assert.Equal(t, farm.ID, userFarms[0].ID)
}

func TestGetOrganizations(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	serverDAO := IntegrationTestCluster.serverDAO

	_, orgDAO, _, _ := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)

	userDAO := NewGenericRaftDAO[*config.UserStruct](
		IntegrationTestCluster.app.Logger,
		raftNode1,
		UserClusterID)
	assert.NotNil(t, userDAO)

	roleDAO := NewRaftRoleDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		RoleClusterID)
	assert.NotNil(t, roleDAO)

	roleName := "test"
	role := config.NewRole()
	role.SetID(idGenerator.NewStringID(roleName))
	role.SetName("test")

	err := roleDAO.Save(role)
	assert.Nil(t, err)

	userEmail := "root@localhost"
	user := config.NewUser()
	user.SetID(idGenerator.NewStringID(userEmail))
	user.SetEmail(userEmail)
	user.SetPassword("$ecret")
	user.SetRoles([]*config.RoleStruct{role})

	err = userDAO.Save(user)
	assert.Nil(t, err)

	// create first org
	testOrgName := "Test Org"
	testFarmName := "Test Farm"

	farmConfig := config.NewFarm()
	farmConfig.SetID(idGenerator.NewStringID(testFarmName))
	farmConfig.SetName(testFarmName)

	farmDAO := NewRaftFarmConfigDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		serverDAO,
		userDAO)
	farmDAO.StartLocalCluster(IntegrationTestCluster, farmConfig.ID, true)

	permissionDAO := NewRaftPermissionDAO(
		IntegrationTestCluster.app.Logger,
		orgDAO,
		farmDAO,
		userDAO)

	orgConfig := &config.OrganizationStruct{
		ID:    idGenerator.NewStringID(testOrgName),
		Name:  testOrgName,
		Farms: []*config.FarmStruct{farmConfig}}

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

	// Org 2
	orgConfig2 := &config.OrganizationStruct{
		ID:    idGenerator.NewStringID(testOrgName2),
		Name:  testOrgName2,
		Farms: []*config.FarmStruct{farmConfig2}}

	err = orgDAO.Save(orgConfig2)
	assert.Nil(t, err)

	// make sure orgs are returned fully hydrated
	page1, err := orgDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(page1.Entities))
	assert.Equal(t, 2, len(page1.Entities[0].GetFarms()))
	assert.Equal(t, 1, len(page1.Entities[1].GetFarms()))

	permissionDAO.Save(&config.PermissionStruct{
		OrganizationID: orgConfig.ID,
		FarmID:         farmConfig.ID,
		UserID:         user.ID,
		RoleID:         role.ID})

	permissionDAO.Save(&config.PermissionStruct{
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
