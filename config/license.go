package config

type License struct {
	OrganizationID int `yaml:"organizationId" json:"organizationId"`
	UserQuota      int `yaml:"userQuota" json:"userQuota"`
	FarmQuota      int `yaml:"farmQuota" json:"farmQuota"`
	DeviceQuota    int `yaml:"deviceQuota" json:"deviceQuota"`
}

func (license *License) GetUserQuota() int {
	return license.UserQuota
}

func (license *License) GetFarmQuota() int {
	return license.FarmQuota
}

func (license *License) GetDeviceQuota() int {
	return license.DeviceQuota
}
