package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormDeviceDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.DeviceDAO
}

func NewDeviceDAO(logger *logging.Logger, db *gorm.DB) dao.DeviceDAO {
	return &GormDeviceDAO{logger: logger, db: db}
}

func (dao *GormDeviceDAO) Save(device *config.Device) error {
	dao.logger.Debugf("Creating device record")
	return dao.db.Save(device).Error
}

// FarmID is used to maintain interface compatibility with Raft
func (dao *GormDeviceDAO) Get(farmID, deviceID uint64,
	CONSISTENCY_LEVEL int) (*config.Device, error) {

	dao.logger.Debugf("Getting device %d", deviceID)
	var devices []*config.Device
	if err := dao.db.
		Preload("Settings").
		Preload("Metrics").
		Preload("Channels").
		Preload("Channels.Conditions").
		Preload("Channels.Schedule").
		First(&devices, deviceID).Error; err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, datastore.ErrNotFound
	}
	if err := devices[0].ParseSettings(); err != nil {
		return nil, err
	}
	return devices[0], nil
}

// func (dao *GormDeviceDAO) GetByOrgAndFarmID(orgID, farmID uint64) ([]config.Device, error) {
// 	dao.logger.Debugf("Getting devices for farm id %d", farmID)
// 	var devices []config.Device
// 	if err := dao.db.
// 		Preload("Settings").
// 		Preload("Metrics").
// 		Preload("Channels").
// 		Where("farm_id = ?", farmID).
// 		Order("id asc").
// 		Find(&devices).Error; err != nil {
// 		return nil, err
// 	}
// 	if len(devices) == 0 {
// 		return nil, fmt.Errorf("Unable to locate devices belonging to farm id %d", farmID)
// 	}
// 	for i, device := range devices {
// 		device.ParseSettings()
// 		devices[i] = device
// 	}
// 	return devices, nil
// }

// func (dao *GormDeviceDAO) Count() (int64, error) {
// 	dao.logger.Debugf("Getting device count")
// 	var device config.Device
// 	var count int64
// 	if err := dao.db.Model(&device).Count(&count).Error; err != nil {
// 		return 0, err
// 	}
// 	return count, nil
// }

/*
func (dao *GormDeviceDAO) GetByOrgId(orgId int) ([]config.Device, error) {
	dao.logger.Debugf("Getting devices for org id %d", orgId)
	var devices []config.Device
	if err := dao.db.Where("organization_id = ?", orgId).Order("id asc").Find(&devices).Error; err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("Unable to locate devices for organization id %d", orgId)
	}
	return devices, nil
}

func (dao *GormDeviceDAO) Create(device entity.DeviceEntity) error {
	dao.app.Logger.Debugf("Saving device record")
	return dao.db.Create(device).Error
}

func (dao *GormDeviceDAO) Update(device entity.DeviceEntity) error {
	dao.app.Logger.Debugf("Updating device record")
	return dao.db.Update(device).Error
}

func (dao *GormDeviceDAO) Get(id int) (entity.DeviceEntity, error) {
	dao.app.Logger.Debugf("Getting device record %s", id)
	var devices entity.Device
	if err := dao.db.First(&devices, id).Error; err != nil {
		return nil, err
	}
	return &devices, nil
}

func (dao *GormDeviceDAO) GetByOrgAndType(orgId int, deviceType string) ([]entity.Device, error) {
	dao.app.Logger.Debugf("Getting %s device for org id %d", deviceType, orgId)
	var devices []entity.Device
	if err := dao.db.Where("organization_id = ? AND type = ?", orgId, deviceType).Order("id asc").Find(&devices).Error; err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("Unable to locate devices for organization id %d with type %s", orgId, deviceType)
	}
	return devices, nil
}
*/
