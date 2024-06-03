package config

type OrganizationLicense interface {
	GetOrganizationID() uint64
	//SetOrganizationID(id int64)
	GetUserQuota() int
	//SetUserQuota(quota int)
	GetFarmQuota() int
	//SetFarmQuota(quota int)
}

type OrganizationLicenseStruct struct {
	OrganizationID uint64 `yaml:"organizationId" json:"organizationId"`
	UserQuota      int    `yaml:"userQuota" json:"userQuota"`
	FarmQuota      int    `yaml:"farmQuota" json:"farmQuota"`
}

func (license *OrganizationLicenseStruct) GetOrganizationID() uint64 {
	return license.OrganizationID
}

func (license *OrganizationLicenseStruct) GetUserQuota() int {
	return license.UserQuota
}

func (license *OrganizationLicenseStruct) GetFarmQuota() int {
	return license.FarmQuota
}
