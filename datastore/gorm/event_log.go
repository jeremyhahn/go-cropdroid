package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
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

func (dao *GormEventLogDAO) Save(log entity.EventLogEntity) error {
	return dao.db.Save(log).Error
}

func (dao *GormEventLogDAO) GetAll(CONSISTENCY_LEVEL int) ([]entity.EventLog, error) {
	var logs []entity.EventLog
	if err := dao.db.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (dao *GormEventLogDAO) Count(CONSISTENCY_LEVEL int) (int64, error) {
	var count int64
	if err := dao.db.Model(&entity.EventLog{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (dao *GormEventLogDAO) GetPage(CONSISTENCY_LEVEL int, offset, size int64) ([]entity.EventLog, error) {
	var logs []entity.EventLog
	if err := dao.db.Limit(size).
		Offset(offset).
		Where("farm_id = ?", dao.farmID).
		Order("timestamp desc").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (dao *GormEventLogDAO) GetAllDesc(CONSISTENCY_LEVEL int) ([]entity.EventLog, error) {
	var logs []entity.EventLog
	if err := dao.db.Order("timestamp desc").Find(&logs).Limit(100).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

/*
func (dao *GormEventLogDAO) Save(log entity.EventLogEntity) error {
	return dao.db.Save(log).Error
}

func (dao *GormEventLogDAO) Update(log entity.EventLogEntity) error {
	return dao.db.Update(log).Error
}

func (dao *GormEventLogDAO) Get(event string) (entity.EventLogEntity, error) {
	var logs []entity.EventLog
	if err := dao.db.Where("event = ?", event).Find(&logs).Error; err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, errors.New(fmt.Sprintf("Event '%s' not found in database", event))
	}
	return &logs[0], nil
}
*/
