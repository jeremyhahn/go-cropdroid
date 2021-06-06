package model

import (
	"github.com/jeremyhahn/go-cropdroid/common"
)

type Farm struct {
	ID          int                 `yaml:"id" json:"id"`
	OrgID       int                 `yaml:"orgId" json:"orgId"`
	Mode        string              `yaml:"mode" json:"mode"`
	Name        string              `yaml:"name" json:"name"`
	Interval    int                 `yaml:"interval" json:"interval"`
	Controllers []common.Controller `yaml:"controllers" json:"controllers"`
	common.Farm `yaml:"-" json:"-"`
}

func NewFarm() common.Farm {
	return &Farm{Controllers: make([]common.Controller, 0)}
}

func CreateFarm(name string, orgID, interval int, controllers []common.Controller) common.Farm {
	return &Farm{
		Name:        name,
		OrgID:       orgID,
		Interval:    interval,
		Controllers: controllers}
}

func (farm *Farm) SetID(id int) {
	farm.ID = id
}

func (farm *Farm) GetID() int {
	return farm.ID
}

func (farm *Farm) SetOrgID(id int) {
	farm.OrgID = id
}

func (farm *Farm) GetOrgID() int {
	return farm.OrgID
}

func (farm *Farm) SetMode(mode string) {
	farm.Mode = mode
}

func (farm *Farm) GetMode() string {
	return farm.Mode
}

func (farm *Farm) SetName(name string) {
	farm.Name = name
}

func (farm *Farm) GetName() string {
	return farm.Name
}

func (farm *Farm) SetInterval(interval int) {
	farm.Interval = interval
}

func (farm *Farm) GetInterval() int {
	return farm.Interval
}

func (farm *Farm) GetControllers() []common.Controller {
	return farm.Controllers
}

func (farm *Farm) SetControllers(controllers []common.Controller) {
	farm.Controllers = controllers
}

/*
func (farm *Farm) AddController(controller common.Controller) {
	farm.Controllers = append(farm.Controllers, controller)
}

func (farm *Farm) GetController(controllerType string) (common.Controller, error) {
	for _, controller := range farm.Controllers {
		if controller.GetType() == controllerType {
			return controller, nil
		}
	}
	return nil, fmt.Errorf("Controller type not found: %s", controllerType)
}

func (farm *Farm) GetControllerById(id int) (common.Controller, error) {
	farmSize := len(farm.Controllers)
	if farmSize < id {
		return nil, fmt.Errorf("Controller ID out of bounds: %d. Farm size: %d", id, farmSize)
	}
	return farm.Controllers[id], nil
}
*/
