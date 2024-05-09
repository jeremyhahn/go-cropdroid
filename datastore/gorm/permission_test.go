package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"

	"github.com/stretchr/testify/assert"
)

func TestUserRoleRelationship(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

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
	user.SetRoles([]*config.Role{role})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	err := userDAO.Save(user)
	assert.Nil(t, err)

	persisted, err := userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)

	assert.Equal(t, persisted.GetID(), user.GetID())
	assert.Equal(t, "root@localhost", persisted.GetEmail())
	assert.Equal(t, "$ecret", persisted.GetPassword())
	assert.Equal(t, 1, len(persisted.GetRoles()))
	assert.Equal(t, "test", persisted.GetRoles()[0].GetName())
}

func TestPermissions(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})
	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

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

	permission := &config.Permission{
		OrganizationID: 0,
		FarmID:         farm.GetID(),
		UserID:         user.GetID(),
		RoleID:         role.GetID()}

	permissionDAO.Save(permission)

	// Verify Get(id)
	persisted, err := farmDAO.Get(farm.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, farm.GetID(), persisted.GetID())

	// Verify GetByUserID(userID)
	userFarms, err := farmDAO.GetByUserID(user.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(userFarms))
	assert.Equal(t, farm.GetID(), userFarms[0].GetID())
}

func TestGetOrganizations(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})
	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

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

	orgConfig := &config.Organization{
		Name:  testOrgName,
		Farms: []*config.Farm{farmConfig}}

	err = orgDAO.Save(orgConfig)
	assert.Nil(t, err)
	assert.Equal(t, orgConfig.GetName(), testOrgName)

	// create second org
	testOrgName2 := "Test Org 2"
	testFarmName2 := "Test Org - Farm 1"

	farmConfig2 := config.NewFarm()
	farmConfig2.SetName(testFarmName2)
	orgConfig2 := &config.Organization{
		Name:  testOrgName2,
		Farms: []*config.Farm{farmConfig2}}

	err = orgDAO.Save(orgConfig2)
	assert.Nil(t, err)

	// make sure orgs are returned fully hydrated
	orgs, err := orgDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(orgs))
	assert.Equal(t, 1, len(orgs[0].GetFarms()))
	assert.Equal(t, 1, len(orgs[1].GetFarms()))

	permissionDAO.Save(&config.Permission{
		OrganizationID: orgConfig.GetID(),
		FarmID:         0,
		UserID:         user.GetID(),
		RoleID:         role.GetID()})

	permissionDAO.Save(&config.Permission{
		OrganizationID: orgConfig2.GetID(),
		FarmID:         0,
		UserID:         user.GetID(),
		RoleID:         role.GetID()})

	persistedOrgs, err := permissionDAO.GetOrganizations(user.GetID(),
		common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(persistedOrgs))
	assert.Equal(t, persistedOrgs[0].GetID(), orgConfig.GetID())
	assert.Equal(t, persistedOrgs[1].GetID(), orgConfig2.GetID())
}
