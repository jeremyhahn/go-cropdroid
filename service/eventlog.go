package service

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/entity"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
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

func (eventLog *EventLog) GetPage(pageQuery query.PageQuery) (dao.PageResult[*entity.EventLog], error) {
	eventLog.app.Logger.Debugf("[GetPage]: %+v", pageQuery)
	page, err := eventLog.dao.GetPage(pageQuery, common.CONSISTENCY_LOCAL)
	if err != nil {
		eventLog.app.Logger.Errorf("[GetPage] Error: %s", err)
	}
	return page, nil
}
