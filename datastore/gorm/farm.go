package gorm

import (
	"fmt"
	"sort"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormFarmDAO struct {
	logger      *logging.Logger
	db          *gorm.DB
	idGenerator util.IdGenerator
	dao.FarmDAO
}

func NewFarmDAO(logger *logging.Logger, db *gorm.DB,
	idGenerator util.IdGenerator) dao.FarmDAO {

	return &GormFarmDAO{
		logger:      logger,
		db:          db,
		idGenerator: idGenerator}
}

func (dao *GormFarmDAO) Delete(farm *config.Farm) error {
	dao.logger.Debugf("Deleting farm record: %s", farm.GetName())
	for _, device := range farm.GetDevices() {
		for _, channel := range device.GetChannels() {
			dao.db.Where("channel_id = ?", channel.GetID()).Delete(&config.Condition{})
			dao.db.Where("channel_id = ?", channel.GetID()).Delete(&config.Schedule{})
		}
		dao.db.Where("device_id = ?", device.GetID()).Delete(&config.DeviceSetting{})
		dao.db.Where("device_id = ?", device.GetID()).Delete(&config.Channel{})
		dao.db.Where("device_id = ?", device.GetID()).Delete(&config.Metric{})
		dao.db.Where("device_id = ?", device.GetID()).Delete(&config.WorkflowStep{})
		dao.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS state_%d", device.GetID()))
	}
	dao.db.Where("farm_id = ?", farm.GetID()).Delete(&config.Device{})
	dao.db.Where("farm_id = ?", farm.GetID()).Delete(&config.Permission{})
	dao.db.Where("farm_id = ?", farm.GetID()).Delete(&config.Workflow{})
	return dao.db.Delete(farm).Error
}

func (dao *GormFarmDAO) Save(farm *config.Farm) error {
	dao.logger.Debugf("Saving farm record: %s", farm.GetName())
	if farm.GetID() == 0 {
		farm.SetID(dao.idGenerator.NewStringID(farm.GetName()))
	}
	if err := dao.db.Save(farm).Error; err != nil {
		return err
	}
	return nil
}

func (dao *GormFarmDAO) Get(farmID uint64, CONSISTENCY_LEVEL int) (*config.Farm, error) {
	dao.logger.Debugf("Getting farm: %d", farmID)
	var farm *config.Farm
	if err := dao.db.
		Preload("Devices").
		Preload("Users").
		Preload("Users.Roles").
		Preload("Devices.Settings").
		Preload("Devices.Metrics").
		Preload("Devices.Channels").
		Preload("Devices.Channels.Conditions").
		Preload("Devices.Channels.Schedule").
		Preload("Workflows").
		Preload("Workflows.Conditions").
		Preload("Workflows.Schedules").
		Preload("Workflows.Steps").
		First(&farm, farmID).Error; err != nil {
		return farm, err
	}
	if farm.ID == 0 {
		return farm, datastore.ErrNotFound
	}
	if err := farm.ParseSettings(); err != nil {
		return farm, err
	}
	for _, workflow := range farm.GetWorkflows() {
		workflowSteps := workflow.GetSteps()
		sort.SliceStable(workflowSteps, func(i, j int) bool {
			return workflowSteps[i].GetSortOrder() < workflowSteps[j].GetSortOrder()
		})
	}
	return farm, nil
}

func (dao *GormFarmDAO) GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) ([]*config.Farm, error) {
	dao.logger.Debugf("Getting farms: %+v", farmIds)
	var farms []*config.Farm
	if err := dao.db.
		Preload("Devices").
		Preload("Users").
		Preload("Users.Roles").
		Preload("Devices.Settings").
		Preload("Devices.Metrics").
		Preload("Devices.Channels").
		Preload("Devices.Channels.Conditions").
		Preload("Devices.Channels.Schedule").
		Preload("Workflows").
		Preload("Workflows.Conditions").
		Preload("Workflows.Schedules").
		Preload("Workflows.Steps").
		Where("id IN (?)", farmIds).
		Find(&farms).Error; err != nil {
		return nil, err
	}
	for _, farm := range farms {
		if err := farm.ParseSettings(); err != nil {
			return nil, err
		}
		for _, workflow := range farm.GetWorkflows() {
			workflowSteps := workflow.GetSteps()
			sort.SliceStable(workflowSteps, func(i, j int) bool {
				return workflowSteps[i].GetSortOrder() < workflowSteps[j].GetSortOrder()
			})
		}
	}
	return farms, nil
}

func (dao *GormFarmDAO) GetAll(CONSISTENCY_LEVEL int) ([]*config.Farm, error) {
	dao.logger.Debug("Getting all farms")
	var farms []*config.Farm
	if err := dao.db.
		Preload("Devices").
		Preload("Users").
		Preload("Users.Roles").
		Preload("Devices.Settings").
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
	for _, farm := range farms {
		if err := farm.ParseSettings(); err != nil {
			return nil, err
		}
		for _, workflow := range farm.GetWorkflows() {
			workflowSteps := workflow.GetSteps()
			sort.SliceStable(workflowSteps, func(i, j int) bool {
				return workflowSteps[i].GetSortOrder() < workflowSteps[j].GetSortOrder()
			})
		}
	}
	return farms, nil
}

func (dao *GormFarmDAO) GetByUserID(userID uint64, CONSISTENCY_LEVEL int) ([]*config.Farm, error) {
	dao.logger.Debug("Getting all farms for user: %d", userID)
	var farms []*config.Farm
	if err := dao.db.
		Preload("Devices").
		Preload("Users").
		Preload("Users.Roles").
		Preload("Devices.Settings").
		Preload("Devices.Metrics").
		Preload("Devices.Channels").
		Preload("Devices.Channels.Conditions").
		Preload("Devices.Channels.Schedule").
		Preload("Workflows").
		Preload("Workflows.Conditions").
		Preload("Workflows.Schedules").
		Preload("Workflows.Steps").
		Joins("JOIN permissions on permissions.farm_id = farms.id").
		Where("permissions.user_id = ?", userID).
		Find(&farms).Error; err != nil {
		return nil, err
	}
	for _, farm := range farms {
		if err := farm.ParseSettings(); err != nil {
			return nil, err
		}
		for _, workflow := range farm.GetWorkflows() {
			workflowSteps := workflow.GetSteps()
			sort.SliceStable(workflowSteps, func(i, j int) bool {
				return workflowSteps[i].GetSortOrder() < workflowSteps[j].GetSortOrder()
			})
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

// func (dao *GormFarmDAO) First() (config.Farm, error) {
// 	dao.logger.Debugf("Getting first farm record")
// 	var farm config.Farm
// 	if err := dao.db.First(&farm).Error; err != nil {
// 		return nil, err
// 	}
// 	if err := farm.ParseConfigs(); err != nil {
// 		return nil, err
// 	}
// 	return &farm, nil
// }

// func (dao *GormFarmDAO) Create(farm config.Farm) error {
// 	dao.logger.Debugf("Creating farm record: %s", farm.GetName())
// 	return dao.db.Create(farm).Error
// }

// func (dao *GormFarmDAO) DeleteById(farmID uint64) error {
// 	dao.logger.Debugf("Deleing farm record: %d", farmID)
// 	//	return dao.db.Delete(&config.Farm{}, farmID).Error
// 	model := dao.db.Model(farm).AddForeignKey("farm_id", "devices(farm_id)", "UPDATE", "UPDATE")
// 	return dao.db.Delete(model).Error
// }
