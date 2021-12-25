package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormUserDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.UserDAO
}

func NewUserDAO(logger *logging.Logger, db *gorm.DB) dao.UserDAO {
	return &GormUserDAO{logger: logger, db: db}
}

func CreateUserDAO(db *gorm.DB, user common.UserAccount) dao.UserDAO {
	return &GormUserDAO{
		logger: &logging.Logger{},
		db:     db}
}

// func (dao *GormUserDAO) GetAll(orgID uint64) ([]config.UserConfig, error) {
// 	var users []config.User
// 	if err := dao.db.
// 		Preload("Roles").
// 		Joins("JOIN permissions on permissions.organization_id = farms.id").
// 		Where("permissions.organization_id = ?", orgID).
// 		Find(&users).Error; err != nil {
// 		dao.logger.Errorf("[UserDAO.GetAll] Error: %s", err.Error())
// 		return nil, err
// 	}
// 	userConfigs := make([]config.UserConfig, len(users))
// 	for i, user := range users {
// 		userConfigs[i] = &user
// 	}
// 	return userConfigs, nil
// }

func (dao *GormUserDAO) GetByEmail(email string) (config.UserConfig, error) {
	var user config.User
	if err := dao.db.
		Preload("Roles").
		First(&user, "email = ?", email).Error; err != nil {
		dao.logger.Errorf("[UserDAO.GetByEmail] %s", err.Error())
		return nil, err
	}
	return &user, nil
}

// Saves a new user to the database.
func (dao *GormUserDAO) Create(user config.UserConfig) error {
	if err := dao.db.Create(user).Error; err != nil {
		dao.logger.Errorf("[UserDAO.Create] Error:%s", err.Error())
		return err
	}
	return nil
}

// Saves or updates a user account.
func (dao *GormUserDAO) Save(user config.UserConfig) error {
	if err := dao.db.Save(user).Error; err != nil {
		dao.logger.Errorf("[UserDAO.Save] Error:%s", err.Error())
		return err
	}
	return nil
}

// Deletes a user from the database
func (dao *GormUserDAO) Delete(user config.UserConfig) error {
	dao.logger.Errorf("[UserDAO.Delete] user: %+v", user)
	return dao.db.Delete(user).Error
}
