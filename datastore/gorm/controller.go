package gorm

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormControllerDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.ControllerDAO
}

func NewControllerDAO(logger *logging.Logger, db *gorm.DB) dao.ControllerDAO {
	return &GormControllerDAO{logger: logger, db: db}
}

func (dao *GormControllerDAO) Save(controller config.ControllerConfig) error {
	dao.logger.Debugf("Creating controller record")
	return dao.db.Save(controller).Error
}

func (dao *GormControllerDAO) GetByFarmId(farmID int) ([]config.Controller, error) {
	dao.logger.Debugf("Getting controllers for farm id %d", farmID)
	var controllers []config.Controller
	if err := dao.db.Preload("Configs").Preload("Metrics").Preload("Channels").
		Where("farm_id = ?", farmID).Order("id asc").Find(&controllers).Error; err != nil {

		return nil, err
	}
	if len(controllers) == 0 {
		return nil, fmt.Errorf("Unable to locate controllers belonging to farm id %d", farmID)
	}
	for i, controller := range controllers {
		controller.ParseConfigs()
		controllers[i] = controller
	}
	return controllers, nil
}

/*
func (dao *GormControllerDAO) GetByOrgId(orgId int) ([]config.Controller, error) {
	dao.logger.Debugf("Getting controllers for org id %d", orgId)
	var controllers []config.Controller
	if err := dao.db.Where("organization_id = ?", orgId).Order("id asc").Find(&controllers).Error; err != nil {
		return nil, err
	}
	if len(controllers) == 0 {
		return nil, fmt.Errorf("Unable to locate controllers for organization id %d", orgId)
	}
	return controllers, nil
}

func (dao *GormControllerDAO) Create(controller entity.ControllerEntity) error {
	dao.app.Logger.Debugf("Saving controller record")
	return dao.db.Create(controller).Error
}

func (dao *GormControllerDAO) Update(controller entity.ControllerEntity) error {
	dao.app.Logger.Debugf("Updating controller record")
	return dao.db.Update(controller).Error
}

func (dao *GormControllerDAO) Get(id int) (entity.ControllerEntity, error) {
	dao.app.Logger.Debugf("Getting controller record %s", id)
	var controllers entity.Controller
	if err := dao.db.First(&controllers, id).Error; err != nil {
		return nil, err
	}
	return &controllers, nil
}

func (dao *GormControllerDAO) GetByOrgAndType(orgId int, controllerType string) ([]entity.Controller, error) {
	dao.app.Logger.Debugf("Getting %s controller for org id %d", controllerType, orgId)
	var controllers []entity.Controller
	if err := dao.db.Where("organization_id = ? AND type = ?", orgId, controllerType).Order("id asc").Find(&controllers).Error; err != nil {
		return nil, err
	}
	if len(controllers) == 0 {
		return nil, fmt.Errorf("Unable to locate controllers for organization id %d with type %s", orgId, controllerType)
	}
	return controllers, nil
}
*/
