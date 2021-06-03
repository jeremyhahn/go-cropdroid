package config

import "fmt"

type Server struct {
	ID       int    `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Interval int    `yaml:"interval" json:"interval"`
	Timezone string `yaml:"timezone" json:"timezone"`
	Mode     string `yaml:"mode" json:"mode"`
	//Smtp          *Smtp          `yaml:"smtp" json:"smtp"`
	License       *License       `yaml:"license" json:"license"`
	Organizations []Organization `yaml:"organizations" json:"organizations"`
	Farms         []Farm         `gorm:"-" yaml:"farms" json:"farms"`
	ServerConfig  `yaml:"-" json:"-"`
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

/*
func (config *Server) SetSmtp(smtp *Smtp) {
	config.Smtp = smtp
}

func (config *Server) GetSmtp() *Smtp {
	return config.Smtp
}
*/

func (config *Server) GetLicense() *License {
	return config.License
}

func (config *Server) SetLicense(license *License) {
	config.License = license
}

func (config *Server) SetOrganizations(orgs []Organization) {
	config.Organizations = orgs
}

func (config *Server) GetOrganizations() []Organization {
	return config.Organizations
}

func (config *Server) GetOrganization(id int) (*Organization, error) {
	for _, org := range config.Organizations {
		if org.GetID() == id {
			return &org, nil
		}
	}
	return nil, fmt.Errorf("Organization id not found: %d", id)
}

func (config *Server) GetFarms() []Farm {
	return config.Farms
}

func (config *Server) SetFarms(farms []Farm) {
	config.Farms = farms
}

func (config *Server) SetFarm(id int, farm FarmConfig) {
	for i, farm := range config.GetFarms() {
		if farm.GetID() == id {
			config.Farms[i] = farm
		}
	}
}
