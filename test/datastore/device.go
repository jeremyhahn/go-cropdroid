package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCRUD(t *testing.T, deviceDAO dao.DeviceDAO,
	farm *config.Farm) {

	device1, err := farm.GetDevice(DEVICE1_TYPE)
	assert.Nil(t, err)
	assert.NotNil(t, device1)

	device1.SetType("newdevice")
	err = deviceDAO.Save(device1)
	assert.Nil(t, err)

	persisetdDevice, err := deviceDAO.Get(farm.GetID(),
		device1.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, device1.GetFarmID(), farm.GetID())
	assert.Equal(t, "newdevice", persisetdDevice.GetType())
}
