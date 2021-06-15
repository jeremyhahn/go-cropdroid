package model

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
)

// Organization groups users and devices
type Organization struct {
	ID                  int           `yaml:"id" json:"id"`
	Name                string        `yaml:"name" json:"name"`
	Farms               []common.Farm `yaml:"farms" json:"farms"`
	common.Organization `yaml:"-" json:"-"`
}

// GetID returns the unique identifier for the org
func (o *Organization) GetID() int {
	return o.ID
}

// GetName returns the org name
func (o *Organization) GetName() string {
	return o.Name
}

func (o *Organization) SetFarms(farms []common.Farm) {
	o.Farms = farms
}

func (o *Organization) GetFarms() []common.Farm {
	return o.Farms
}

func (o *Organization) GetFarm(id int) (common.Farm, error) {
	for _, farm := range o.Farms {
		if farm.GetID() == id {
			return farm, nil
		}
	}
	return nil, fmt.Errorf("Farm not found at index: %d", id)
}
