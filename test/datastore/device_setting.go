package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/assert"
)

func TestDeviceSettingCRUD(t *testing.T, deviceDAO dao.DeviceDAO,
	deviceSettingDAO dao.DeviceSettingDAO, org *config.Organization) {

	newFirmwareVersion := "new-v1.2.3"
	farm1 := org.GetFarms()[0]
	device1 := farm1.GetDevices()[0]
	device1.SetFirmwareVersion(newFirmwareVersion)

	err := deviceDAO.Save(device1)
	assert.Nil(t, err)

	deviceSetting := config.NewDeviceSetting()
	deviceSetting.SetDeviceID(device1.GetID())
	deviceSetting.SetKey("test")
	deviceSetting.SetValue("testvalue")

	err = deviceSettingDAO.Save(farm1.GetID(), deviceSetting)
	assert.Nil(t, err)

	persistedDeviceSetting, err := deviceDAO.Get(
		farm1.GetID(), device1.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, newFirmwareVersion, persistedDeviceSetting.GetFirmwareVersion())
}
