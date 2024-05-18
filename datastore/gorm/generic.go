package gorm

import (
	logging "github.com/op/go-logging"
	"gorm.io/gorm"

	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
)

type GenericGormDAO[E any] struct {
	db     *gorm.DB
	logger *logging.Logger
	dao.GenericDAO[E]
}

func NewGenericGormDAO[E any](logger *logging.Logger, db *gorm.DB) dao.GenericDAO[E] {
	logger.Infof("Creating new %T GORM DAO", *new(E))
	return &GenericGormDAO[E]{logger: logger, db: db}
}

func (genericDAO *GenericGormDAO[E]) Save(entity E) error {
	genericDAO.logger.Infof("Save GORM entity: %+v", entity)
	return genericDAO.db.Save(&entity).Error
}

func (genericDAO *GenericGormDAO[E]) Get(id uint64, CONSISTENCY_LEVEL int) (E, error) {
	genericDAO.logger.Infof("Get GORM entity with id: %d", id)
	var entity = new(E)
	if err := genericDAO.db.
		First(entity, id).Error; err != nil {
		genericDAO.logger.Warningf("GenericGormDAO.Get error: %s", err.Error())
		return *entity, err
	}
	return *entity, nil
}

func (genericDAO *GenericGormDAO[E]) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[E], error) {
	page := pageQuery.Page
	pageSize := pageQuery.PageSize
	pageResult := dao.PageResult[E]{
		Page:     page,
		PageSize: pageSize}
	if page < 1 {
		page = 1
	}
	var offset = (page - 1) * pageSize
	var entities []E
	if err := genericDAO.db.
		Limit(pageSize + 1). // peek one record to set HasMore flag
		Offset(offset).
		Find(&entities).Error; err != nil {
		return pageResult, err
	}
	// If the peek record was returned, set the HasMore flag and remove the +1 record
	if len(entities) == pageSize+1 {
		pageResult.HasMore = true
		entities = entities[:len(entities)-1]
	}
	pageResult.Entities = entities
	return pageResult, nil
}

func (genericDAO *GenericGormDAO[E]) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[E], CONSISTENCY_LEVEL int) error {

	pageResult, err := genericDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
	if err != nil {
		return nil
	}
	if err = pagerProcFunc(pageResult.Entities); err != nil {
		return err
	}
	if pageResult.HasMore {
		nextPageQuery := query.PageQuery{
			Page:      pageQuery.Page + 1,
			PageSize:  pageQuery.PageSize,
			SortOrder: pageQuery.SortOrder}
		return genericDAO.ForEachPage(nextPageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
	}
	return nil
}

func (genericDAO *GenericGormDAO[E]) Update(entity E) error {
	genericDAO.logger.Infof("Update GORM entity: %+v", entity)
	return genericDAO.db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(entity).Error
}

func (genericDAO *GenericGormDAO[E]) Delete(entity E) error {
	genericDAO.logger.Infof("Delete GORM entity: %+v", entity)
	return genericDAO.db.Delete(entity).Error
}

func (genericDAO *GenericGormDAO[E]) Count(CONSISTENCY_LEVEL int) (int64, error) {
	var count int64
	var entity E
	if err := genericDAO.db.Model(&entity).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
