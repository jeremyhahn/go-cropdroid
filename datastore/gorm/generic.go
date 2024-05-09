package gorm

import (
	logging "github.com/op/go-logging"
	"gorm.io/gorm"

	"github.com/jeremyhahn/go-cropdroid/config/dao"
)

type GenericGormDAO[E any] struct {
	db     *gorm.DB
	logger *logging.Logger
	dao.GenericDAO[E]
}

func NewGenericGormDAO[E any](logger *logging.Logger, db *gorm.DB) dao.GenericDAO[E] {
	logger.Infof("Creatnig new %T GORM DAO", *new(E))
	return &GenericGormDAO[E]{logger: logger, db: db}
}

func (dao *GenericGormDAO[E]) Save(entity *E) error {
	dao.logger.Infof("Save GORM entity: %+v", entity)
	return dao.db.Save(&entity).Error
}

func (dao *GenericGormDAO[E]) Get(id uint64, CONSISTENCY_LEVEL int) (E, error) {
	dao.logger.Infof("Get GORM entity with id: %d", id)
	var entity = new(E)
	if err := dao.db.
		First(entity, id).Error; err != nil {
		dao.logger.Warningf("GenericGormDAO.Get error: %s", err.Error())
		return *entity, err
	}
	return *entity, nil
}

func (dao *GenericGormDAO[E]) GetPage(page, pageSize, CONSISTENCY_LEVEL int) ([]E, error) {
	if page < 1 {
		page = 1
	}
	//var offset = (page-1)*pageSize + 1
	var offset = (page - 1) * pageSize
	var entities []E
	if err := dao.db.Limit(pageSize).
		Offset(offset).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (dao *GenericGormDAO[E]) Update(entity *E) error {
	dao.logger.Infof("Update GORM entity: %+v", entity)
	return dao.db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(entity).Error
}

func (dao *GenericGormDAO[E]) Delete(entity *E) error {
	dao.logger.Infof("Delete GORM entity: %+v", entity)
	return dao.db.Delete(entity).Error
}

// This method is only here to provide compatiiblty with the interface while refactoring
// to prevent the rest of the project from breaking if its removed
func (dao *GenericGormDAO[E]) GetAll(CONSISTENCY_LEVEL int) ([]*E, error) {
	return make([]*E, 0), nil
}

// func (dao *GenericGormDAO[E]) GetAll(CONSISTENCY_LEVEL int) ([]E, error) {
// 	dao.logger.Infof("Getting all entities")
// 	var entities []E
// 	if err := dao.db.Find(&entities).Error; err != nil {
// 		return nil, err
// 	}
// 	return entities, nil
// }
