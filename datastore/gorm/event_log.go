package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type EventLogDAO interface {
	Create(EventLog entity.EventLogEntity) error
	GetAll() ([]entity.EventLog, error)
	GetLogs() ([]entity.EventLog, error)
	GetPage(page, size int64) ([]entity.EventLog, error)
	Count() (int64, error)
}

type GormEventLogDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	EventLogDAO
}

func NewEventLogDAO(logger *logging.Logger, db *gorm.DB) EventLogDAO {
	return &GormEventLogDAO{logger: logger, db: db}
}

func (dao *GormEventLogDAO) Create(log entity.EventLogEntity) error {
	return dao.db.Create(log).Error
}

func (dao *GormEventLogDAO) GetAll() ([]entity.EventLog, error) {
	var logs []entity.EventLog
	if err := dao.db.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (dao *GormEventLogDAO) Count() (int64, error) {
	var count int64
	if err := dao.db.Model(&entity.EventLog{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (dao *GormEventLogDAO) GetPage(offset, size int64) ([]entity.EventLog, error) {
	var logs []entity.EventLog
	if err := dao.db.Limit(size).Offset(offset).Order("timestamp desc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (dao *GormEventLogDAO) GetLogs() ([]entity.EventLog, error) {
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
