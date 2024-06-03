package config

type ServerLicense interface {
	GetOrganizationQuota() int
}

type ServerLicenseStruct struct {
	OrganizationQuota int
}

func (license *ServerLicenseStruct) GetOrganizationQuota() int {
	return license.OrganizationQuota
}
