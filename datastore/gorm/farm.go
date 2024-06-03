package gorm

import (
	"fmt"
	"sort"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
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

func (farmDAO *GormFarmDAO) Delete(farm *config.FarmStruct) error {
	farmDAO.logger.Debugf("Deleting farm record: %s", farm.GetName())
	for _, device := range farm.GetDevices() {
		for _, channel := range device.GetChannels() {
			farmDAO.db.Where("channel_id = ?", channel.ID).Delete(&config.ConditionStruct{})
			farmDAO.db.Where("channel_id = ?", channel.ID).Delete(&config.ScheduleStruct{})
		}
		farmDAO.db.Where("device_id = ?", device.ID).Delete(&config.DeviceSettingStruct{})
		farmDAO.db.Where("device_id = ?", device.ID).Delete(&config.ChannelStruct{})
		farmDAO.db.Where("device_id = ?", device.ID).Delete(&config.MetricStruct{})
		farmDAO.db.Where("device_id = ?", device.ID).Delete(&config.WorkflowStepStruct{})
		farmDAO.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS state_%d", device.ID))
	}
	farmDAO.db.Where("farm_id = ?", farm.ID).Delete(&config.DeviceStruct{})
	farmDAO.db.Where("farm_id = ?", farm.ID).Delete(&config.PermissionStruct{})
	farmDAO.db.Where("farm_id = ?", farm.ID).Delete(&config.WorkflowStruct{})
	return farmDAO.db.Delete(farm).Error
}

func (farmDAO *GormFarmDAO) Save(farm *config.FarmStruct) error {
	farmDAO.logger.Debugf("Saving farm record: %s", farm.GetName())
	if farm.ID == 0 {
		farm.SetID(farmDAO.idGenerator.NewStringID(farm.GetName()))
	}
	if err := farmDAO.db.Save(farm).Error; err != nil {
		return err
	}
	return nil
}

func (farmDAO *GormFarmDAO) Get(farmID uint64, CONSISTENCY_LEVEL int) (*config.FarmStruct, error) {
	farmDAO.logger.Debugf("Getting farm: %d", farmID)
	var farm *config.FarmStruct
	if err := farmDAO.db.
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

		if err == gorm.ErrRecordNotFound {
			farmDAO.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		farmDAO.logger.Error(err)
		return farm, err
	}
	if farm.ID == 0 {
		return farm, datastore.ErrRecordNotFound
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

func (farmDAO *GormFarmDAO) GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) ([]*config.FarmStruct, error) {
	farmDAO.logger.Debugf("Getting farms: %+v", farmIds)
	var farms []*config.FarmStruct
	if err := farmDAO.db.
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

		if err == gorm.ErrRecordNotFound {
			farmDAO.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		farmDAO.logger.Error(err)
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

func (farmDAO *GormFarmDAO) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.FarmStruct], error) {
	farmDAO.logger.Debug("Getting farm page %+v", pageQuery)
	pageResult := dao.PageResult[*config.FarmStruct]{
		Page:     pageQuery.Page,
		PageSize: pageQuery.PageSize}
	page := pageQuery.Page
	if page < 1 {
		page = 1
	}
	var offset = (page - 1) * pageQuery.PageSize
	var farms []*config.FarmStruct
	if err := farmDAO.db.
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
		Offset(offset).
		Limit(pageQuery.PageSize + 1). // peek one record to set HasMore flag
		Find(&farms).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			farmDAO.logger.Warning(err)
			return pageResult, datastore.ErrRecordNotFound
		}
		farmDAO.logger.Error(err)
		return pageResult, err
	}
	for _, farm := range farms {
		if err := farm.ParseSettings(); err != nil {
			return pageResult, err
		}
		for _, workflow := range farm.GetWorkflows() {
			workflowSteps := workflow.GetSteps()
			sort.SliceStable(workflowSteps, func(i, j int) bool {
				return workflowSteps[i].GetSortOrder() < workflowSteps[j].GetSortOrder()
			})
		}
	}
	// If the peek record was returned, set the HasMore flag and remove the +1 record
	if len(farms) == pageQuery.PageSize+1 {
		pageResult.HasMore = true
		farms = farms[:len(farms)-1]
	}
	pageResult.Entities = farms
	return pageResult, nil
}

func (farmDAO *GormFarmDAO) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*config.FarmStruct], CONSISTENCY_LEVEL int) error {

	pageResult, err := farmDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
	if err != nil {
		farmDAO.logger.Error(err)
		return nil
	}
	if err = pagerProcFunc(pageResult.Entities); err != nil {
		farmDAO.logger.Error(err)
		return err
	}
	if pageResult.HasMore {
		nextPageQuery := query.PageQuery{
			Page:      pageQuery.Page + 1,
			PageSize:  pageQuery.PageSize,
			SortOrder: pageQuery.SortOrder}
		return farmDAO.ForEachPage(nextPageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
	}
	return nil
}

func (farmDAO *GormFarmDAO) GetByUserID(userID uint64, CONSISTENCY_LEVEL int) ([]*config.FarmStruct, error) {
	farmDAO.logger.Debug("Getting all farms for user: %d", userID)
	var farms []*config.FarmStruct
	if err := farmDAO.db.
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
		if err == gorm.ErrRecordNotFound {
			farmDAO.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		farmDAO.logger.Error(err)
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

func (farmDAO *GormFarmDAO) Count(CONSISTENCY_LEVEL int) (int64, error) {
	farmDAO.logger.Debugf("Getting farm count")
	var farm config.Farm
	var count int64
	if err := farmDAO.db.Model(&farm).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
