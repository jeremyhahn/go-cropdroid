package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"

	"github.com/stretchr/testify/assert"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestOrganizationCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	orgDAO := NewOrganizationDAO(currentTest.logger,
		currentTest.gorm, currentTest.idGenerator)
	assert.NotNil(t, orgDAO)

	dstest.TestOrganizationCRUD(t, orgDAO)
}

func TestOrganizationGetAll(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.Organization{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_32BIT)
	orgDAO := NewOrganizationDAO(currentTest.logger, currentTest.gorm, idGenerator)
	assert.NotNil(t, orgDAO)

	dstest.TestOrganizationGetAll(t, orgDAO)
}

func TestOrganizationDelete(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.Organization{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_32BIT)
	orgDAO := NewOrganizationDAO(currentTest.logger, currentTest.gorm, idGenerator)
	assert.NotNil(t, orgDAO)

	// create first org
	testOrgName := "Test Org"
	testFarmName := "Test Farm"

	farm := config.NewFarm()
	farm.SetName(testFarmName)

	orgConfig := &config.Organization{
		Name:  testOrgName,
		Farms: []*config.Farm{farm}}

	err := orgDAO.Save(orgConfig)
	assert.Nil(t, err)
	assert.Equal(t, orgConfig.GetName(), testOrgName)

	// create second org
	testOrgName2 := "Test Org 2"
	testFarmName2 := "Test Org - Farm 1"

	farm2 := config.NewFarm()
	farm2.SetName(testFarmName2)
	orgConfig2 := &config.Organization{
		Name:  testOrgName2,
		Farms: []*config.Farm{farm2}}

	err = orgDAO.Save(orgConfig2)
	assert.Nil(t, err)

	// make sure orgs are returned fully hydrated
	orgs, err := orgDAO.GetAll(DEFAULT_CONSISTENCY_LEVEL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(orgs))
	assert.Equal(t, 1, len(orgs[0].GetFarms()))
	assert.Equal(t, 1, len(orgs[1].GetFarms()))

	err = orgDAO.Delete(orgConfig)
	assert.Nil(t, err)

	orgs, err = orgDAO.GetAll(DEFAULT_CONSISTENCY_LEVEL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(orgs))
}

func TestOrganizationEnchilada(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.DeviceSetting{})
	currentTest.gorm.AutoMigrate(&config.Metric{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Algorithm{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.License{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_32BIT)

	orgDAO := NewOrganizationDAO(currentTest.logger, currentTest.gorm, idGenerator)
	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	permissionDAO := NewPermissionDAO(currentTest.logger, currentTest.gorm)

	org := dstest.CreateTestOrganization(idGenerator)

	err := orgDAO.Save(org)
	assert.Nil(t, err)
	assert.NotNil(t, org.GetID())

	role := config.NewRole()
	role.SetName("admin")
	err = roleDAO.Save(role)
	assert.Nil(t, err)

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("test")

	err = userDAO.Save(user)
	assert.Nil(t, err)

	farmID := org.GetFarms()[0].GetID()
	permission := config.NewPermission()
	permission.SetFarmID(farmID)
	permission.SetOrgID(org.GetID())
	permission.SetUserID(user.GetID())
	permission.SetRoleID(role.GetID())
	err = permissionDAO.Save(permission)
	assert.Nil(t, err)

	allOrgs, err := orgDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allOrgs))

	persistedOrg := allOrgs[0]
	assert.Equal(t, org.GetName(), persistedOrg.GetName())

	farms := persistedOrg.GetFarms()
	assert.NotNil(t, farms)
	assert.Equal(t, 2, len(farms))
	assert.Equal(t, dstest.FARM1_NAME, farms[0].GetName())
	assert.Equal(t, "test", farms[0].GetMode())
	assert.Equal(t, dstest.FARM2_NAME, farms[1].GetName())
	assert.Equal(t, "test2", farms[1].GetMode())

	farm1 := farms[0]
	serverDevice, err := farm1.GetDevice(dstest.SERVER_TYPE)
	assert.Nil(t, err)
	assert.NotNil(t, serverDevice)

	device1, err := farm1.GetDevice(dstest.DEVICE1_TYPE)
	assert.Nil(t, err)
	assert.NotNil(t, device1)

	assert.Equal(t, 2, len(farm1.GetDevices()))
	assert.Equal(t, 10, len(serverDevice.GetSettings()))
	assert.Equal(t, 3, len(device1.GetSettings()))

	configEnableKey := "fakedevice.enable"
	configEnable := device1.GetSetting(configEnableKey)
	assert.NotNil(t, configEnable)

	//configEnable := devices[1].GetSettings()[0]
	assert.Equal(t, configEnableKey, configEnable.GetKey())
	assert.Equal(t, "true", configEnable.GetValue())
	assert.Equal(t, true, device1.IsEnabled())

	configNotifyKey := "fakedevice.notify"
	//configNotify := devices[1].GetSettings()[1]
	configNotify := device1.GetSetting(configNotifyKey)
	assert.Equal(t, configNotifyKey, configNotify.GetKey())
	assert.Equal(t, "false", configNotify.GetValue())
	assert.Equal(t, false, device1.IsNotify())

	configUriKey := "fakedevice.uri"
	//configURI := devices[1].GetSettings()[2]
	configURI := device1.GetSetting(configUriKey)
	assert.Equal(t, configUriKey, configURI.GetKey())
	assert.Equal(t, "http://mydevice.mydomain.com", configURI.GetValue())
	assert.Equal(t, "http://mydevice.mydomain.com", device1.GetURI())

	// dstest.TestOrganizationEnchilada(t, orgDAO,
	// 	roleDAO, userDAO, permissionDAO, org)
}
