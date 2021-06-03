package gorm

import (
	"errors"
	"fmt"

	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormControllerConfigDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.ControllerConfigDAO
}

func NewControllerConfigDAO(logger *logging.Logger, db *gorm.DB) dao.ControllerConfigDAO {
	return &GormControllerConfigDAO{logger: logger, db: db}
}

func (configDAO *GormControllerConfigDAO) Save(config config.ControllerConfigConfig) error {
	configDAO.logger.Debugf("Saving config record")
	return configDAO.db.Save(config).Error
}

func (configDAO *GormControllerConfigDAO) Get(controllerID int, name string) (*config.ControllerConfigItem, error) {
	configDAO.logger.Debugf("Getting config record '%s'", name)
	var configs []config.ControllerConfigItem
	if err := configDAO.db.Where("controller_id = ? AND key = ?", controllerID, name).Find(&configs).Error; err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, errors.New(fmt.Sprintf("Config '%s' not found in database for controller_id=%d", name, controllerID))
	}
	return &configs[0], nil
}

func (configDAO *GormControllerConfigDAO) GetAll(controllerID int) ([]config.ControllerConfigItem, error) {
	configDAO.logger.Debugf("Getting config record for controller_id '%d'", controllerID)
	var configs []config.ControllerConfigItem
	if err := configDAO.db.Where("controller_id = ?", controllerID).Order("controller_id asc").Find(&configs).Error; err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("Controller ID '%d' not found in configs table", controllerID)
	}
	/*
		configEntities := make([]config.ControllerConfigConfig, len(configs))
		for i, configEntity := range configs {
			var ce config.ControllerConfigConfig = &config.ControllerConfigItem{
				ID:           configEntity.GetID(),
				UserID:       configEntity.GetUserID(),
				OrgID:        configEntity.GetOrgID(),
				ControllerID: configEntity.GetControllerID(),
				Key:          configEntity.GetKey(),
				Value:        configEntity.GetValue()}
			configEntities[i] = ce
		}
		return configEntities, nil
	*/
	return configs, nil
}
