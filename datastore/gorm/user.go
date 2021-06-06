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
	//scope.SetUser(user)
	return &GormUserDAO{db: db}
}

/*
func (dao *GormUserDAO) GetById(userId int) (entity.UserEntity, error) {
	var user entity.User
	user.ID = userId
	if err := dao.db.First(&user).Error; err != nil {
		dao.logger.Errorf("[UserDAO.GetById] Error: %s", err.Error())
		return nil, err
	}
	return &user, nil
}*/

func (dao *GormUserDAO) GetByEmail(email string) (config.UserConfig, error) {
	var user config.User
	if err := dao.db.Preload("Roles").First(&user, "email = ?", email).Error; err != nil {
		dao.logger.Errorf("[UserDAO.GetByEmail] %s", err.Error())
		return nil, err
	}
	return &user, nil
}

func (dao *GormUserDAO) Create(user config.UserConfig) error {
	if err := dao.db.Create(user).Error; err != nil {
		dao.logger.Errorf("[UserDAO.Create] Error:%s", err.Error())
		return err
	}
	return nil
}

func (dao *GormUserDAO) Save(user config.UserConfig) error {
	if err := dao.db.Save(user).Error; err != nil {
		dao.logger.Errorf("[UserDAO.Save] Error:%s", err.Error())
		return err
	}
	return nil
}

/*
func (dao *GormUserDAO) Find() ([]entity.User, error) {
	var users []entity.User
	if err := dao.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
*/
