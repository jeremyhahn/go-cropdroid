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
	//consistencyLevel := common.CONSISTENCY_LOCAL
	//testFarmStateName := "root@localhost"

	farmName := "device1"
	farmID := idGenerator.NewID(farmName)

	Cluster.CreateEventLogCluster(farmID)

	EventLogDAO := NewRaftEventLogDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmID)

	assert.NotNil(t, EventLogDAO)

	eventLogRecord1 := &entity.EventLog{
		FarmID:    farmID,
		Device:    "testDevice",
		Type:      "test",
		Message:   "This is a test log entry",
		Timestamp: time.Now(),
	}

	eventLogRecord2 := &entity.EventLog{
		FarmID:    farmID,
		Device:    "testDevice2",
		Type:      "test2",
		Message:   "This is a second test log entry",
		Timestamp: time.Now(),
	}

	err := EventLogDAO.Save(eventLogRecord1)
	assert.Nil(t, err)

	err = EventLogDAO.Save(eventLogRecord2)
	assert.Nil(t, err)

	eventLogRecords, err := EventLogDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, eventLogRecords)
	assert.Equal(t, 2, len(eventLogRecords))

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
