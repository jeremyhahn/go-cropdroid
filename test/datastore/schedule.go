package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/assert"
)

func TestScheduleCRUD(t *testing.T, scheduleDAO dao.ScheduleDAO,
	org *config.Organization) {

	farm1 := org.GetFarms()[0]
	device1 := farm1.GetDevices()[1]
	channel1 := device1.GetChannels()[0]
	schedules := channel1.GetSchedule()
	schedule1 := schedules[0]
	schedule2 := schedules[1]

	err := scheduleDAO.Save(farm1.GetID(), device1.GetID(), schedule1)
	assert.Nil(t, err)

	err = scheduleDAO.Save(farm1.GetID(), device1.GetID(), schedule2)
	assert.Nil(t, err)

	persistedSchedules, err := scheduleDAO.GetByChannelID(farm1.GetID(),
		device1.GetID(), channel1.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(persistedSchedules))

	found := false
	for _, persistedSchedule := range persistedSchedules {
		if schedule1.GetID() == persistedSchedule.GetID() {
			assert.Equal(t, schedule1.GetID(), persistedSchedule.GetID())
			assert.Equal(t, schedule1.GetChannelID(), persistedSchedule.GetChannelID())
			assert.Equal(t, schedule1.GetFrequency(), persistedSchedule.GetFrequency())
			assert.Equal(t, schedule1.GetInterval(), persistedSchedule.GetInterval())
			assert.Equal(t, schedule1.GetCount(), persistedSchedule.GetCount())
			assert.Equal(t, schedule1.GetDays(), persistedSchedule.GetDays())
			assert.Equal(t, schedule1.GetExecutionCount(), persistedSchedule.GetExecutionCount())
			found = true
		}
	}
	assert.True(t, found)

	err = scheduleDAO.Delete(farm1.GetID(), device1.GetID(),
		persistedSchedules[0])

	persistedSchedules, err = scheduleDAO.GetByChannelID(farm1.GetID(),
		device1.GetID(), channel1.GetID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedSchedules))
}
