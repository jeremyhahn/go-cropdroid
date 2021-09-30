package gorm

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormFarmDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.FarmDAO
}

func NewFarmDAO(logger *logging.Logger, db *gorm.DB) dao.FarmDAO {
	return &GormFarmDAO{logger: logger, db: db}
}

func (dao *GormFarmDAO) Create(farm config.FarmConfig) error {
	dao.logger.Debugf("Creating farm record: %s", farm.GetName())
	return dao.db.Create(farm).Error
}

func (dao *GormFarmDAO) Delete(farm config.FarmConfig) error {

	dao.logger.Debugf("Deleting farm record: %s", farm.GetName())

	for _, device := range farm.GetDevices() {

		for _, channel := range device.GetChannels() {
			dao.db.Where("channel_id = ?", channel.GetID()).Delete(&config.Condition{})
			dao.db.Where("channel_id = ?", channel.GetID()).Delete(&config.Schedule{})
			// dao.db.Delete(&config.Condition{}).Where("channel_id = ?", channel.GetID())
			// dao.db.Delete(&config.Schedule{}).Where("channel_id = ?", channel.GetID())
		}
		dao.db.Where("device_id = ?", device.GetID()).Delete(&config.DeviceConfigItem{})
		dao.db.Where("device_id = ?", device.GetID()).Delete(&config.Channel{})
		dao.db.Where("device_id = ?", device.GetID()).Delete(&config.Metric{})
		dao.db.Where("device_id = ?", device.GetID()).Delete(&config.WorkflowStep{})
		//dao.db.Delete(&config.DeviceConfigItem{}).Where("device_id = ?", device.GetID())
		//dao.db.Delete(&config.Channel{}).Where("device_id = ?", device.GetID())
		//dao.db.Delete(&config.Metric{}).Where("device_id = ?", device.GetID())
		//dao.db.Delete(&config.WorkflowStep{}).Where("device_id = ?", device.GetID())

		dao.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS state_%d", device.GetID()))
	}

	dao.db.Where("farm_id = ?", farm.GetID()).Delete(&config.Device{})
	dao.db.Where("farm_id = ?", farm.GetID()).Delete(&config.Permission{})
	dao.db.Where("farm_id = ?", farm.GetID()).Delete(&config.Workflow{})

	// dao.db.Delete(&config.Device{}).Where("farm_id = ?", farm.GetID())
	// dao.db.Delete(&config.Permission{}).Where("farm_id = ?", farm.GetID())
	// dao.db.Delete(&config.Workflow{}).Where("farm_id = ?", farm.GetID())

	return dao.db.Delete(farm).Error
}

// func (dao *GormFarmDAO) DeleteById(farmID uint64) error {
// 	dao.logger.Debugf("Deleing farm record: %d", farmID)
// 	//	return dao.db.Delete(&config.Farm{}, farmID).Error
// 	model := dao.db.Model(farm).AddForeignKey("farm_id", "devices(farm_id)", "UPDATE", "UPDATE")
// 	return dao.db.Delete(model).Error
// }

func (dao *GormFarmDAO) Save(farm config.FarmConfig) error {
	dao.logger.Debugf("Saving farm record: %s", farm.GetName())
	return dao.db.Save(farm).Error
}

func (dao *GormFarmDAO) First() (config.FarmConfig, error) {
	dao.logger.Debugf("Getting first farm record")
	var farm config.Farm
	if err := dao.db.First(&farm).Error; err != nil {
		return nil, err
	}
	if err := farm.ParseConfigs(); err != nil {
		return nil, err
	}
	return &farm, nil
}

func (dao *GormFarmDAO) Get(farmID uint64, CONSISTENCY_LEVEL int) (config.FarmConfig, error) {
	dao.logger.Debugf("Getting farm: %d", farmID)
	var farm config.Farm
	if err := dao.db.
		Preload("Devices").
		Preload("Users").
		Preload("Users.Roles").
		Preload("Devices.Configs").
		Preload("Devices.Metrics").
		Preload("Devices.Channels").
		Preload("Devices.Channels.Conditions").
		Preload("Devices.Channels.Schedule").
		Preload("Workflows").
		Preload("Workflows.Conditions").
		Preload("Workflows.Schedules").
		Preload("Workflows.Steps").
		First(&farm, farmID).Error; err != nil {
		return nil, err
	}
	if err := farm.ParseConfigs(); err != nil {
		return nil, err
	}
	return &farm, nil
}

func (dao *GormFarmDAO) GetAll() ([]config.Farm, error) {
	dao.logger.Debug("Getting all farms")
	var farms []config.Farm
	if err := dao.db.
		Preload("Devices").
		Preload("Users").
		Preload("Users.Roles").
		Preload("Devices.Configs").
		Preload("Devices.Metrics").
		Preload("Devices.Channels").
		Preload("Devices.Channels.Conditions").
		Preload("Devices.Channels.Schedule").
		Preload("Workflows").
		Preload("Workflows.Conditions").
		Preload("Workflows.Schedules").
		Preload("Workflows.Steps").
		Find(&farms).Error; err != nil {
		return nil, err
	}
	for i, farm := range farms {
		farms[i] = farm
		if err := farms[i].ParseConfigs(); err != nil {
			return nil, err
		}
	}
	return farms, nil
}

func (dao *GormFarmDAO) GetByOrgAndUserID(orgID, userID uint64) ([]config.Farm, error) {
	dao.logger.Debug("Getting all farms for user: %d", userID)
	var farms []config.Farm
	if err := dao.db.
		Preload("Devices").
		Preload("Users").
		Preload("Users.Roles").
		Preload("Devices.Configs").
		Preload("Devices.Metrics").
		Preload("Devices.Channels").
		Preload("Devices.Channels.Conditions").
		Preload("Devices.Channels.Schedule").
		Preload("Workflows").
		Preload("Workflows.Conditions").
		Preload("Workflows.Schedules").
		Preload("Workflows.Steps").
		Joins("JOIN permissions on farms.organization_id = permissions.organization_id AND permissions.farm_id = farms.id").
		Where("permissions.organization_id = ? AND permissions.user_id = ?", orgID, userID).
		Find(&farms).Error; err != nil {
		return nil, err
	}
	for i, farm := range farms {
		farms[i] = farm
		if err := farms[i].ParseConfigs(); err != nil {
			return nil, err
		}
	}
	return farms, nil
}

func (dao *GormFarmDAO) Count() (int64, error) {
	dao.logger.Debugf("Getting farm count")
	var farm config.Farm
	var count int64
	if err := dao.db.Model(&farm).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

/*
func (dao *GormFarmDAO) Save(organization entity.FarmEntity) error {
	dao.app.Logger.Debugf("Saving organization record")
	return dao.db.Save(organization).Error
}

func (dao *GormFarmDAO) Update(organization entity.FarmEntity) error {
	dao.app.Logger.Debugf("Updating organization record")
	return dao.db.Update(organization).Error
}

func (dao *GormFarmDAO) Get(name string) (entity.FarmEntity, error) {
	dao.app.Logger.Debugf("Getting organization record '%s'", name)
	var Farms []entity.Farm
	if err := dao.db.Where("name = ?", name).Find(&Farms).Error; err != nil {
		return nil, err
	}
	if len(Farms) == 0 {
		return nil, errors.New(fmt.Sprintf("Farm name '%s' not found in database", name))
	}
	return &Farms[0], nil
}

func (dao *GormFarmDAO) GetByID(id string) (entity.FarmEntity, error) {
	dao.app.Logger.Debugf("Getting organization id '%s'", id)
	var Farms []entity.Farm
	if err := dao.db.Where("id = ?", id).Find(&Farms).Error; err != nil {
		return nil, err
	}
	if len(Farms) == 0 {
		return nil, errors.New(fmt.Sprintf("Farm id '%s' not found in database", id))
	}
	return &Farms[0], nil
}


func (dao *GormFarmDAO) GetByUserID(userID int) ([]entity.Farm, error) {
	dao.app.Logger.Debugf("Getting organization id for user %d", userID)
	var Farms []entity.Farm
	if err := dao.db.
		Table("organizations").
		Select("organizations.id, organizations.name").
		Joins("JOIN permissions on organizations.id = permissions.organization_id AND permissions.user_id = ?", userID).
		Find(&Farms).Error; err != nil {
		return nil, err
	}
	return Farms, nil
}
*/
