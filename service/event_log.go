package service

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type EventLog struct {
	app    *app.App
	dao    dao.EventLogDAO
	farmID uint64
	EventLogService
}

func NewEventLogService(app *app.App, dao dao.EventLogDAO,
	farmID uint64) EventLogService {
	return &EventLog{
		app:    app,
		dao:    dao,
		farmID: farmID}
}

func (eventLog *EventLog) GetFarmID() uint64 {
	return eventLog.farmID
}

func (eventLog *EventLog) Create(deviceID uint64, deviceName, eventType, message string) {

	eventLogEntry := &entity.EventLog{
		FarmID:     eventLog.farmID,
		DeviceID:   deviceID,
		DeviceName: deviceName,
		EventType:  eventType,
		Message:    message,
		Timestamp:  time.Now()}

	eventLog.app.Logger.Debugf("Event log entry: %+v", eventLogEntry)

	err := eventLog.dao.Save(eventLogEntry)
	if err != nil {
		eventLog.app.Logger.Errorf("[Create] Error: %s", err)
	}
}

func (eventLog *EventLog) GetPage(page int64) *viewmodel.EventsPage {
	var pageSize int64 = 10
	if page < 1 {
		page = 1
	}
	var offset = (page-1)*pageSize + 1
	eventLog.app.Logger.Debugf("[GetPage]")
	entities, err := eventLog.dao.GetPage(common.CONSISTENCY_LOCAL, offset, pageSize)
	if err != nil {
		eventLog.app.Logger.Errorf("[GetPage] Error: %s", err)
	}
	count, err := eventLog.dao.Count(common.CONSISTENCY_LOCAL)
	if err != nil {
		eventLog.app.Logger.Errorf("[GetPage] Error: %s", err)
	}
	return &viewmodel.EventsPage{
		Events: entities,
		Page:   page,
		Size:   pageSize,
		Count:  count,
		Start:  offset,
		End:    offset + pageSize}
}

func (eventLog *EventLog) GetAll() []entity.EventLog {
	eventLog.app.Logger.Debugf("[GetAll]")
	entities, err := eventLog.dao.GetAllDesc(common.CONSISTENCY_LOCAL)
	if err != nil {
		eventLog.app.Logger.Errorf("[GetAll] Error: %s", err)
	}
	return entities
}
