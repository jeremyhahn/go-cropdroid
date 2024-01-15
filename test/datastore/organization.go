package datastore

import (
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/util"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationCRUD(t *testing.T, orgDAO dao.OrganizationDAO) {

	orgs, err := orgDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(orgs))

	err = orgDAO.Save(&config.Organization{
		Name: "Test Org"})
	assert.Nil(t, err)

	assert.NotNil(t, orgDAO)

	orgs, err = orgDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(orgs))
	assert.Equal(t, "Test Org", orgs[0].GetName())
}

func TestOrganizationGetAll(t *testing.T, orgDAO dao.OrganizationDAO) {

	// create first org
	testOrgName := "Test Org"
	testFarmName := "Test Farm"

	farmConfig := config.NewFarm()
	farmConfig.SetName(testFarmName)

	orgConfig := &config.Organization{
		Name:  testOrgName,
		Farms: []*config.Farm{farmConfig}}

	err := orgDAO.Save(orgConfig)
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
}

func TestOrganizationDelete(t *testing.T, orgDAO dao.OrganizationDAO) {

	// create first org
	testOrgName := "Test Org"
	testFarmName := "Test Farm"

	farmConfig := config.NewFarm()
	farmConfig.SetName(testFarmName)

	orgConfig := &config.Organization{
		Name:  testOrgName,
		Farms: []*config.Farm{farmConfig}}

	err := orgDAO.Save(orgConfig)
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

func TestOrganizationEnchilada(t *testing.T, orgDAO dao.OrganizationDAO,
	roleDAO dao.RoleDAO, userDAO dao.UserDAO,
	permissionDAO dao.PermissionDAO, org *config.Organization) {

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
	assert.Equal(t, FARM1_NAME, farms[0].GetName())
	assert.Equal(t, "test", farms[0].GetMode())
	assert.Equal(t, FARM2_NAME, farms[1].GetName())
	assert.Equal(t, "test2", farms[1].GetMode())

	farm1 := farms[0]
	serverDevice, err := farm1.GetDevice(SERVER_TYPE)
	assert.Nil(t, err)
	assert.NotNil(t, serverDevice)

	device1, err := farm1.GetDevice(DEVICE1_TYPE)
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
}

func CreateTestOrganization(idGenerator util.IdGenerator) *config.Organization {

	org := config.NewOrganization()
	org.SetName("Test Org")

	farm1 := CreateFarm1(idGenerator)
	farm2 := CreateFarm2(idGenerator)

	farms := []*config.Farm{farm1, farm2}

	org.SetFarms(farms)

	return org
}

func CreateFarm1(idGenerator util.IdGenerator) *config.Farm {

	farmID := idGenerator.NewID(FARM1_NAME)
	device1ID := idGenerator.NewID("farm1-device1")
	channel1ID := idGenerator.NewID("farm1-channel1")
	channel2ID := idGenerator.NewID("farm1-channel2")

	roleName := "test"
	role := config.NewRole()
	role.SetName(roleName)

	userEmail := "root@localhost"
	user := config.NewUser()
	user.SetEmail(userEmail)
	user.SetPassword("$ecret")
	user.SetRoles([]*config.Role{role})

	schedule1 := config.NewSchedule()
	days := "MO,WE,FR"
	endDate := time.Now().AddDate(0, 1, 0)
	schedule1.SetID(idGenerator.NewID("farm1-schedule1"))
	schedule1.SetChannelID(channel1ID)
	schedule1.SetStartDate(time.Now())
	schedule1.SetEndDate(&endDate)
	schedule1.SetFrequency(1) // daily
	schedule1.SetInterval(60) // seconds
	schedule1.SetCount(1)     // total number of times to run
	schedule1.SetDays(&days)
	schedule1.SetLastExecuted(time.Now())
	schedule1.SetExecutionCount(1)

	schedule2 := config.NewSchedule()
	days = "TU,TH,SU"
	endDate = time.Now().AddDate(0, 3, 0)
	schedule2.SetID(idGenerator.NewID("farm1-schedule2"))
	schedule2.SetChannelID(channel1ID)
	schedule2.SetStartDate(time.Now())
	schedule2.SetEndDate(&endDate)
	schedule2.SetFrequency(1)   // daily
	schedule2.SetInterval(3600) // seconds
	schedule2.SetCount(3)       // total number of times to run
	schedule2.SetDays(&days)
	schedule2.SetLastExecuted(time.Now())
	schedule2.SetExecutionCount(1)

	schedules := []*config.Schedule{schedule1, schedule2}

	metric1 := config.NewMetric()
	metric1.SetID(idGenerator.NewID("farm1-metric1"))
	metric1.SetDeviceID(device1ID)
	metric1.SetName("Fake Temperature")
	metric1.SetKey("sensor1")
	metric1.SetEnable(true)
	metric1.SetNotify(true)
	metric1.SetUnit("°")
	metric1.SetAlarmLow(50.8)
	metric1.SetAlarmHigh(100.0)

	metric2 := config.NewMetric()
	metric2.SetID(idGenerator.NewID("farm1-metric2"))
	metric2.SetDeviceID(device1ID)
	metric2.SetName("Fake Relative Humidity")
	metric2.SetKey("sensor2")
	metric2.SetEnable(true)
	metric2.SetNotify(true)
	metric2.SetUnit("%")
	metric2.SetAlarmLow(30.0)
	metric2.SetAlarmHigh(70.0)

	condition1 := config.NewCondition()
	condition1.SetID(idGenerator.NewID("farm1-condition1"))
	condition1.SetChannelID(channel1ID)
	//condition1.SetMetricID(metric1.GetID()) // TODO: How to get this to persist?!
	condition1.SetComparator(">")
	condition1.SetThreshold(120.0)

	condition2 := config.NewCondition()
	condition2.SetID(idGenerator.NewID("farm1-condition2"))
	condition2.SetChannelID(channel1ID)
	//condition2.SetMetricID(metric1.GetID()) // TODO: How to get this to persist?!
	condition2.SetComparator("<")
	condition2.SetThreshold(10)

	conditions := []*config.Condition{condition1, condition2}

	channel1 := config.NewChannel()
	channel1.SetID(channel1ID)
	channel1.SetDeviceID(device1ID)
	channel1.SetName(CHANNEL1_NAME)
	channel1.SetEnable(true)
	channel1.SetNotify(true)
	channel1.SetConditions(conditions)
	channel1.SetSchedule(schedules)
	channel1.SetDuration(1)
	channel1.SetDebounce(2)
	channel1.SetBackoff(3)
	//channel1.SetAlgorithm()

	channel2 := config.NewChannel()
	channel2.SetID(channel2ID)
	channel2.SetDeviceID(device1ID)
	channel2.SetName(CHANNEL2_NAME)
	channel2.SetEnable(true)
	channel2.SetNotify(true)
	channel2.SetConditions(conditions)
	channel2.SetDuration(1)
	channel2.SetDebounce(2)
	channel2.SetBackoff(3)

	enableConfigItem := config.NewDeviceSetting()
	//enableConfigItem.SetID(idGenerator.NewID("enableConfigItem"))
	enableConfigItem.SetKey("fakedevice.enable")
	enableConfigItem.SetValue("true")

	notifyConfigItem := config.NewDeviceSetting()
	//notifyConfigItem.SetID(idGenerator.NewID("notifyConfigItem"))
	notifyConfigItem.SetKey("fakedevice.notify")
	notifyConfigItem.SetValue("false")

	uriConfigItem := config.NewDeviceSetting()
	//uriConfigItem.SetID(idGenerator.NewID("uriConfigItem"))
	uriConfigItem.SetKey("fakedevice.uri")
	uriConfigItem.SetValue("http://mydevice.mydomain.com")

	serverDevice := &config.Device{
		Type: SERVER_TYPE,
		Settings: []*config.DeviceSetting{
			{
				ID:    idGenerator.NewID("farm1-server-name"),
				Key:   "name",
				Value: FARM1_NAME},
			{
				ID:    idGenerator.NewID("farm1-server-interval"),
				Key:   "interval",
				Value: "58"},
			{
				ID:    idGenerator.NewID("farm1-server-mode"),
				Key:   "mode",
				Value: "test"},
			{
				ID:    idGenerator.NewID("farm1-server-timezone"),
				Key:   "timezone",
				Value: "America/New_York"},
			{
				ID:    idGenerator.NewID("farm1-server-smtp-enable"),
				Key:   "smtp.enable",
				Value: "true"},
			{
				ID:    idGenerator.NewID("farm1-server-smtp-host"),
				Key:   "smtp.host",
				Value: "127.0.0.1"},
			{
				ID:    idGenerator.NewID("farm1-server-smtp-port"),
				Key:   "smtp.port",
				Value: "587"},
			{
				ID:    idGenerator.NewID("farm1-server-smtp-username"),
				Key:   "smtp.username",
				Value: "foo"},
			{
				ID:    idGenerator.NewID("farm1-server-smtp-password"),
				Key:   "smtp.password",
				Value: "bar"},
			{
				ID:    idGenerator.NewID("farm1-server-smtp-recipient"),
				Key:   "smtp.recipient",
				Value: "user@domain.com"},
		}}
	serverDevice.SetID(idGenerator.NewID("farm1-server"))
	serverDevice.SetFarmID(farmID)

	settings := []*config.DeviceSetting{
		enableConfigItem, notifyConfigItem, uriConfigItem}
	metrics := []*config.Metric{metric1, metric2}
	channels := []*config.Channel{channel1, channel2}

	device1 := config.NewDevice()
	device1.SetID(device1ID)
	device1.SetFarmID(farmID)
	device1.SetType(DEVICE1_TYPE)
	device1.SetInterval(59)
	device1.SetDescription("This is a fake device used for testing")
	device1.SetHardwareVersion("hw-v0.0.1a")
	device1.SetFirmwareVersion("fw-v0.0.1a")
	device1.SetSettings(settings)
	device1.SetMetrics(metrics)
	device1.SetChannels(channels)

	devices := []*config.Device{serverDevice, device1}

	farm1 := config.NewFarm()
	farm1.SetID(farmID)
	farm1.SetMode("test")
	farm1.SetName(FARM1_NAME)
	farm1.SetInterval(60)
	farm1.SetDevices(devices)
	farm1.SetUsers([]*config.User{user})

	return farm1
}

func CreateFarm2(idGenerator util.IdGenerator) *config.Farm {

	farmID := idGenerator.NewID(FARM2_NAME)
	device1ID := idGenerator.NewID("farm2-device1")
	channel1ID := idGenerator.NewID("farm2-channel1")
	channel2ID := idGenerator.NewID("farm2-channel2")

	roleName := "test"
	role := config.NewRole()
	role.SetName(roleName)

	userEmail := "root@localhost"
	user := config.NewUser()
	user.SetEmail(userEmail)
	user.SetPassword("$ecret")
	user.SetRoles([]*config.Role{role})

	schedule1 := config.NewSchedule()
	days := "MO,WE,FR"
	endDate := time.Now().AddDate(0, 1, 0)
	schedule1.SetID(idGenerator.NewID("farm2-schedule1"))
	schedule1.SetChannelID(channel1ID)
	schedule1.SetStartDate(time.Now())
	schedule1.SetEndDate(&endDate)
	schedule1.SetFrequency(1) // daily
	schedule1.SetInterval(60) // seconds
	schedule1.SetCount(1)     // total number of times to run
	schedule1.SetDays(&days)
	schedule1.SetLastExecuted(time.Now())
	schedule1.SetExecutionCount(1)

	schedule2 := config.NewSchedule()
	days = "TU,TH,SU"
	endDate = time.Now().AddDate(0, 3, 0)
	schedule2.SetID(idGenerator.NewID("farm2-schedule2"))
	schedule1.SetChannelID(channel2ID)
	schedule2.SetStartDate(time.Now())
	schedule2.SetEndDate(&endDate)
	schedule2.SetFrequency(1)   // daily
	schedule2.SetInterval(3600) // seconds
	schedule2.SetCount(3)       // total number of times to run
	schedule2.SetDays(&days)
	schedule2.SetLastExecuted(time.Now())
	schedule2.SetExecutionCount(1)

	schedules := []*config.Schedule{schedule1, schedule2}

	metric1 := config.NewMetric()
	metric1.SetID(idGenerator.NewID("farm2-metric1"))
	metric1.SetDeviceID(device1ID)
	metric1.SetName("Fake Temperature")
	metric1.SetKey("sensor1")
	metric1.SetEnable(true)
	metric1.SetNotify(true)
	metric1.SetUnit("°")
	metric1.SetAlarmLow(50.8)
	metric1.SetAlarmHigh(100.0)

	metric2 := config.NewMetric()
	metric2.SetID(idGenerator.NewID("farm2-metric2"))
	metric2.SetDeviceID(device1ID)
	metric2.SetName("Fake Relative Humidity")
	metric2.SetKey("sensor2")
	metric2.SetEnable(true)
	metric2.SetNotify(true)
	metric2.SetUnit("%")
	metric2.SetAlarmLow(30.0)
	metric2.SetAlarmHigh(70.0)

	condition1 := config.NewCondition()
	condition1.SetID(idGenerator.NewID("farm2-condition1"))
	condition1.SetChannelID(channel1ID)
	//condition1.SetMetricID(metric1.GetID()) // TODO: How to get this to persist?!
	condition1.SetComparator(">")
	condition1.SetThreshold(120.0)

	condition2 := config.NewCondition()
	condition2.SetID(idGenerator.NewID("farm2-condition2"))
	condition2.SetChannelID(channel1ID)
	//condition2.SetMetricID(metric1.GetID()) // TODO: How to get this to persist?!
	condition2.SetComparator("<")
	condition2.SetThreshold(10)

	conditions := []*config.Condition{condition1, condition2}

	channel1 := config.NewChannel()
	channel1.SetID(channel1ID)
	channel1.SetDeviceID(device1ID)
	channel1.SetName(CHANNEL1_NAME)
	channel1.SetEnable(true)
	channel1.SetNotify(true)
	channel1.SetConditions(conditions)
	channel1.SetSchedule(schedules)
	channel1.SetDuration(1)
	channel1.SetDebounce(2)
	channel1.SetBackoff(3)
	//channel1.SetAlgorithm()

	channel2 := config.NewChannel()
	channel2.SetID(channel2ID)
	channel2.SetDeviceID(device1ID)
	channel2.SetName(CHANNEL2_NAME)
	channel2.SetEnable(true)
	channel2.SetNotify(true)
	channel2.SetConditions(conditions)
	channel2.SetDuration(1)
	channel2.SetDebounce(2)
	channel2.SetBackoff(3)

	enableConfigItem := config.NewDeviceSetting()
	//enableConfigItem.SetID(idGenerator.NewID("enableConfigItem"))
	enableConfigItem.SetKey("fakedevice.enable")
	enableConfigItem.SetValue("true")

	notifyConfigItem := config.NewDeviceSetting()
	//notifyConfigItem.SetID(idGenerator.NewID("notifyConfigItem"))
	notifyConfigItem.SetKey("fakedevice.notify")
	notifyConfigItem.SetValue("false")

	uriConfigItem := config.NewDeviceSetting()
	//uriConfigItem.SetID(idGenerator.NewID("uriConfigItem"))
	uriConfigItem.SetKey("fakedevice.uri")
	uriConfigItem.SetValue("http://mydevice.mydomain.com")

	serverDevice := &config.Device{
		Type: SERVER_TYPE,
		Settings: []*config.DeviceSetting{
			{
				ID:    idGenerator.NewID("farm2-server-name"),
				Key:   "name",
				Value: FARM2_NAME},
			{
				ID:    idGenerator.NewID("farm2-server-interval"),
				Key:   "interval",
				Value: "58"},
			{
				ID:    idGenerator.NewID("farm2-server-mode"),
				Key:   "mode",
				Value: "test2"},
			{
				ID:    idGenerator.NewID("farm2-server-timezone"),
				Key:   "timezone",
				Value: "America/New_York"},
			{
				ID:    idGenerator.NewID("farm2-server-smtp-enable"),
				Key:   "smtp.enable",
				Value: "true"},
			{
				ID:    idGenerator.NewID("farm2-server-smtp-host"),
				Key:   "smtp.host",
				Value: "127.0.0.1"},
			{
				ID:    idGenerator.NewID("farm2-server-smtp-port"),
				Key:   "smtp.port",
				Value: "587"},
			{
				ID:    idGenerator.NewID("farm2-server-smtp-username"),
				Key:   "smtp.username",
				Value: "foo"},
			{
				ID:    idGenerator.NewID("farm2-server-smtp=password"),
				Key:   "smtp.password",
				Value: "bar"},
			{
				ID:    idGenerator.NewID("farm2-server-smtp-recipient"),
				Key:   "smtp.recipient",
				Value: "user@domain.com"},
		}}
	serverDevice.SetID(idGenerator.NewID("farm2-server"))
	serverDevice.SetFarmID(farmID)

	settings := []*config.DeviceSetting{
		enableConfigItem, notifyConfigItem, uriConfigItem}
	metrics := []*config.Metric{metric1, metric2}
	channels := []*config.Channel{channel1, channel2}

	device1 := config.NewDevice()
	device1.SetID(idGenerator.NewID("farm2-device1"))
	device1.SetFarmID(farmID)
	device1.SetType(DEVICE1_TYPE)
	device1.SetInterval(59)
	device1.SetDescription("This is a 2nd fake device used for testing")
	device1.SetHardwareVersion("hw-v0.0.2a")
	device1.SetFirmwareVersion("fw-v0.0.2a")
	device1.SetSettings(settings)
	device1.SetMetrics(metrics)
	device1.SetChannels(channels)

	devices := []*config.Device{serverDevice, device1}

	farm2 := config.NewFarm()
	farm2.SetID(idGenerator.NewID(FARM2_NAME))
	farm2.SetMode("test")
	farm2.SetName(FARM2_NAME)
	farm2.SetInterval(60)
	farm2.SetDevices(devices)
	farm2.SetUsers([]*config.User{user})

	return farm2
}
