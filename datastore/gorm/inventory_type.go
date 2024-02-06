package gorm

import (
	"errors"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type InventoryTypeDAO interface {
	Create(InventoryType entity.InventoryTypeEntity) error
	Save(InventoryType entity.InventoryTypeEntity) error
	Update(InventoryType entity.InventoryTypeEntity) error
	Get(name string) (entity.InventoryTypeEntity, error)
}

type SqliteInventoryTypeDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	InventoryTypeDAO
}

func NewInventoryTypeDAO(logger *logging.Logger, db *gorm.DB) InventoryTypeDAO {
	return &SqliteInventoryTypeDAO{logger: logger, db: db}
}

func (dao *SqliteInventoryTypeDAO) Create(entity entity.InventoryTypeEntity) error {
	return dao.db.Create(entity).Error
}

func (dao *SqliteInventoryTypeDAO) Save(entity entity.InventoryTypeEntity) error {
	return dao.db.Save(entity).Error
}

func (dao *SqliteInventoryTypeDAO) Update(entity entity.InventoryTypeEntity) error {
	return dao.db.Save(entity).Error
}

func (dao *SqliteInventoryTypeDAO) Get(channel string) (entity.InventoryTypeEntity, error) {
	var InventoryTypes []entity.InventoryType
	if err := dao.db.Where("channel = ?", channel).Find(&InventoryTypes).Error; err != nil {
		return nil, err
	}
	if len(InventoryTypes) == 0 {
		return nil, errors.New(fmt.Sprintf("InventoryType item '%s' not found in database", channel))
	}
	return &InventoryTypes[0], nil
}
