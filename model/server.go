package model

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
)

type Server struct {
	ID            int            `yaml:"id" json:"id"`
	Organizations []Organization `yaml:"organizations" json:"organizations"`
	Interval      int            `yaml:"interval" json:"interval"`
	Timezone      string         `yaml:"timezone" json:"timezone"`
	Mode          string         `yaml:"mode" json:"mode"`
	Smtp          config.Smtp    `yaml:"smtp" json:"smtp"`
	common.Server `yaml:"-" json:"-"`
}

func NewServer() common.Server {
	return &Server{Organizations: make([]Organization, 0)}
}

func (model *Server) SetID(id int) {
	model.ID = id
}

func (model *Server) GetID() int {
	return model.ID
}

func (model *Server) SetOrganizations(orgs []Organization) {
	model.Organizations = orgs
}

func (config *Server) GetOrganizations() []Organization {
	return config.Organizations
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

func (config *Server) SetSmtp(smtp config.Smtp) {
	config.Smtp = smtp
}

func (config *Server) GetSmtp() config.Smtp {
	return config.Smtp
}
