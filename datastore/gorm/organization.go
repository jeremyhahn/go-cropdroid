package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormOrganizationDAO struct {
	logger      *logging.Logger
	db          *gorm.DB
	idGenerator util.IdGenerator
	dao.OrganizationDAO
	dao.GenericDAO[*config.Organization]
}

func NewOrganizationDAO(logger *logging.Logger, db *gorm.DB,
	idGenerator util.IdGenerator) dao.OrganizationDAO {

	return &GormOrganizationDAO{
		logger:      logger,
		db:          db,
		idGenerator: idGenerator}
}

func (dao *GormOrganizationDAO) Save(organization *config.OrganizationStruct) error {
	dao.logger.Debugf("Save GORM organization: %+v", organization)
	if organization.ID == 0 {
		id := dao.idGenerator.NewStringID(organization.GetName())
		organization.SetID(id)
	}
	return dao.db.
		Omit("Users").
		Omit("Permissions").
		Save(&organization).Error
}

func (dao *GormOrganizationDAO) Delete(organization *config.OrganizationStruct) error {
	dao.logger.Debugf("Delete GORM organization: %+v", organization)
	return dao.db.Delete(&organization).Error
}

func (dao *GormOrganizationDAO) Get(id uint64, CONSISTENCY_LEVEL int) (*config.OrganizationStruct, error) {
	dao.logger.Debugf("Fetching organization ID: %d", id)
	var org *config.OrganizationStruct
	if err := dao.db.
		Preload("Farms").
		Preload("Users").
		Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		Preload("Farms.Devices").
		Preload("Farms.Devices.Settings").
		Preload("Farms.Devices.Metrics").
		Preload("Farms.Devices.Channels").
		Preload("Farms.Devices.Channels.Conditions").
		Preload("Farms.Devices.Channels.Schedule").
		Preload("Farms.Workflows").
		Preload("Farms.Workflows.Conditions").
		Preload("Farms.Workflows.Schedules").
		Preload("Farms.Workflows.Steps").
		First(&org, id).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			dao.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		dao.logger.Error(err)
		return nil, err
	}
	return org, nil
}

func (organizationDAO *GormOrganizationDAO) GetPage(pageQuery query.PageQuery,
	CONSISTENCY_LEVEL int) (dao.PageResult[*config.OrganizationStruct], error) {

	organizationDAO.logger.Debugf("GetPage GORM organization: %+v", pageQuery)
	pageResult := dao.PageResult[*config.OrganizationStruct]{
		Page:     pageQuery.Page,
		PageSize: pageQuery.PageSize}
	page := pageQuery.Page
	if page < 1 {
		page = 1
	}
	var offset = (page - 1) * pageQuery.PageSize
	var orgs []*config.OrganizationStruct
	if err := organizationDAO.db.
		Preload("Farms").
		Preload("Users").
		Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		Preload("Farms.Devices").
		Preload("Farms.Devices.Settings").
		Preload("Farms.Devices.Metrics").
		Preload("Farms.Devices.Channels").
		Preload("Farms.Devices.Channels.Conditions").
		Preload("Farms.Devices.Channels.Schedule").
		Preload("Farms.Workflows").
		Preload("Farms.Workflows.Conditions").
		Preload("Farms.Workflows.Schedules").
		Preload("Farms.Workflows.Steps").
		Offset(offset).
		Limit(pageQuery.PageSize + 1). // peek one record to set HasMore flag
		Find(&orgs).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			organizationDAO.logger.Warning(err)
			return pageResult, datastore.ErrRecordNotFound
		}
		organizationDAO.logger.Error(err)
		return pageResult, err
	}
	//orgConfigs := make([]config.Organization, len(orgs))
	for _, org := range orgs {
		for _, farm := range org.GetFarms() {
			farm.ParseSettings()
		}
		//	orgConfigs[i] = org
	}
	// If the peek record was returned, set the HasMore flag and remove the +1 record
	if len(orgs) == pageQuery.PageSize+1 {
		pageResult.HasMore = true
		orgs = orgs[:len(orgs)-1]
	}
	pageResult.Entities = orgs
	return pageResult, nil
}

func (dao *GormOrganizationDAO) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*config.OrganizationStruct], CONSISTENCY_LEVEL int) error {

	dao.logger.Debugf("ForEachPage GORM organization: %T,  %+v", pagerProcFunc, pageQuery)
	pageResult, err := dao.GetPage(pageQuery, CONSISTENCY_LEVEL)
	if err != nil {
		dao.logger.Error(err)
		return nil
	}
	if err = pagerProcFunc(pageResult.Entities); err != nil {
		dao.logger.Error(err)
		return err
	}
	if pageResult.HasMore {
		nextPageQuery := query.PageQuery{
			Page:      pageQuery.Page + 1,
			PageSize:  pageQuery.PageSize,
			SortOrder: pageQuery.SortOrder}
		return dao.ForEachPage(nextPageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
	}
	return nil
}

func (dao *GormOrganizationDAO) GetUsers(orgID uint64) ([]*config.UserStruct, error) {
	dao.logger.Debugf("GetUsers GORM organization ID: %d", orgID)
	var org config.OrganizationStruct
	if err := dao.db.
		Preload("Users").
		Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		First(&org, orgID).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			dao.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		dao.logger.Error(err)
		return nil, err
	}
	return org.Users, nil
}

func (dao *GormOrganizationDAO) Count(CONSISTENCY_LEVEL int) (int64, error) {
	dao.logger.Debugf("Count GORM organizations")
	var count int64
	if err := dao.db.Model(&config.OrganizationStruct{}).Count(&count).Error; err != nil {
		dao.logger.Error(err)
		return 0, err
	}
	return count, nil
}
