package data

import "github.com/jeremyhahn/go-cropdroid/config"

func OrgsEqual(org1, org2 *config.OrganizationStruct) bool {
	if org1.ID != org2.ID {
		return false
	}
	if org1.GetName() != org2.GetName() {
		return false
	}
	farms := org1.GetFarms()
	if !FarmsEqual(farms[0], farms[1]) {
		return false
	}
	return true
}

func FarmsEqual(farm1, farm2 *config.FarmStruct) bool {
	if farm1.ID != farm2.ID {
		return false
	}
	if farm1.GetName() != farm2.GetName() {
		return false
	}
	if farm1.GetMode() != farm2.GetMode() {
		return false
	}
	farm1Devices := farm1.GetDevices()
	farm2Devices := farm2.GetDevices()
	if !FarmDevicesEqual(farm1Devices, farm2Devices) {
		return false
	}
	return true
}

func FarmDevicesEqual(farm1Devices, farm2Devices []*config.DeviceStruct) bool {
	if len(farm1Devices) != len(farm2Devices) {
		return false
	}
	for i, f1Device := range farm1Devices {
		f2Device := farm2Devices[i]
		if f1Device.GetType() != f2Device.GetType() {
			return false
		}
		f1DeviceConfigs := f1Device.GetSettings()
		f2DeviceConfigs := f2Device.GetSettings()
		if !DeviceConfigsEqual(f1DeviceConfigs, f2DeviceConfigs) {
			return false
		}
	}
	return true
}

func DeviceConfigsEqual(f1DeviceConfigs, f2DeviceConfigs []*config.DeviceSettingStruct) bool {
	for j, f1DeviceConfig := range f1DeviceConfigs {
		f2DeviceConfig := f2DeviceConfigs[j]
		if f1DeviceConfig.GetKey() != f2DeviceConfig.GetKey() {
			return false
		}
		if f1DeviceConfig.GetValue() != f2DeviceConfig.GetValue() {
			return false
		}
	}
	return true
}
