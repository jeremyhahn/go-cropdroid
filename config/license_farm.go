package config

type FarmLicense interface {
	GetFarmID() uint64
	GetDeviceQuota() int
	GetUserQuota() int
}

type FarmLicenseStruct struct {
	FarmID      uint64 `yaml:"farmID" json:"farmID"`
	DeviceQuota int    `yaml:"deviceQuota" json:"deviceQuota"`
	UserQuota   int    `yaml:"userQuota" json:"userQuota"`
}

func (license *FarmLicenseStruct) GetFarmID() uint64 {
	return license.FarmID
}

func (license *FarmLicenseStruct) GetDeviceQuota() int {
	return license.DeviceQuota
}

func (license *FarmLicenseStruct) GetUserQuota() int {
	return license.UserQuota
}
