package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
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

func (dao *GormOrganizationDAO) Save(organization config.OrganizationConfig) error {
	dao.logger.Debugf("Creating organization record")
	return dao.db.Omit("Users").Save(organization).Error
}

func (dao *GormOrganizationDAO) Delete(organization config.OrganizationConfig) error {
	return dao.db.Delete(organization).Error
}

func (dao *GormOrganizationDAO) First() (config.OrganizationConfig, error) {
	dao.logger.Debugf("Updating organization record")
	var org config.Organization
	if err := dao.db.Preload("Farms").
		Preload("Users").
		Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		Preload("Farms.Devices").
		Preload("Farms.Devices.Configs").
		Preload("Farms.Devices.Metrics").
		Preload("Farms.Devices.Channels").
		Preload("Farms.Devices.Channels.Conditions").
		Preload("Farms.Devices.Channels.Schedule").
		Preload("Farms.Workflows").
		Preload("Farms.Workflows.Conditions").
		Preload("Farms.Workflows.Schedules").
		Preload("Farms.Workflows.Steps").
		First(&org).Error; err != nil {
		return nil, err
	}
	for i, farm := range org.GetFarms() {
		if err := farm.ParseConfigs(); err != nil {
			return nil, err
		}
		org.Farms[i] = *farm.(*config.Farm)
	}
	return &org, nil
}

func (dao *GormOrganizationDAO) GetAll() ([]config.OrganizationConfig, error) {
	dao.logger.Debugf("Fetching all organizations")
	var orgs []config.Organization
	if err := dao.db.
		Preload("Farms").
		Preload("Users").
		Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		Preload("Farms.Devices").
		Preload("Farms.Devices.Configs").
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
	orgConfigs := make([]config.OrganizationConfig, len(orgs))
	for i, org := range orgs {
		for j, farm := range org.GetFarms() {
			farm.ParseConfigs()
			orgs[i].Farms[j] = *farm.(*config.Farm)
		}
		orgConfig := new(config.Organization)
		*orgConfig = org
		orgConfigs[i] = orgConfig
	}
	return orgConfigs, nil
}

func (dao *GormOrganizationDAO) Get(id uint64) (config.OrganizationConfig, error) {
	dao.logger.Debugf("Fetching organization ID: %d", id)
	var org config.Organization
	if err := dao.db.
		Preload("Farms").
		Preload("Users").
		Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		Preload("Farms.Devices").
		Preload("Farms.Devices.Configs").
		Preload("Farms.Devices.Metrics").
		Preload("Farms.Devices.Channels").
		Preload("Farms.Devices.Channels.Conditions").
		Preload("Farms.Devices.Channels.Schedule").
		Preload("Farms.Workflows").
		Preload("Farms.Workflows.Conditions").
		Preload("Farms.Workflows.Schedules").
		Preload("Farms.Workflows.Steps").
		First(&org, id).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// This method returns a minimal depth organization with its associated farms and users.
// No device or workflows are returned.
func (dao *GormOrganizationDAO) GetByUserID(userID uint64, shallow bool) ([]config.OrganizationConfig, error) {
	dao.logger.Debugf("Getting organizations for user.id %d", userID)
	var orgs []config.Organization
	if shallow {
		if err := dao.db.
			Table("organizations").
			Preload("Farms").
			Preload("Users").
			Preload("Users.Roles").
			Select("organizations.id, organizations.name").
			Joins("JOIN permissions on organizations.id = permissions.organization_id AND permissions.user_id = ?", userID).
			Find(&orgs).Error; err != nil {
			return nil, err
		}
	} else {
		if err := dao.db.
			Table("organizations").
			Preload("Farms").
			Preload("Users").
			Preload("Users.Roles").
			//Preload("Farms.Users").Preload("Farms.Users.Roles").
			Preload("Farms.Devices").
			Preload("Farms.Devices.Configs").
			Preload("Farms.Devices.Metrics").
			Preload("Farms.Devices.Channels").
			Preload("Farms.Devices.Channels.Conditions").
			Preload("Farms.Devices.Channels.Schedule").
			Preload("Farms.Workflows").
			Preload("Farms.Workflows.Conditions").
			Preload("Farms.Workflows.Schedules").
			Preload("Farms.Workflows.Steps").
			Select("organizations.id, organizations.name").
			Joins("JOIN permissions on organizations.id = permissions.organization_id AND permissions.user_id = ?", userID).
			Find(&orgs).Error; err != nil {
			return nil, err
		}
	}
	configs := make([]config.OrganizationConfig, len(orgs))
	for i, org := range orgs {
		for j, farm := range org.Farms {
			farm.ParseConfigs()
			orgs[i].Farms[j] = farm
		}
		orgConfig := new(config.Organization)
		*orgConfig = org
		configs[i] = orgConfig
	}
	return configs, nil
}

func (dao *GormOrganizationDAO) GetUsers(id uint64) ([]config.UserConfig, error) {
	dao.logger.Debugf("Fetching users for organization ID: %d", id)
	var org config.Organization
	if err := dao.db.
		Preload("Users").
		Preload("Users.Roles").
		//Preload("Farms.Users").Preload("Farms.Users.Roles").
		First(&org, id).Error; err != nil {
		return nil, err
	}
	users := make([]config.UserConfig, len(org.Users))
	for i, user := range org.Users {
		userConfig := new(config.User)
		*userConfig = user
		users[i] = userConfig
	}
	return users, nil
}
