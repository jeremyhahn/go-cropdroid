package gorm

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/entity"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormEventLogDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	farmID int
	dao.EventLogDAO
}

func NewEventLogDAO(logger *logging.Logger, db *gorm.DB, farmID int) dao.EventLogDAO {
	return &GormEventLogDAO{logger: logger, db: db, farmID: farmID}
}

func (dao *GormEventLogDAO) Save(log *entity.EventLog) error {
	return dao.db.Save(log).Error
}

func (eventLogDAO *GormEventLogDAO) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*entity.EventLog], error) {
	pageResult := dao.PageResult[*entity.EventLog]{
		Page:     pageQuery.Page,
		PageSize: pageQuery.PageSize}
	var sortOrder string
	if pageQuery.SortOrder == query.SORT_ASCENDING {
		sortOrder = "asc"
	} else {
		sortOrder = "desc"
	}
	var logs []*entity.EventLog
	if err := eventLogDAO.db.Limit(pageQuery.PageSize).
		Offset(pageQuery.Page).
		Where("farm_id = ?", eventLogDAO.farmID).
		Order(fmt.Sprintf("timestamp %s", sortOrder)).
		Limit(pageQuery.PageSize + 1). // peek one record to set HasMore flag
		Find(&logs).Error; err != nil {
		return pageResult, err
	}
	// If the peek record was returned, set the HasMore flag and remove the +1 record
	if len(logs) == pageQuery.PageSize+1 {
		pageResult.HasMore = true
		logs = logs[:len(logs)-1]
	}
	pageResult.Entities = logs
	return pageResult, nil
}

func (eventLogDAO *GormEventLogDAO) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*entity.EventLog], CONSISTENCY_LEVEL int) error {

	pageResult, err := eventLogDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
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
		return eventLogDAO.ForEachPage(nextPageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
	}
	return nil
}

func (eventLogDAO *GormEventLogDAO) Count(CONSISTENCY_LEVEL int) (int64, error) {
	var count int64
	if err := eventLogDAO.db.Model(&entity.EventLog{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
