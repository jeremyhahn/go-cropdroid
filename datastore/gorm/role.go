package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormRoleDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.RoleDAO
}

func NewRoleDAO(logger *logging.Logger, db *gorm.DB) dao.RoleDAO {
	return &GormRoleDAO{logger: logger, db: db}
}

func (dao *GormRoleDAO) Create(role config.RoleConfig) error {
	return dao.db.Create(role).Error
}

func (dao *GormRoleDAO) Save(role config.RoleConfig) error {
	return dao.db.Save(role).Error
}

func (dao *GormRoleDAO) GetByUserAndOrgID(userID, orgID int) ([]config.Role, error) {
	dao.logger.Debugf("Getting role for user %d and org %d", userID, orgID)
	var roles []config.Role
	if err := dao.db.
		Table("roles").
		Select("roles.id, roles.name").
		Joins("JOIN permissions on roles.id = permissions.role_id AND permissions.user_id = ? and permissions.organization_id = ?", userID, orgID).
		Find(&roles).Error; err != nil {

		return nil, err
	}
	return roles, nil
}

func (dao *GormRoleDAO) GetByName(name string) (config.RoleConfig, error) {
	dao.logger.Debugf("Getting role %s", name)
	var role config.Role
	if err := dao.db.
		Table("roles").
		First(&role, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (dao *GormRoleDAO) GetAll() ([]config.RoleConfig, error) {
	dao.logger.Debug("Getting all roles")
	var roles []config.Role
	if err := dao.db.Order("name asc").Find(&roles).Error; err != nil {
		return nil, err
	}
	roleConfigs := make([]config.RoleConfig, len(roles))
	for i, role := range roles {
		roleConfig := new(config.Role)
		*roleConfig = role
		roleConfigs[i] = roleConfig
	}
	return roleConfigs, nil
}
