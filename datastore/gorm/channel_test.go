package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestChannelCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.LogMode(true)
	currentTest.gorm.AutoMigrate(&config.DeviceConfigItem{})
	currentTest.gorm.AutoMigrate(&config.Channel{})

	channelDAO := NewChannelDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, channelDAO)

	channel1 := &config.Channel{
		DeviceID:    1,
		ChannelID:   3,
		Name:        "Test Channel 1",
		Enable:      true,
		Notify:      true,
		Duration:    2,
		Debounce:    3,
		Backoff:     4,
		AlgorithmID: 1}

	channel2 := &config.Channel{
		DeviceID:    3,
		ChannelID:   4,
		Name:        "Test Channel 2",
		Enable:      false,
		Notify:      false,
		Duration:    10,
		Debounce:    20,
		Backoff:     30,
		AlgorithmID: 2}

	err := channelDAO.Save(channel1)
	assert.Nil(t, err)

	err = channelDAO.Save(channel2)
	assert.Nil(t, err)

	persistedChannel1, err := channelDAO.Get(channel1.GetID())
	assert.Nil(t, err)

	assert.Equal(t, channel1.ID, persistedChannel1.GetID())
	assert.Equal(t, channel1.DeviceID, persistedChannel1.GetDeviceID())
	assert.Equal(t, channel1.Name, persistedChannel1.GetName())
	assert.Equal(t, channel1.Enable, persistedChannel1.IsEnabled())
	assert.Equal(t, channel1.Notify, persistedChannel1.IsNotify())
	assert.Equal(t, channel1.Duration, persistedChannel1.GetDuration())
	assert.Equal(t, channel1.Debounce, persistedChannel1.GetDebounce())
	assert.Equal(t, channel1.Backoff, persistedChannel1.GetBackoff())
	assert.Equal(t, channel1.AlgorithmID, persistedChannel1.GetAlgorithmID())

	persistedChannels, err := channelDAO.GetByDeviceID(3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedChannels))

	persistedChannel2 := persistedChannels[0]

	assert.Equal(t, channel2.ID, persistedChannel2.GetID())
	assert.Equal(t, channel2.DeviceID, persistedChannel2.GetDeviceID())
	assert.Equal(t, channel2.Name, persistedChannel2.GetName())
	assert.Equal(t, channel2.Enable, persistedChannel2.IsEnabled())
	assert.Equal(t, channel2.Notify, persistedChannel2.IsNotify())
	assert.Equal(t, channel2.Duration, persistedChannel2.GetDuration())
	assert.Equal(t, channel2.Debounce, persistedChannel2.GetDebounce())
	assert.Equal(t, channel2.Backoff, persistedChannel2.GetBackoff())
	assert.Equal(t, channel2.AlgorithmID, persistedChannel2.GetAlgorithmID())

	currentTest.Cleanup()
}

func TestChannelGetByUserOrgAndDeviceID(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})
	currentTest.gorm.AutoMigrate(&config.Metric{})
	currentTest.gorm.AutoMigrate(&config.Channel{})
	currentTest.gorm.AutoMigrate(&config.DeviceConfigItem{})
	currentTest.gorm.AutoMigrate(&config.Device{})
	currentTest.gorm.AutoMigrate(&config.Farm{})
	currentTest.gorm.AutoMigrate(&config.Organization{})

	deviceDAO := NewDeviceDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, deviceDAO)

	channelDAO := NewChannelDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, channelDAO)

	farmDAO := NewFarmDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, farmDAO)

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, userDAO)

	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, roleDAO)

	role := config.NewRole()
	role.SetName("test")

	user := config.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("$ecret")
	user.SetRoles([]config.Role{*role})

	channel1 := config.NewChannel()
	channel1.SetDeviceID(1)
	channel1.SetChannelID(3)
	channel1.SetName("Test Channel 1")
	channel1.SetEnable(true)
	channel1.SetNotify(true)
	channel1.SetDuration(2)
	channel1.SetDebounce(3)
	channel1.SetBackoff(4)
	//channel1.SetAlgorithmID(1)

	channel2 := config.NewChannel()
	channel2.SetDeviceID(3)
	channel2.SetChannelID(4)
	channel2.SetName("Test Channel 2")
	channel2.SetEnable(false)
	channel2.SetNotify(false)
	channel2.SetDuration(10)
	channel2.SetDebounce(20)
	channel2.SetBackoff(30)
	//channel1.SetAlgorithmID(2)

	device1 := config.NewDevice()
	device1.SetType("fake")
	device1.SetDescription("This is a fake device used for integration testing")
	device1.SetInterval(30)
	device1.SetChannels([]config.Channel{*channel1})

	farm := config.NewFarm()
	farm.SetName("Test Farm")
	farm.SetMode("test")
	farm.SetInterval(60)
	farm.SetDevices([]config.Device{*device1})
	// gorm doesnt handle multiple many to many relationships in a single entity,
	// need to manage the relationship manually
	//farm.SetUsers([]config.User{*user})

	err := farmDAO.Save(farm)
	assert.Nil(t, err)

	err = userDAO.Create(user)
	assert.Nil(t, err)

	err = roleDAO.Create(role)
	assert.Nil(t, err)

	currentTest.gorm.Create(&config.Permission{
		OrganizationID: 0,
		FarmID:         farm.GetID(),
		UserID:         user.GetID(),
		RoleID:         role.GetID()})

	err = channelDAO.Save(channel2)
	assert.Nil(t, err)

	devices, err := deviceDAO.GetByFarmId(farm.GetID())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(devices))
	assert.Equal(t, true, devices[0].GetID() > 0)

	persistedChannels, err := channelDAO.GetByOrgUserAndDeviceID(0, user.GetID(), devices[0].GetID())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedChannels))

	persistedChannel1 := persistedChannels[0]
	//assert.Equal(t, channel1.ID, persistedChannel1.GetID())
	//assert.Equal(t, channel1.DeviceID, persistedChannel1.GetDeviceID())
	assert.Equal(t, channel1.Name, persistedChannel1.GetName())
	assert.Equal(t, channel1.Enable, persistedChannel1.IsEnabled())
	assert.Equal(t, channel1.Notify, persistedChannel1.IsNotify())
	assert.Equal(t, channel1.Duration, persistedChannel1.GetDuration())
	assert.Equal(t, channel1.Debounce, persistedChannel1.GetDebounce())
	assert.Equal(t, channel1.Backoff, persistedChannel1.GetBackoff())
	assert.Equal(t, channel1.AlgorithmID, persistedChannel1.GetAlgorithmID())

	currentTest.Cleanup()
}
