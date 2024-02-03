//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"
)

func TestEventLogCRUD(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)

	farmName := "farm1"
	farmID := idGenerator.NewID(farmName)

	Cluster.CreateEventLogCluster(farmID)

	EventLogDAO := NewRaftEventLogDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmID)

	assert.NotNil(t, EventLogDAO)

	deviceName1 := "device1"
	deviceID1 := idGenerator.NewDeviceID(farmID, deviceName1)
	timestamp1 := time.Now()
	eventLogRecord1 := &entity.EventLog{
		FarmID:     farmID,
		DeviceID:   deviceID1,
		DeviceName: deviceName1,
		EventType:  "test",
		Message:    "this is a test log entry",
		Timestamp:  timestamp1,
	}

	deviceName2 := "device2"
	deviceID2 := idGenerator.NewDeviceID(farmID, deviceName2)
	eventLogRecord2 := &entity.EventLog{
		FarmID:     farmID,
		DeviceID:   deviceID2,
		DeviceName: deviceName2,
		EventType:  "test2",
		Message:    "this is a second test log entry",
		Timestamp:  time.Now(),
	}

	err := EventLogDAO.Save(eventLogRecord1)
	assert.Nil(t, err)

	err = EventLogDAO.Save(eventLogRecord2)
	assert.Nil(t, err)

	eventLogRecords, err := EventLogDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, eventLogRecords)
	assert.Equal(t, 2, len(eventLogRecords))
	assert.Equal(t, farmID, eventLogRecords[0].FarmID)
	assert.Equal(t, deviceID1, eventLogRecords[0].DeviceID)
	assert.Equal(t, deviceName1, eventLogRecords[0].DeviceName)
	assert.Equal(t, "test", eventLogRecords[0].EventType)
	assert.Equal(t, "this is a test log entry", eventLogRecords[0].Message)
	//assert.Equal(t, timestamp1, eventLogRecords[0].GetTimestampAsObject())

	descEventLogRecords, err := EventLogDAO.GetAllDesc(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, descEventLogRecords)
	assert.Equal(t, 2, len(descEventLogRecords))

	count, err := EventLogDAO.Count(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), int64(count))

	pageRecords, err := EventLogDAO.GetPage(common.CONSISTENCY_LOCAL, 1, 10)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), int64(len(pageRecords)))
}
