package gorm

import (
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormOrganizationDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.OrganizationDAO
}

func NewOrganizationDAO(logger *logging.Logger, db *gorm.DB) dao.OrganizationDAO {
	return &GormOrganizationDAO{logger: logger, db: db}
}

func (dao *GormOrganizationDAO) Create(organization config.OrganizationConfig) error {
	dao.logger.Debugf("Creating organization record")
	return dao.db.Create(organization).Error
}

func (dao *GormOrganizationDAO) First() (config.OrganizationConfig, error) {
	dao.logger.Debugf("Updating organization record")
	var org config.Organization
	if err := dao.db.Preload("Farms").Preload("Users").Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		Preload("Farms.Controllers").
		Preload("Farms.Controllers.Configs").Preload("Farms.Controllers.Metrics").Preload("Farms.Controllers.Channels").
		Preload("Farms.Controllers.Channels.Conditions").
		Preload("Farms.Controllers.Channels.Schedule").
		First(&org).Error; err != nil {
		return nil, err
	}
	for i, farm := range org.GetFarms() {
		if err := farm.ParseConfigs(); err != nil {
			return nil, err
		}
		org.Farms[i] = farm
	}
	return &org, nil
}

func (dao *GormOrganizationDAO) Get(orgID int) (config.OrganizationConfig, error) {
	dao.logger.Debugf("Updating organization record")
	var org config.Organization
	if err := dao.db.Preload("Farms").Preload("Users").Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		Preload("Farms.Controllers").
		Preload("Farms.Controllers.Configs").Preload("Farms.Controllers.Metrics").Preload("Farms.Controllers.Channels").
		Preload("Farms.Controllers.Channels.Conditions").
		Preload("Farms.Controllers.Channels.Schedule").
		First(&org, orgID).Error; err != nil {
		return nil, err
	}
	for i, farm := range org.GetFarms() {
		farm.ParseConfigs()
		org.Farms[i] = farm
	}

	return &org, nil
}

func (dao *GormOrganizationDAO) GetAll() ([]config.Organization, error) {
	dao.logger.Debugf("Fetching all organizations")
	var orgs []config.Organization
	if err := dao.db.Preload("Farms").Preload("Users").Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		Preload("Farms.Controllers").
		Preload("Farms.Controllers.Configs").Preload("Farms.Controllers.Metrics").Preload("Farms.Controllers.Channels").
		Preload("Farms.Controllers.Channels.Conditions").
		Preload("Farms.Controllers.Channels.Schedule").
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	for i, org := range orgs {
		for j, farm := range org.GetFarms() {
			farm.ParseConfigs()
			orgs[i].Farms[j] = farm
		}
	}
	return orgs, nil
}

func (dao *GormOrganizationDAO) CreateUserRole(org config.OrganizationConfig, user config.UserConfig, role config.RoleConfig) error {
	return dao.db.Create(&config.Permission{
		UserID:         user.GetID(),
		RoleID:         role.GetID(),
		OrganizationID: org.GetID()}).Error
}

func (dao *GormOrganizationDAO) GetByUserID(userID int) ([]config.OrganizationConfig, error) {
	dao.logger.Debugf("Getting organization id for user.id %d", userID)
	var orgs []config.Organization
	if err := dao.db.
		Table("organizations").
		Preload("Farms").
		Select("organizations.id, organizations.name").
		Joins("JOIN permissions on organizations.id = permissions.organization_id AND permissions.user_id = ?", userID).
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	configs := make([]config.OrganizationConfig, len(orgs))
	for i, org := range orgs {
		configs[i] = &org
		for j, farm := range org.Farms {
			farm.ParseConfigs()
			orgs[i].Farms[j] = farm
		}
	}
	return configs, nil
}

/*
func (dao *GormOrganizationDAO) Find(orgID int) ([]config.Organization, error) {
	dao.logger.Debugf("Updating organization record")
	var orgs []config.Organization
	if err := dao.db.Preload("Farms").Find(&orgs, orgID).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}*/
