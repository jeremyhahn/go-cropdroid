package gorm

import (
	"fmt"
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	orgDAO := NewOrganizationDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, orgDAO)

	org, err := orgDAO.First()
	assert.NotNil(t, err)
	assert.Nil(t, org)

	err = orgDAO.Create(&config.Organization{
		Name: "Test Org"})
	assert.Nil(t, err)

	assert.NotNil(t, orgDAO)

	org, err = orgDAO.First()
	assert.Nil(t, err)
	assert.Equal(t, "Test Org", org.GetName())

	currentTest.Cleanup()
}

func TestOrganizationGetByUserID(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.DeviceConfigItem{})
	currentTest.gorm.AutoMigrate(&config.Organization{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Metric{})

	testOrgName := "Test Org"
	testMode := "virtual"
	testFarmName := "Test Farm"

	farmConfig := config.NewFarm()
	farmConfig.SetDevices([]config.Device{
		{
			Type: "server",
			Configs: []config.DeviceConfigItem{
				{
					Key:   "name",
					Value: testFarmName},
				{
					Key:   "interval",
					Value: "55"},
				{
					Key:   "mode",
					Value: testMode},
				{
					Key:   "timezone",
					Value: "America/New_York"},
				{
					Key:   "smtp.enable",
					Value: "true"},
				{
					Key:   "smtp.host",
					Value: "127.0.0.1"},
				{
					Key:   "smtp.port",
					Value: "587"},
				{
					Key:   "smtp.username",
					Value: "foo"},
				{
					Key:   "smtp.password",
					Value: "bar"},
				{
					Key:   "smtp.recipient",
					Value: "user@domain.com"},
			}}})

	orgConfig := &config.Organization{
		Name:  testOrgName,
		Farms: []config.Farm{*farmConfig}}

	orgDAO := NewOrganizationDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, orgDAO)

	err := orgDAO.Create(orgConfig)
	assert.Nil(t, err)

	assert.Equal(t, orgConfig.GetID() > 0, true)

	persistedOrg, err := orgDAO.First()
	assert.Nil(t, err)
	assert.Equal(t, orgConfig.GetID(), persistedOrg.GetID())
	assert.Equal(t, testOrgName, persistedOrg.GetName())

	assert.Equal(t, 1, len(persistedOrg.GetFarms()))

	farm := persistedOrg.GetFarms()[0]
	assert.Equal(t, testFarmName, farm.GetName())
	assert.Equal(t, testMode, farm.GetMode())

	currentTest.Cleanup()
}

func TestOrganizationGetAll(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.Organization{})

	orgDAO := NewOrganizationDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, orgDAO)

	// create first org
	testOrgName := "Test Org"
	testFarmName := "Test Farm"

	farmConfig := config.NewFarm()
	farmConfig.SetName(testFarmName)

	orgConfig := &config.Organization{
		Name:  testOrgName,
		Farms: []config.Farm{*farmConfig}}

	err := orgDAO.Create(orgConfig)
	assert.Nil(t, err)
	assert.Equal(t, orgConfig.GetName(), testOrgName)

	// create second org
	testOrgName2 := "Test Org 2"
	testFarmName2 := "Test Org - Farm 1"

	farmConfig2 := config.NewFarm()
	farmConfig2.SetName(testFarmName2)
	orgConfig2 := &config.Organization{
		Name:  testOrgName2,
		Farms: []config.Farm{*farmConfig2}}

	err = orgDAO.Create(orgConfig2)
	assert.Nil(t, err)

	// make sure orgs are returned fully hydrated
	orgs, err := orgDAO.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(orgs))
	assert.Equal(t, 1, len(orgs[0].GetFarms()))
	assert.Equal(t, 1, len(orgs[1].GetFarms()))

	fmt.Printf("persisted farms: %+v\n", orgs[0].GetFarms())

	currentTest.Cleanup()
}

func TestOrganizationEnchilada(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.DeviceConfigItem{})
	currentTest.gorm.AutoMigrate(&config.Metric{})
	currentTest.gorm.AutoMigrate(&config.Condition{})
	currentTest.gorm.AutoMigrate(&config.Schedule{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.Algorithm{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.License{})
	currentTest.gorm.AutoMigrate(&config.Organization{})

	org := createTestOrganization()
	currentTest.logger.Infof("Org: %+v", org)

	orgDAO := NewOrganizationDAO(currentTest.logger, currentTest.gorm)
	err := orgDAO.Create(org)
	assert.Nil(t, err)
	assert.NotNil(t, org.GetID())

	role := config.NewRole()
	role.SetName("admin")
	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	err = roleDAO.Save(role)
	assert.Nil(t, err)

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("test")

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	err = userDAO.Create(user)
	assert.Nil(t, err)

	// Gorm doesn't handle multiple many-to-many fields in one entity,
	// create the user/role/org associations manually
	err = orgDAO.CreateUserRole(org, user, role)
	assert.Nil(t, err)

	allOrgs, err := orgDAO.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allOrgs))

	currentTest.logger.Infof("Persisted Organization: %+v", allOrgs[0])

	persistedOrg := allOrgs[0]
	assert.Equal(t, "Test Org", persistedOrg.GetName())

	farms := persistedOrg.GetFarms()
	assert.NotNil(t, farms)
	assert.Equal(t, 1, len(farms))
	assert.Equal(t, "Fake Farm", farms[0].GetName())
	assert.Equal(t, "test", farms[0].GetMode())
	assert.Equal(t, 58, farms[0].GetInterval())

	devices := farms[0].GetDevices()
	assert.Equal(t, 2, len(devices))
	assert.Equal(t, 10, len(devices[0].GetConfigs()))
	assert.Equal(t, 3, len(devices[1].GetConfigs()))

	configEnable := devices[1].GetConfigs()[0]
	assert.Equal(t, "fakedevice.enable", configEnable.GetKey())
	assert.Equal(t, "true", configEnable.GetValue())
	assert.Equal(t, true, devices[1].IsEnabled())

	configNotify := devices[1].GetConfigs()[1]
	assert.Equal(t, "fakedevice.notify", configNotify.GetKey())
	assert.Equal(t, "false", configNotify.GetValue())
	assert.Equal(t, false, devices[1].IsNotify())

	configURI := devices[1].GetConfigs()[2]
	assert.Equal(t, "fakedevice.uri", configURI.GetKey())
	assert.Equal(t, "http://mydevice.mydomain.com", configURI.GetValue())
	assert.Equal(t, "http://mydevice.mydomain.com", devices[1].GetURI())

	currentTest.Cleanup()
}

func createTestOrganization() *config.Organization {

	org := config.NewOrganization()
	org.SetName("Test Org")

	schedule1 := config.NewSchedule()
	days := "MO,WE,FR"
	endDate := time.Now().AddDate(0, 1, 0)
	schedule1.SetStartDate(time.Now())
	schedule1.SetEndDate(&endDate)
	schedule1.SetFrequency(1) // daily
	schedule1.SetInterval(60) // seconds
	schedule1.SetCount(1)     // total number of times to run
	schedule1.SetDays(&days)
	schedule1.SetLastExecuted(time.Now())
	schedule1.SetExecutionCount(1)

	schedules := []config.Schedule{*schedule1}

	metric1 := config.NewMetric()
	metric1.SetName("Fake Temperature")
	metric1.SetKey("sensor1")
	metric1.SetEnable(true)
	metric1.SetNotify(true)
	metric1.SetUnit("°")
	metric1.SetAlarmLow(50.8)
	metric1.SetAlarmHigh(100.0)

	metric2 := config.NewMetric()
	metric2.SetName("Fake Relative Humidity")
	metric2.SetKey("sensor2")
	metric2.SetEnable(true)
	metric2.SetNotify(true)
	metric2.SetUnit("%")
	metric2.SetAlarmLow(30.0)
	metric2.SetAlarmHigh(70.0)

	condition1 := config.NewCondition()
	//condition1.SetMetricID(metric1.GetID()) // TODO: How to get this to persist?!
	condition1.SetComparator(">")
	condition1.SetThreshold(120.0)

	conditions := []config.Condition{*condition1}

	channel1 := config.NewChannel()
	channel1.SetName("Test Channel 1")
	channel1.SetEnable(true)
	channel1.SetNotify(true)
	channel1.SetConditions(conditions)
	channel1.SetSchedule(schedules)
	channel1.SetDuration(1)
	channel1.SetDebounce(2)
	channel1.SetBackoff(3)
	//channel1.SetAlgorithm()

	enableConfigItem := config.NewDeviceConfigItem()
	enableConfigItem.SetKey("fakedevice.enable")
	enableConfigItem.SetValue("true")

	notifyConfigItem := config.NewDeviceConfigItem()
	notifyConfigItem.SetKey("fakedevice.notify")
	notifyConfigItem.SetValue("false")

	uriConfigItem := config.NewDeviceConfigItem()
	uriConfigItem.SetKey("fakedevice.uri")
	uriConfigItem.SetValue("http://mydevice.mydomain.com")

	serverDevice := config.Device{
		Type: "server",
		Configs: []config.DeviceConfigItem{
			{
				Key:   "name",
				Value: "Fake Farm"},
			{
				Key:   "interval",
				Value: "58"},
			{
				Key:   "mode",
				Value: "test"},
			{
				Key:   "timezone",
				Value: "America/New_York"},
			{
				Key:   "smtp.enable",
				Value: "true"},
			{
				Key:   "smtp.host",
				Value: "127.0.0.1"},
			{
				Key:   "smtp.port",
				Value: "587"},
			{
				Key:   "smtp.username",
				Value: "foo"},
			{
				Key:   "smtp.password",
				Value: "bar"},
			{
				Key:   "smtp.recipient",
				Value: "user@domain.com"},
		}}

	configs := []config.DeviceConfigItem{*enableConfigItem, *notifyConfigItem, *uriConfigItem}
	metrics := []config.Metric{*metric1, *metric2}
	channels := []config.Channel{*channel1}

	device1 := config.NewDevice()
	device1.SetType("fakedevice")
	device1.SetInterval(59)
	device1.SetDescription("This is a fake device used for testing")
	device1.SetHardwareVersion("hw-v0.0.1a")
	device1.SetFirmwareVersion("fw-v0.0.1a")
	device1.SetConfigs(configs)
	device1.SetMetrics(metrics)
	device1.SetChannels(channels)

	devices := []config.Device{serverDevice, *device1}

	farm1 := config.NewFarm()
	farm1.SetMode("test")
	farm1.SetName("Fake Farm")
	farm1.SetInterval(58)
	farm1.SetDevices(devices)

	farms := []config.Farm{*farm1}

	org.SetFarms(farms)

	return org
}
