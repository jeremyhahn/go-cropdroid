package config

type License struct {
	OrganizationID  int `yaml:"organizationId" json:"organizationId"`
	UserQuota       int `yaml:"userQuota" json:"userQuota"`
	FarmQuota       int `yaml:"farmQuota" json:"farmQuota"`
	ControllerQuota int `yaml:"controllerQuota" json:"controllerQuota"`
	LicenseConfig   `yaml:"-" json:"-"`
}

func (license *License) GetUserQuota() int {
	return license.UserQuota
}

func (license *License) GetFarmQuota() int {
	return license.FarmQuota
}

func (license *License) GetControllerQuota() int {
	return license.ControllerQuota
}
