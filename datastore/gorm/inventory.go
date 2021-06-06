package gorm

import (
	"errors"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type InventoryDAO interface {
	Create(Inventory entity.InventoryEntity) error
	Save(Inventory entity.InventoryEntity) error
	Update(Inventory entity.InventoryEntity) error
	Get(name string) (entity.InventoryEntity, error)
}

type GormInventoryDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	InventoryDAO
}

func NewInventoryDAO(logger *logging.Logger, db *gorm.DB) InventoryDAO {
	return &GormInventoryDAO{logger: logger, db: db}
}

func (dao *GormInventoryDAO) Create(entity entity.InventoryEntity) error {
	return dao.db.Create(entity).Error
}

func (dao *GormInventoryDAO) Save(entity entity.InventoryEntity) error {
	return dao.db.Save(entity).Error
}

func (dao *GormInventoryDAO) Update(entity entity.InventoryEntity) error {
	return dao.db.Update(entity).Error
}

func (dao *GormInventoryDAO) Get(channel string) (entity.InventoryEntity, error) {
	var Inventorys []entity.Inventory
	if err := dao.db.Where("channel = ?", channel).Find(&Inventorys).Error; err != nil {
		return nil, err
	}
	if len(Inventorys) == 0 {
		return nil, errors.New(fmt.Sprintf("Inventory item '%s' not found in database", channel))
	}
	return &Inventorys[0], nil
}
