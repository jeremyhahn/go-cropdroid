package config

import (
	"errors"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
)

type Server struct {
	ID                  int            `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id" mapstructure:"id"`
	Interval            int            `yaml:"interval" json:"interval" mapstructure:"interval"`
	Timezone            string         `yaml:"timezone" json:"timezone" mapstructure:"timezone"`
	Mode                string         `yaml:"mode" json:"mode" mapstructure:"mode"`
	DefaultRole         string         `yaml:"default_role" json:"default_role" mapstructure:"default_role"`
	DefaultPermission   string         `yaml:"default_permission" json:"default_permission" mapstructure:"default_permission"`
	DataStoreEngine     string         `yaml:"datastore" json:"datastore" mapstructure:"datastore"`
	DataStoreCDC        bool           `yaml:"datastore_cdc" json:"datastore_cdc" mapstructure:"datastore_cdc"`
	DataDir             string         `yaml:"datadir" json:"datadir" mapstructure:"datadir"`
	DowngradeUser       string         `yaml:"www_user" json:"www_user" mapstructure:"www_user"`
	EnableRegistrations bool           `yaml:"enable_registrations" json:"enable_registrations" mapstructure:"enable_registrations"`
	EnableDefaultFarm   bool           `yaml:"enable_default_farm" json:"enable_default_farm" mapstructure:"enable_default_farm"`
	NodeID              int            `yaml:"node_id" json:"node_id" mapstructure:"node_id"`
	RedirectHttpToHttps bool           `yaml:"redirect_http_https" json:"redirect_http_https" mapstructure:"redirect_http_https"`
	SSLFlag             bool           `yaml:"ssl" json:"ssl" mapstructure:"ssl"`
	WebPort             int            `yaml:"port" json:"port" mapstructure:"port"`
	Smtp                *Smtp          `yaml:"smtp" json:"smtp" mapstructure:"smtp"`
	LicenseBlob         string         `yaml:"license" json:"license" mapstructure:"license"`
	License             *License       `yaml:"-" json:"-" mapstructure:"-"`
	Organizations       []Organization `yaml:"organizations" json:"organizations" mapstructure:"organizations"`
	Farms               []Farm         `gorm:"-" yaml:"farms" json:"farms"`
	ServerConfig        `yaml:"-" json:"-"`
}

func NewServer() ServerConfig {
	return &Server{
		ID:            1,
		Interval:      60,
		Mode:          "virtual",
		Organizations: make([]Organization, 0)}
}

func (config *Server) SetID(id int) {
	config.ID = id
}

func (config *Server) GetID() int {
	return config.ID
}

func (config *Server) SetInterval(interval int) {
	config.Interval = interval
}

func (config *Server) GetInterval() int {
	return config.Interval
}

func (config *Server) SetTimezone(timezone string) {
	config.Timezone = timezone
}

func (config *Server) GetTimezone() string {
	return config.Timezone
}

func (config *Server) SetMode(mode string) {
	config.Mode = mode
}

func (config *Server) GetMode() string {
	return config.Mode
}

func (config *Server) SetDefaultRole(role string) {
	config.DefaultRole = role
}

func (config *Server) GetDefaultRole() string {
	return config.DefaultRole
}

func (config *Server) SetDefaultPermission(permission string) {
	config.DefaultPermission = permission
}

func (config *Server) GetDefaultPermission() string {
	return config.DefaultPermission
}

func (config *Server) SetSmtp(smtp *Smtp) {
	config.Smtp = smtp
}

func (config *Server) GetSmtp() *Smtp {
	return config.Smtp
}

func (config *Server) GetLicense() *License {
	return config.License
}

func (config *Server) SetLicense(license *License) {
	config.License = license
}

func (config *Server) SetOrganizations(orgs []OrganizationConfig) {
	orgStructs := make([]Organization, len(orgs))
	for _, org := range orgs {
		orgStructs = append(orgStructs, *org.(*Organization))
	}
	config.Organizations = orgStructs
}

func (config *Server) GetOrganizations() []OrganizationConfig {
	orgConfigs := make([]OrganizationConfig, len(config.Organizations))
	for _, org := range config.Organizations {
		orgConfigs = append(orgConfigs, &org)
	}
	return orgConfigs
}

func (config *Server) GetOrganization(id uint64) (OrganizationConfig, error) {
	for _, org := range config.Organizations {
		if org.GetID() == id {
			return &org, nil
		}
	}
	return nil, ErrOrganizationNotFound
}

func (config *Server) GetFarms() []Farm {
	return config.Farms
}

func (config *Server) SetFarms(farms []FarmConfig) {
	farmStructs := make([]Farm, len(farms))
	for _, farm := range farms {
		farmStructs = append(farmStructs, *farm.(*Farm))
	}
	config.Farms = farmStructs
}

func (config *Server) AddFarm(farm FarmConfig) {
	config.Farms = append(config.Farms, *farm.(*Farm))
}
