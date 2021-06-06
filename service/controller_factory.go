package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

// "Global" controller service used to manage all controllers used by the platform
type ControllerFactory interface {
	GetAll(session Session) ([]config.Controller, error)
	GetControllers(session Session) ([]common.Controller, error)
	//BuildController(name string) (common.Controller, error)
}

type MicroCntrollerFactory struct {
	dao    dao.ControllerDAO
	mapper mapper.ControllerMapper
}

func NewControllerFactory(dao dao.ControllerDAO, mapper mapper.ControllerMapper) ControllerFactory {
	return &MicroCntrollerFactory{dao: dao, mapper: mapper}
}

func (service *MicroCntrollerFactory) GetAll(session Session) ([]config.Controller, error) {
	controllerEntities, err := service.dao.GetByFarmId(session.GetFarmService().GetConfig().GetID())
	if err != nil {
		return nil, err
	}
	return controllerEntities, nil
}

func (service *MicroCntrollerFactory) GetControllers(session Session) ([]common.Controller, error) {
	var controllers []common.Controller
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	controllerEntities, err := service.dao.GetByFarmId(farmConfig.GetID())
	if err != nil {
		return nil, err
	}
	//controllers := make([]common.Controller, len(controllerEntities)-1) // -1 for server controller
	for _, entity := range controllerEntities {
		if entity.GetType() == common.CONTROLLER_TYPE_SERVER {
			continue
		}
		controllerState, err := farmService.GetState().GetController(entity.GetType())
		if err != nil {
			return nil, err
		}
		controllerConfig, err := farmConfig.GetController(entity.GetType())
		if err != nil {
			return nil, err
		}
		controller, err := service.mapper.MapStateToController(controllerState, *controllerConfig)
		if err != nil {
			return nil, err
		}
		//controllers[i] = controller
		controllers = append(controllers, controller) // dynamically add - not sure how many server controllers there will be
	}
	return controllers, nil
}
