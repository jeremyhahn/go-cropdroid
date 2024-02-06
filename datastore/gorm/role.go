package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormRoleDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.RoleDAO
}

func NewRoleDAO(logger *logging.Logger, db *gorm.DB) dao.RoleDAO {
	return &GormRoleDAO{logger: logger, db: db}
}

func (dao *GormRoleDAO) Save(role *config.Role) error {
	return dao.db.Save(&role).Error
}

func (dao *GormRoleDAO) Delete(role *config.Role) error {
	return dao.db.Delete(role).Error
}

// This method is only here for sake of completeness for the interface. This method
// is only used by the Raft datastore.
func (dao *GormRoleDAO) Get(roleID uint64, CONSISTENCY_LEVEL int) (*config.Role, error) {
	dao.logger.Debugf("Getting role %s", roleID)
	var role *config.Role
	if err := dao.db.
		First(&role, roleID).Error; err != nil {
		dao.logger.Errorf("[RoleDAO.Get] %s", err.Error())
		return nil, err
	}
	return role, nil
}

func (dao *GormRoleDAO) GetByName(name string, CONSISTENCY_LEVEL int) (*config.Role, error) {
	dao.logger.Debugf("Getting role %s", name)
	var role config.Role
	if err := dao.db.
		Table("roles").
		First(&role, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (dao *GormRoleDAO) GetAll(CONSISTENCY_LEVEL int) ([]*config.Role, error) {
	dao.logger.Debug("Getting all roles")
	var roles []*config.Role
	if err := dao.db.Order("name asc").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
