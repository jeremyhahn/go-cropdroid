package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormConditionDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.ConditionDAO
}

func NewConditionDAO(logger *logging.Logger, db *gorm.DB) dao.ConditionDAO {
	return &GormConditionDAO{logger: logger, db: db}
}

func (dao *GormConditionDAO) Create(condition config.ConditionConfig) error {
	return dao.db.Create(condition).Error
}

func (dao *GormConditionDAO) Save(condition config.ConditionConfig) error {
	return dao.db.Save(condition).Error
}

func (dao *GormConditionDAO) Delete(condition config.ConditionConfig) error {
	return dao.db.Delete(condition).Error
}

func (dao *GormConditionDAO) Get(id uint64) (config.ConditionConfig, error) {
	var condition config.Condition
	if err := dao.db.First(&condition, id).Error; err != nil {
		return nil, err
	}
	return &condition, nil
}

func (dao *GormConditionDAO) GetByChannelID(id uint64) ([]config.Condition, error) {
	var entities []config.Condition
	if err := dao.db.Where("channel_id = ?", id).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (dao *GormConditionDAO) GetByOrgUserAndChannelID(orgID, userID, channelID uint64) ([]config.Condition, error) {
	dao.logger.Debugf("Getting conditions for orgID %d and channel %d", orgID, channelID)
	var entities []config.Condition
	if err := dao.db.Table("conditions").
		Select("conditions.*").
		Joins("JOIN channels on conditions.channel_id = channels.id").
		Joins("JOIN devices on devices.id = channels.device_id").
		Joins("JOIN farms on farms.id = devices.farm_id AND farms.organization_id = ?", orgID).
		Joins("JOIN permissions on farms.id = permissions.farm_id").
		Where("channels.id = ? and permissions.user_id = ?", channelID, userID).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

/*
func (dao *GormConditionDAO) Update(condition config.ConditionConfig) error {
	return dao.db.Update(condition).Error
}*/
