package gorm

import (
	"strings"
	"testing"
	"time"

	"github.com/jeremyhahn/cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestScheduleCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Schedule{})

	scheduleDAO := NewScheduleDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, scheduleDAO)

	startDate := time.Now().In(currentTest.location)
	endDate := time.Now().In(currentTest.location).Add(time.Hour * 10)
	schedule1 := &config.Schedule{
		ID:             1,
		ChannelID:      1,
		StartDate:      startDate,
		EndDate:        &endDate,
		Frequency:      1,  // Daily
		Interval:       60, // Seconds
		Count:          0,
		Days:           nil,
		LastExecuted:   startDate,
		ExecutionCount: 0}

	days := "MO,TU"
	schedule2 := &config.Schedule{
		ID:             2,
		ChannelID:      3,
		StartDate:      time.Now().In(currentTest.location),
		EndDate:        nil,
		Frequency:      2,  // Daily
		Interval:       15, // Seconds
		Count:          0,
		Days:           &days,
		LastExecuted:   time.Now().In(currentTest.location),
		ExecutionCount: 0}

	err := scheduleDAO.Save(schedule1)
	assert.Nil(t, err)

	err = scheduleDAO.Save(schedule2)
	assert.Nil(t, err)

	persistedSchedules, err := scheduleDAO.GetByChannelID(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedSchedules))

	persistedSchedule1 := persistedSchedules[0]
	//persistedEndDate := persistedSchedule1.GetEndDate().In(currentTest.location)

	assert.Equal(t, schedule1.ID, persistedSchedule1.GetID())
	assert.Equal(t, schedule1.ChannelID, persistedSchedule1.GetChannelID())
	//assert.Equal(t, schedule1.StartDate, persistedSchedule1.GetStartDate().In(currentTest.location))  cockroach truncates to microseconds with 6 digit precision
	//assert.Equal(t, schedule1.EndDate, &persistedEndDate)  same as above
	assert.Equal(t, schedule1.Frequency, persistedSchedule1.GetFrequency())
	assert.Equal(t, schedule1.Interval, persistedSchedule1.GetInterval())
	assert.Equal(t, schedule1.Count, persistedSchedule1.GetCount())
	assert.Equal(t, schedule1.Days, persistedSchedule1.GetDays())
	//assert.Equal(t, schedule1.LastExecuted, persistedSchedule1.GetLastExecuted().In(currentTest.location))
	assert.Equal(t, schedule1.ExecutionCount, persistedSchedule1.GetExecutionCount())

	persistedSchedules, err = scheduleDAO.GetByChannelID(3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedSchedules))

	persistedSchedule2 := persistedSchedules[0]

	_days := strings.Split(*schedule2.Days, ",")
	assert.Equal(t, schedule2.ID, persistedSchedule2.GetID())
	assert.Equal(t, schedule2.ChannelID, persistedSchedule2.GetChannelID())
	//assert.Equal(t, schedule2.StartDate, persistedSchedule2.GetStartDate().In(currentTest.location)) same as above
	assert.Equal(t, schedule2.EndDate, persistedSchedule2.GetEndDate())
	assert.Equal(t, schedule2.Frequency, persistedSchedule2.GetFrequency())
	assert.Equal(t, schedule2.Interval, persistedSchedule2.GetInterval())
	assert.Equal(t, schedule2.Count, persistedSchedule2.GetCount())
	assert.Equal(t, schedule2.Days, persistedSchedule2.GetDays())
	assert.Equal(t, 2, len(_days))
	assert.Equal(t, _days[0], "MO")
	assert.Equal(t, _days[1], "TU")
	//assert.Equal(t, schedule2.LastExecuted, persistedSchedule2.GetLastExecuted())
	assert.Equal(t, schedule2.ExecutionCount, persistedSchedule2.GetExecutionCount())

	currentTest.Cleanup()
}

func TestScheduleCRUDUsingEntity(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Schedule{})

	scheduleDAO := NewScheduleDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, scheduleDAO)

	days := "MO,WE,SA"
	startDate := time.Now().In(currentTest.location)
	endDate := time.Now().In(currentTest.location).Add(time.Hour * 10)
	schedule1 := &config.Schedule{
		ID:             1,
		ChannelID:      1,
		StartDate:      startDate,
		EndDate:        &endDate,
		Frequency:      1,  // Daily
		Interval:       60, // Seconds
		Count:          0,
		Days:           &days,
		LastExecuted:   startDate,
		ExecutionCount: 0}

	schedule2 := &config.Schedule{
		ID:             2,
		ChannelID:      3,
		StartDate:      time.Now().In(currentTest.location),
		EndDate:        nil,
		Frequency:      2,  // Daily
		Interval:       15, // Seconds
		Count:          0,
		Days:           &days,
		LastExecuted:   time.Now().In(currentTest.location),
		ExecutionCount: 0}

	err := scheduleDAO.Save(schedule1)
	assert.Nil(t, err)

	err = scheduleDAO.Save(schedule2)
	assert.Nil(t, err)

	persistedSchedules, err := scheduleDAO.GetByChannelID(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedSchedules))

	persistedSchedule1 := persistedSchedules[0]
	//persistedEndDate := persistedSchedule1.GetEndDate().In(currentTest.location)

	assert.Equal(t, schedule1.ID, persistedSchedule1.GetID())
	assert.Equal(t, schedule1.ChannelID, persistedSchedule1.GetChannelID())
	//assert.Equal(t, schedule1.StartDate, persistedSchedule1.GetStartDate().In(currentTest.location))
	//assert.Equal(t, schedule1.EndDate, &persistedEndDate)
	assert.Equal(t, schedule1.Frequency, persistedSchedule1.GetFrequency())
	assert.Equal(t, schedule1.Interval, persistedSchedule1.GetInterval())
	assert.Equal(t, schedule1.Count, persistedSchedule1.GetCount())
	assert.Equal(t, schedule1.Days, persistedSchedule1.GetDays())
	//assert.Equal(t, schedule1.LastExecuted, persistedSchedule1.GetLastExecuted().In(currentTest.location))
	assert.Equal(t, schedule1.ExecutionCount, persistedSchedule1.GetExecutionCount())

	persistedSchedules, err = scheduleDAO.GetByChannelID(3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedSchedules))

	persistedSchedule2 := persistedSchedules[0]

	assert.Equal(t, schedule2.ID, persistedSchedule2.GetID())
	//assert.Equal(t, schedule2.ChannelID, persistedSchedule2.GetChannelID())
	//assert.Equal(t, schedule2.StartDate, persistedSchedule2.GetStartDate().In(currentTest.location))
	assert.Equal(t, schedule2.EndDate, persistedSchedule2.GetEndDate())
	assert.Equal(t, schedule2.Frequency, persistedSchedule2.GetFrequency())
	assert.Equal(t, schedule2.Interval, persistedSchedule2.GetInterval())
	assert.Equal(t, schedule2.Count, persistedSchedule2.GetCount())
	assert.Equal(t, schedule2.Days, persistedSchedule2.GetDays())
	//assert.Equal(t, schedule2.LastExecuted, persistedSchedule2.GetLastExecuted())
	assert.Equal(t, schedule2.ExecutionCount, persistedSchedule2.GetExecutionCount())

	currentTest.Cleanup()
}
