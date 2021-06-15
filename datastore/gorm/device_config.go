package gorm

import (
	"errors"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormDeviceConfigDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.DeviceConfigDAO
}

func NewDeviceConfigDAO(logger *logging.Logger, db *gorm.DB) dao.DeviceConfigDAO {
	return &GormDeviceConfigDAO{logger: logger, db: db}
}

func (configDAO *GormDeviceConfigDAO) Save(config config.DeviceConfigConfig) error {
	configDAO.logger.Debugf("Saving config record")
	return configDAO.db.Save(config).Error
}

func (configDAO *GormDeviceConfigDAO) Get(deviceID uint64, name string) (*config.DeviceConfigItem, error) {
	configDAO.logger.Debugf("Getting config record '%s'", name)
	var configs []config.DeviceConfigItem
	if err := configDAO.db.Where("device_id = ? AND key = ?", deviceID, name).Find(&configs).Error; err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, errors.New(fmt.Sprintf("Config '%s' not found in database for device_id=%d", name, deviceID))
	}
	return &configs[0], nil
}

func (configDAO *GormDeviceConfigDAO) GetAll(deviceID uint64) ([]config.DeviceConfigItem, error) {
	configDAO.logger.Debugf("Getting config record for device_id '%d'", deviceID)
	var configs []config.DeviceConfigItem
	if err := configDAO.db.Where("device_id = ?", deviceID).Order("device_id asc").Find(&configs).Error; err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("Device ID '%d' not found in configs table", deviceID)
	}
	/*
		configEntities := make([]config.DeviceConfigConfig, len(configs))
		for i, configEntity := range configs {
			var ce config.DeviceConfigConfig = &config.DeviceConfigItem{
				ID:           configEntity.GetID(),
				UserID:       configEntity.GetUserID(),
				OrgID:        configEntity.GetOrgID(),
				DeviceID: configEntity.GetDeviceID(),
				Key:          configEntity.GetKey(),
				Value:        configEntity.GetValue()}
			configEntities[i] = ce
		}
		return configEntities, nil
	*/
	return configs, nil
}
