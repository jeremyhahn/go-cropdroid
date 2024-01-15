package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/assert"
)

func TestChannelCRUD(t *testing.T, channelDAO dao.ChannelDAO,
	org *config.Organization) {

	farm1 := org.GetFarms()[0]
	device1 := farm1.GetDevices()[1]
	channel1 := device1.GetChannels()[0]
	channel2 := device1.GetChannels()[1]

	err := channelDAO.Save(farm1.GetID(), channel1)
	assert.Nil(t, err)

	err = channelDAO.Save(farm1.GetID(), channel2)
	assert.Nil(t, err)

	orgID := uint64(0)
	farmID := farm1.GetID()
	persistedChannel, err := channelDAO.Get(orgID, farmID,
		channel1.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)

	assert.Equal(t, channel1.GetID(), persistedChannel.GetID())
	assert.Equal(t, channel1.GetDeviceID(), persistedChannel.GetDeviceID())
	assert.Equal(t, channel1.GetName(), persistedChannel.GetName())
	assert.Equal(t, channel1.IsEnabled(), persistedChannel.IsEnabled())
	assert.Equal(t, channel1.IsNotify(), persistedChannel.IsNotify())
	assert.Equal(t, channel1.GetDuration(), persistedChannel.GetDuration())
	assert.Equal(t, channel1.GetDebounce(), persistedChannel.GetDebounce())
	assert.Equal(t, channel1.GetBackoff(), persistedChannel.GetBackoff())
	assert.Equal(t, channel1.GetAlgorithmID(), persistedChannel.GetAlgorithmID())
}

func TestChannelGetByDevice(t *testing.T, farmDAO dao.FarmDAO,
	deviceDAO dao.DeviceDAO, channelDAO dao.ChannelDAO,
	permissionDAO dao.PermissionDAO, org *config.Organization) {

	farm1 := org.GetFarms()[0]
	device1 := farm1.GetDevices()[1]
	channel1 := device1.GetChannels()[0]

	err := farmDAO.Save(farm1)
	assert.Nil(t, err)

	permissionDAO.Save(&config.Permission{
		OrganizationID: 0,
		FarmID:         farm1.GetID(),
		UserID:         farm1.GetUsers()[0].GetID(),
		RoleID:         farm1.GetUsers()[0].GetRoles()[0].GetID()})

	newChannelName := "newtest"
	channel1.SetName(newChannelName)
	err = channelDAO.Save(farm1.GetID(), channel1)
	assert.Nil(t, err)

	device, err := deviceDAO.Get(farm1.GetID(), device1.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, device1.GetID(), device.GetID())

	persistedChannels, err := channelDAO.GetByDevice(org.GetID(),
		farm1.GetID(), device1.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(persistedChannels))

	// Gorm and Raft items are stored in different order.
	// Raft returns records in the same order they were saved.
	// GORM returns records ordered by id.
	// This loop performs assertions regardless of order
	found := false
	for _, persistedChannel := range persistedChannels {
		if channel1.GetID() == persistedChannel.GetID() {
			assert.Equal(t, channel1.GetDeviceID(), persistedChannel.GetDeviceID())
			assert.Equal(t, newChannelName, persistedChannel.GetName())
			assert.Equal(t, channel1.IsEnabled(), persistedChannel.IsEnabled())
			assert.Equal(t, channel1.IsNotify(), persistedChannel.IsNotify())
			assert.Equal(t, channel1.GetDuration(), persistedChannel.GetDuration())
			assert.Equal(t, channel1.GetDebounce(), persistedChannel.GetDebounce())
			assert.Equal(t, channel1.GetBackoff(), persistedChannel.GetBackoff())
			assert.Equal(t, channel1.GetAlgorithmID(), persistedChannel.GetAlgorithmID())
			found = true
		}
	}
	assert.True(t, found)
}
