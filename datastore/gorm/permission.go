package gorm

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormPermissionDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.PermissionDAO
}

func NewPermissionDAO(logger *logging.Logger, db *gorm.DB) dao.PermissionDAO {
	return &GormPermissionDAO{logger: logger, db: db}
}

// This method returns a minimal depth organization with its associated farms and users.
// No device or workflows are returned.
func (permissionDAO *GormPermissionDAO) GetOrganizations(userID uint64,
	CONSISTENCY_LEVEL int) ([]*config.OrganizationStruct, error) {

	permissionDAO.logger.Debugf("Getting organizations for user.id %d", userID)
	var orgs []*config.OrganizationStruct
	//if shallow {
	if err := permissionDAO.db.
		Table("organizations").
		Preload("Farms").
		Preload("Users").
		Preload("Users.Roles").
		Select("organizations.id, organizations.name").
		Joins("JOIN permissions on organizations.id = permissions.organization_id AND permissions.user_id = ?", userID).
		Find(&orgs).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			permissionDAO.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		permissionDAO.logger.Error(err)
		return nil, err
	}
	// } else {
	// 	if err := permissionDAO.db.
	// 		Table("organizations").
	// 		Preload("Farms").
	// 		Preload("Users").
	// 		Preload("Users.Roles").
	// 		//Preload("Farms.Users").Preload("Farms.Users.Roles").
	// 		Preload("Farms.Devices").
	// 		Preload("Farms.Devices.Configs").
	// 		Preload("Farms.Devices.Metrics").
	// 		Preload("Farms.Devices.Channels").
	// 		Preload("Farms.Devices.Channels.Conditions").
	// 		Preload("Farms.Devices.Channels.Schedule").
	// 		Preload("Farms.Workflows").
	// 		Preload("Farms.Workflows.Conditions").
	// 		Preload("Farms.Workflows.Schedules").
	// 		Preload("Farms.Workflows.Steps").
	// 		Select("organizations.id, organizations.name").
	// 		Joins("JOIN permissions on organizations.id = permissions.organization_id AND permissions.user_id = ?", userID).
	// 		Find(&orgs).Error; err != nil {
	// 		return nil, err
	// 	}
	// }
	return orgs, nil
}

// Returns all users belonging to the specified organization id
func (permissionDAO *GormPermissionDAO) GetUsers(orgID uint64,
	CONSISTENCY_LEVEL int) ([]*config.UserStruct, error) {

	permissionDAO.logger.Debugf("Getting user for orgID: %d", orgID)
	userIDs := make(map[uint64]bool, 0)
	var users []*config.UserStruct
	var permissions []config.PermissionStruct
	if err := permissionDAO.db.
		Where("organization_id = ? AND farm_id > 0", orgID).
		Find(&permissions).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			permissionDAO.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		permissionDAO.logger.Error(err)
		return nil, err
	}
	for _, permission := range permissions {
		if permission.FarmID == 0 {
			// this is the default admin user permission
			continue
		}
		if _, exists := userIDs[permission.UserID]; !exists {
			var user config.UserStruct
			var roles []*config.RoleStruct
			if err := permissionDAO.db.Find(&user, permission.UserID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					permissionDAO.logger.Warning(err)
					return nil, datastore.ErrRecordNotFound
				}
				permissionDAO.logger.Error(err)
				return nil, err
			}
			if err := permissionDAO.db.Find(&roles, permission.RoleID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					permissionDAO.logger.Warning(err)
					return nil, datastore.ErrRecordNotFound
				}
				permissionDAO.logger.Error(err)
				return nil, err
			}
			user.SetRoles(roles)
			users = append(users, &user)
			userIDs[permission.UserID] = true
		}
	}
	return users, nil
}

func (permissionDAO *GormPermissionDAO) GetFarms(orgID uint64,
	CONSISTENCY_LEVEL int) ([]*config.FarmStruct, error) {

	var permissions []config.PermissionStruct
	farms := make(map[uint64]config.FarmStruct, 0)
	if err := permissionDAO.db.
		Where("organization_id = ? AND farm_id > 0", orgID).
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	for _, perm := range permissions {
		if _, ok := farms[perm.FarmID]; !ok {
			var farm config.FarmStruct
			if err := permissionDAO.db.First(&farm, perm.FarmID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					permissionDAO.logger.Warning(err)
					return nil, datastore.ErrRecordNotFound
				}
				permissionDAO.logger.Error(err)
				return nil, err
			}
			farms[perm.FarmID] = farm
		}
	}
	farmConfigs := make([]*config.FarmStruct, len(farms))
	i := 0
	for _, farm := range farms {
		farmConfigs[i] = &farm
		i++
	}
	return farmConfigs, nil
}

func (permissionDAO *GormPermissionDAO) Save(permission *config.PermissionStruct) error {
	permissionDAO.logger.Debugf(fmt.Sprintf("Saving permission record: %+v", permission))
	return permissionDAO.db.Save(permission).Error
}

func (permissionDAO *GormPermissionDAO) Update(permission *config.PermissionStruct) error {
	permissionDAO.logger.Debugf(fmt.Sprintf("Updating permission record: %+v", permission))
	return permissionDAO.db.Model(&config.PermissionStruct{}).
		Where("organization_id = ? AND farm_id = ? AND user_id = ?",
			permission.GetOrgID(), permission.GetFarmID(), permission.GetUserID()).
		Update("role_id", permission.GetRoleID()).Error
}

func (permissionDAO *GormPermissionDAO) Delete(permission *config.PermissionStruct) error {
	permissionDAO.logger.Debugf(fmt.Sprintf("Deleting permission record: %+v", permission))
	return permissionDAO.db.Model(&config.PermissionStruct{}).
		Where("organization_id = ? AND farm_id = ? AND user_id = ?",
			permission.GetOrgID(), permission.GetFarmID(), permission.GetUserID()).
		Delete(permission).Error
}

// func (permissionDAO *GormPermissionDAO) Get(id uint64,
// 	CONSISTENCY_LEVEL int) (config.Permission, error) {

// 	permissionDAO.logger.Debugf("Getting permission id: %d", id)
// 	var permission config.Permission
// 	if err := permissionDAO.db.First(&permission, id).Error; err != nil {
// 		return nil, err
// 	}
// 	return &permission, nil
// }
