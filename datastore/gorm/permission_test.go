package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/util"

	"github.com/stretchr/testify/assert"
)

func TestUserRoleRelationship(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_SQLITE)
	email := "root@localhost"
	userID := idGenerator.NewStringID(email)

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm,
		currentTest.idGenerator)
	assert.NotNil(t, farmDAO)

	role := config.NewRole()
	role.SetName("test")

	user := config.NewUser()
	user.SetID(userID)
	user.SetEmail(email)
	user.SetPassword("$ecret")
	user.SetRoles([]*config.RoleStruct{role})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	err := userDAO.Save(user)
	assert.Nil(t, err)

	persisted, err := userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)

	assert.Equal(t, persisted.ID, user.ID)
	assert.Equal(t, "root@localhost", persisted.GetEmail())
	assert.Equal(t, "$ecret", persisted.GetPassword())
	assert.Equal(t, 1, len(persisted.GetRoles()))
	assert.Equal(t, "test", persisted.GetRoles()[0].GetName())
}

func TestPermissions(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})
	currentTest.gorm.AutoMigrate(&config.OrganizationStruct{})
	currentTest.gorm.AutoMigrate(&config.FarmStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceStruct{})
	currentTest.gorm.AutoMigrate(&config.ChannelStruct{})
	currentTest.gorm.AutoMigrate(&config.ConditionStruct{})
	currentTest.gorm.AutoMigrate(&config.ScheduleStruct{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStruct{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStepStruct{})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm,
		currentTest.idGenerator)
	permissionDAO := NewPermissionDAO(currentTest.logger, currentTest.gorm)

	role := config.NewRole()
	role.SetName("test")

	err := roleDAO.Save(role)
	assert.Nil(t, err)

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("$ecret")
	//user.SetRoles([]config.Role{*role})

	err = userDAO.Save(user)
	assert.Nil(t, err)

	farm := config.NewFarm()
	farm.SetName("Test Farm")
	farm.SetMode("test")
	farm.SetInterval(60)
	//farm.SetUsers([]config.User{*user})

	err = farmDAO.Save(farm)
	assert.Nil(t, err)

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

	// Verify GetByUserID(userID)
	userFarms, err := farmDAO.GetByUserID(user.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(userFarms))
	assert.Equal(t, farm.ID, userFarms[0].ID)
}

func TestGetOrganizations(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})
	currentTest.gorm.AutoMigrate(&config.OrganizationStruct{})
	currentTest.gorm.AutoMigrate(&config.FarmStruct{})
	currentTest.gorm.AutoMigrate(&config.DeviceStruct{})
	currentTest.gorm.AutoMigrate(&config.ChannelStruct{})
	currentTest.gorm.AutoMigrate(&config.ConditionStruct{})
	currentTest.gorm.AutoMigrate(&config.ScheduleStruct{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStruct{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStepStruct{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_32BIT)
	orgDAO := NewOrganizationDAO(currentTest.logger, currentTest.gorm, idGenerator)
	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	permissionDAO := NewPermissionDAO(currentTest.logger, currentTest.gorm)
	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm,
		currentTest.idGenerator)

	assert.NotNil(t, orgDAO)
	assert.NotNil(t, userDAO)
	assert.NotNil(t, roleDAO)
	assert.NotNil(t, permissionDAO)
	assert.NotNil(t, farmDAO)

	role := config.NewRole()
	role.SetName("test")

	err := roleDAO.Save(role)
	assert.Nil(t, err)

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("$ecret")
	//user.SetRoles([]config.Role{*role})

	err = userDAO.Save(user)
	assert.Nil(t, err)

	// create first org
	testOrgName := "Test Org"
	testFarmName := "Test Farm"

	farmConfig := config.NewFarm()
	farmConfig.SetName(testFarmName)

	orgConfig := &config.OrganizationStruct{
		Name:  testOrgName,
		Farms: []*config.FarmStruct{farmConfig}}

	err = orgDAO.Save(orgConfig)
	assert.Nil(t, err)
	assert.Equal(t, orgConfig.GetName(), testOrgName)

	// create second org
	testOrgName2 := "Test Org 2"
	testFarmName2 := "Test Org - Farm 1"

	farmConfig2 := config.NewFarm()
	farmConfig2.SetName(testFarmName2)
	orgConfig2 := &config.OrganizationStruct{
		Name:  testOrgName2,
		Farms: []*config.FarmStruct{farmConfig2}}

	err = orgDAO.Save(orgConfig2)
	assert.Nil(t, err)

	// make sure orgs are returned fully hydrated
	page1, err := orgDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(page1.Entities))
	assert.Equal(t, 1, len(page1.Entities[0].GetFarms()))
	assert.Equal(t, 1, len(page1.Entities[1].GetFarms()))

	permissionDAO.Save(&config.PermissionStruct{
		OrganizationID: orgConfig.ID,
		FarmID:         0,
		UserID:         user.ID,
		RoleID:         role.ID})

	permissionDAO.Save(&config.PermissionStruct{
		OrganizationID: orgConfig2.ID,
		FarmID:         0,
		UserID:         user.ID,
		RoleID:         role.ID})

	persistedOrgs, err := permissionDAO.GetOrganizations(user.ID,
		common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(persistedOrgs))
	assert.Equal(t, persistedOrgs[0].ID, orgConfig.ID)
	assert.Equal(t, persistedOrgs[1].ID, orgConfig2.ID)
}
