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

func (permissionDAO *GormPermissionDAO) Save(permission config.PermissionConfig) error {
	perm := *permission.(*config.Permission)
	permissionDAO.logger.Debugf(fmt.Sprintf("Saving permission record: %+v", permission))
	return permissionDAO.db.Save(perm).Error
}

func (permissionDAO *GormPermissionDAO) Get(id uint64) (config.PermissionConfig, error) {
	permissionDAO.logger.Debugf("Getting permission id: %d", id)
	var permission config.Permission
	if err := permissionDAO.db.First(&permission, id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (permissionDAO *GormPermissionDAO) Delete(permission config.PermissionConfig) error {
	permissionDAO.logger.Debugf(fmt.Sprintf("Deleting permission record: %+v", permission))
	return permissionDAO.db.Delete(permission).Error
}
