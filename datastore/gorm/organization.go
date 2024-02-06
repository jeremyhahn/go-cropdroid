package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormOrganizationDAO struct {
	logger      *logging.Logger
	db          *gorm.DB
	idGenerator util.IdGenerator
	dao.OrganizationDAO
}

func NewOrganizationDAO(logger *logging.Logger, db *gorm.DB,
	idGenerator util.IdGenerator) dao.OrganizationDAO {
	return &GormOrganizationDAO{
		logger:      logger,
		db:          db,
		idGenerator: idGenerator}
}

func (dao *GormOrganizationDAO) Save(organization *config.Organization) error {
	dao.logger.Debugf("Creating organization record")
	if organization.GetID() == 0 {
		id := dao.idGenerator.NewID(organization.GetName())
		organization.SetID(id)
	}
	return dao.db.
		Omit("Users").
		Omit("Permissions").
		Save(&organization).Error
}

func (dao *GormOrganizationDAO) Delete(organization *config.Organization) error {
	return dao.db.Delete(organization).Error
}

func (dao *GormOrganizationDAO) Get(id uint64, CONSISTENCY_LEVEL int) (*config.Organization, error) {
	dao.logger.Debugf("Fetching organization ID: %d", id)
	var org *config.Organization
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
		First(org, id).Error; err != nil {
		return nil, err
	}
	return org, nil
}

func (dao *GormOrganizationDAO) GetAll(CONSISTENCY_LEVEL int) ([]*config.Organization, error) {
	dao.logger.Debugf("Fetching all organizations")
	var orgs []*config.Organization
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
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	//orgConfigs := make([]config.Organization, len(orgs))
	for _, org := range orgs {
		for _, farm := range org.GetFarms() {
			farm.ParseSettings()
		}
		//	orgConfigs[i] = org
	}
	return orgs, nil
}

func (dao *GormOrganizationDAO) GetUsers(orgID uint64) ([]*config.User, error) {
	dao.logger.Debugf("Fetching users for organization ID: %d", orgID)
	var org config.Organization
	if err := dao.db.
		Preload("Users").
		Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		First(&org, orgID).Error; err != nil {
		return nil, err
	}
	return org.Users, nil
}

// func (dao *GormOrganizationDAO) First() (config.Organization, error) {
// 	dao.logger.Debugf("Updating organization record")
// 	var org config.Organization
// 	if err := dao.db.Preload("Farms").
// 		Preload("Users").
// 		Preload("Users.Roles").
// 		//Preload("Farms.Users").Preload("Farms.Users.Roles").
// 		Preload("Farms.Devices").
// 		Preload("Farms.Devices.Settings").
// 		Preload("Farms.Devices.Metrics").
// 		Preload("Farms.Devices.Channels").
// 		Preload("Farms.Devices.Channels.Conditions").
// 		Preload("Farms.Devices.Channels.Schedule").
// 		Preload("Farms.Workflows").
// 		Preload("Farms.Workflows.Conditions").
// 		Preload("Farms.Workflows.Schedules").
// 		Preload("Farms.Workflows.Steps").
// 		First(&org).Error; err != nil {
// 		return nil, err
// 	}
// 	for i, farm := range org.GetFarms() {
// 		if err := farm.ParseConfigs(); err != nil {
// 			return nil, err
// 		}
// 		org.Farms[i] = *farm.(*config.Farm)
// 	}
// 	return &org, nil
// }
