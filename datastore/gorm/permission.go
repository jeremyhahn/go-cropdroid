package gorm

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormPermissionDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.PermissionDAO
}

func NewPermissionDAO(logger *logging.Logger, db *gorm.DB) dao.PermissionDAO {
	return &GormPermissionDAO{logger: logger, db: db}
}

// Returns all users belonging to the specified organization id
func (permissionDAO *GormPermissionDAO) GetUsers(orgID uint64) ([]config.UserConfig, error) {
	permissionDAO.logger.Debugf("Getting user for orgID: %d", orgID)
	userIDs := make(map[uint64]bool, 0)
	var users []config.UserConfig
	var permissions []config.Permission
	if err := permissionDAO.db.
		Where("organization_id = ? AND farm_id > 0", orgID).
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	for _, permission := range permissions {
		if permission.FarmID == 0 {
			// this is the default admin user permission
			continue
		}
		if _, exists := userIDs[permission.UserID]; !exists {
			var user config.User
			var roles []config.Role
			if err := permissionDAO.db.Find(&user, permission.UserID).Error; err != nil {
				return nil, err
			}
			if err := permissionDAO.db.Find(&roles, permission.RoleID).Error; err != nil {
				return nil, err
			}
			roleConfigs := make([]config.RoleConfig, len(roles))
			for i, role := range roles {
				roleConfigs[i] = &role
			}
			user.SetRoles(roleConfigs)
			users = append(users, &user)
			userIDs[permission.UserID] = true
		}
	}
	return users, nil
}

func (permissionDAO *GormPermissionDAO) GetFarms(orgID uint64) ([]config.FarmConfig, error) {
	var permissions []config.Permission
	farms := make(map[uint64]config.Farm, 0)
	if err := permissionDAO.db.
		Where("organization_id = ? AND farm_id > 0", orgID).
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	for _, perm := range permissions {
		if _, ok := farms[perm.FarmID]; !ok {
			var farm config.Farm
			if err := permissionDAO.db.First(&farm, perm.FarmID).Error; err != nil {
				return nil, err
			}
			farms[perm.FarmID] = farm
		}
	}
	farmConfigs := make([]config.FarmConfig, len(farms))
	i := 0
	for _, farm := range farms {
		farmConfigs[i] = &farm
		i++
	}
	return farmConfigs, nil
}

func (permissionDAO *GormPermissionDAO) Get(id uint64) (config.PermissionConfig, error) {
	permissionDAO.logger.Debugf("Getting permission id: %d", id)
	var permission config.Permission
	if err := permissionDAO.db.First(&permission, id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (permissionDAO *GormPermissionDAO) Save(permission config.PermissionConfig) error {
	perm := *permission.(*config.Permission)
	permissionDAO.logger.Debugf(fmt.Sprintf("Saving permission record: %+v", permission))
	return permissionDAO.db.Save(perm).Error
}

func (permissionDAO *GormPermissionDAO) Update(permission config.PermissionConfig) error {
	permissionDAO.logger.Debugf(fmt.Sprintf("Updating permission record: %+v", permission))
	return permissionDAO.db.Model(&config.Permission{}).
		Where("organization_id = ? AND farm_id = ? AND user_id = ?",
			permission.GetOrgID(), permission.GetFarmID(), permission.GetUserID()).
		Update("role_id", permission.GetRoleID()).Error
}

func (permissionDAO *GormPermissionDAO) Delete(permission config.PermissionConfig) error {
	permissionDAO.logger.Debugf(fmt.Sprintf("Deleting permission record: %+v", permission))
	return permissionDAO.db.Model(&config.Permission{}).
		Where("organization_id = ? AND farm_id = ? AND user_id = ?",
			permission.GetOrgID(), permission.GetFarmID(), permission.GetUserID()).
		Delete(permission).Error
}
