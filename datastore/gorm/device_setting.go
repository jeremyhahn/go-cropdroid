package gorm

import (
	"errors"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormDeviceSettingDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.DeviceSettingDAO
}

func NewDeviceSettingDAO(logger *logging.Logger, db *gorm.DB) dao.DeviceSettingDAO {
	return &GormDeviceSettingDAO{logger: logger, db: db}
}

// FarmID used for compatibility with Raft
func (configDAO *GormDeviceSettingDAO) Save(farmID uint64, config *config.DeviceSetting) error {
	configDAO.logger.Debugf("Saving config record")
	return configDAO.db.Save(config).Error
}

// FarmID and CONSISTENCY_LEVEL used for compatibility with Raft
func (configDAO *GormDeviceSettingDAO) Get(farmID, deviceID uint64, name string, CONSISTENCY_LEVEL int) (*config.DeviceSetting, error) {
	configDAO.logger.Debugf("Getting config record '%s'", name)
	var settings []config.DeviceSetting
	if err := configDAO.db.Where("device_id = ? AND key = ?",
		deviceID, name).Find(&settings).Error; err != nil {
		return nil, err
	}
	if len(settings) == 0 {
		return nil, errors.New(fmt.Sprintf("Config '%s' not found in database for device_id=%d", name, deviceID))
	}
	return &settings[0], nil
}

// func (configDAO *GormDeviceSettingDAO) GetAll(deviceID uint64) ([]config.DeviceConfigItem, error) {
// 	configDAO.logger.Debugf("Getting config record for device_id '%d'", deviceID)
// 	var configs []config.DeviceConfigItem
// 	if err := configDAO.db.Where("device_id = ?", deviceID).Order("device_id asc").Find(&configs).Error; err != nil {
// 		return nil, err
// 	}
// 	if len(configs) == 0 {
// 		return nil, fmt.Errorf("Device ID '%d' not found in configs table", deviceID)
// 	}
// 	/*
// 		configEntities := make([]config.DeviceConfigConfig, len(configs))
// 		for i, configEntity := range configs {
// 			var ce config.DeviceConfigConfig = &config.DeviceConfigItem{
// 				ID:           configEntity.GetID(),
// 				UserID:       configEntity.GetUserID(),
// 				OrgID:        configEntity.GetOrgID(),
// 				DeviceID: configEntity.GetDeviceID(),
// 				Key:          configEntity.GetKey(),
// 				Value:        configEntity.GetValue()}
// 			configEntities[i] = ce
// 		}
// 		return configEntities, nil
// 	*/
// 	return configs, nil
// }
