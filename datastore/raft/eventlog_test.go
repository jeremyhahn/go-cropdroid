//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/entity"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/stretchr/testify/assert"
)

func TestEventLogCRUD(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	idGenerator := IntegrationTestCluster.app.IdGenerator

	farmID := FarmConfigClusterID
	eventLogDAO := NewRaftEventLogDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmID)

	assert.NotNil(t, eventLogDAO)
	eventLogDAO.StartLocalCluster(IntegrationTestCluster, true)

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

	err := eventLogDAO.Save(eventLogRecord1)
	assert.Nil(t, err)

	err = eventLogDAO.Save(eventLogRecord2)
	assert.Nil(t, err)

	page1, err := eventLogDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, page1.Entities)
	assert.Equal(t, 2, len(page1.Entities))
	assert.Equal(t, farmID, page1.Entities[0].FarmID)
	assert.Equal(t, deviceID1, page1.Entities[0].DeviceID)
	assert.Equal(t, deviceName1, page1.Entities[0].DeviceName)
	assert.Equal(t, "test", page1.Entities[0].EventType)
	assert.Equal(t, "test2", page1.Entities[1].EventType)
	//assert.Equal(t, timestamp1, page1.Entities[0].GetTimestampAsObject())

	descendingPageQuery := query.PageQuery{Page: 1, PageSize: 25, SortOrder: query.SORT_DESCENDING}
	page1, err = eventLogDAO.GetPage(descendingPageQuery, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, page1)
	assert.Equal(t, 2, len(page1.Entities))
	assert.Equal(t, "test2", page1.Entities[0].EventType)
	assert.Equal(t, "test", page1.Entities[1].EventType)

	count, err := eventLogDAO.Count(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), int64(count))

	page1, err = eventLogDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), int64(len(page1.Entities)))
}
