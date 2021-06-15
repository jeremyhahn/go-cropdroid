// +build ignore

package test

import (
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/assert"
)

func TestGetSchedule(t *testing.T) {

	scope, dao, service := newTestScheduleService()

	endDate := time.Now().In(scope.GetLocation())

	schedules := []config.Schedule{
		config.Schedule{
			ID:        1,
			ChannelID: 1,
			StartDate: time.Now().In(scope.GetLocation()),
			EndDate:   &endDate,
			Frequency: 0,
			Interval:  0,
			Count:     0,
			Days:      []string(nil)}}

	dao.On("GetByChannelID", 1).Return(schedules)

	service.GetSchedule(1)

	dao.AssertNumberOfCalls(t, "GetByChannelID", 1)
	dao.AssertExpectations(t)
}

func TestIsScheduled_AtStartTime(t *testing.T) {

	_, scope := NewUnitTestContext()
	now := time.Now().In(scope.GetLocation())
	_, _, service := createTestScheduleService(now)

	endTime := now.Add(time.Duration(1) * time.Hour)

	schedule := &model.Schedule{
		StartDate: now,
		EndDate:   &endTime}

	result := service.IsScheduled(schedule, 0)
	assert.Equal(t, true, result)
}

func TestIsScheduled_MinutePriorToEndTime(t *testing.T) {

	_, scope := NewUnitTestContext()
	now := time.Now().In(scope.GetLocation())
	_, _, service := createTestScheduleService(now)

	startTime := now.Add(time.Duration(-1) * time.Minute)
	endTime := now.Add(time.Duration(1) * time.Minute)

	schedule := &model.Schedule{
		StartDate: startTime,
		EndDate:   &endTime}

	result := service.IsScheduled(schedule, 0)
	assert.Equal(t, true, result)
}

func TestIsScheduled_BetweenStartAndEnd(t *testing.T) {

	_, scope := NewUnitTestContext()
	now := time.Now().In(scope.GetLocation())
	_, _, service := createTestScheduleService(now)

	startTime := now.Add(time.Duration(-1) * time.Hour)
	endTime := now.Add(time.Duration(1) * time.Hour)

	schedule := &model.Schedule{
		StartDate: startTime,
		EndDate:   &endTime}

	result := service.IsScheduled(schedule, 0)
	assert.Equal(t, true, result)
}

func TestIsNotScheduled_NowBeforeStartTime(t *testing.T) {

	_, scope := NewUnitTestContext()
	now := time.Now().In(scope.GetLocation())

	oneHourAgo := now.Add(time.Duration(-1) * time.Hour)

	_, _, service := createTestScheduleService(oneHourAgo)

	oneHourFromNow := now.Add(time.Duration(1) * time.Hour)

	tomorrow1030 := now.AddDate(0, 0, 1).
		Add(time.Duration(10) * time.Hour).
		Add(time.Duration(30) * time.Minute)

	schedule := &model.Schedule{
		StartDate: oneHourFromNow,
		EndDate:   &tomorrow1030}

	result := service.IsScheduled(schedule, 0)
	assert.Equal(t, false, result)
}

func TestIsNotScheduled_WhenNowIsEndTime(t *testing.T) {

	_, scope := NewUnitTestContext()
	now := time.Now().In(scope.GetLocation())

	oneHourFromNow := now.Add(time.Duration(1) * time.Hour)

	_, _, service := createTestScheduleService(oneHourFromNow)

	oneHourAgo := now.Add(time.Duration(-1) * time.Hour)

	endDate := time.Now()
	schedule := &model.Schedule{
		StartDate: oneHourAgo,
		EndDate:   &endDate}

	result := service.IsScheduled(schedule, 0)
	assert.Equal(t, false, result)
}

func TestIsNotScheduled_WhenNowAfterEndTime(t *testing.T) {

	_, scope := NewUnitTestContext()
	now := time.Now().In(scope.GetLocation())
	_, _, service := createTestScheduleService(now)

	startTime := now.Add(time.Duration(-2) * time.Hour)
	endTime := now.Add(time.Duration(-1) * time.Minute)

	schedule := &model.Schedule{
		StartDate: startTime,
		EndDate:   &endTime}

	result := service.IsScheduled(schedule, 0)
	assert.Equal(t, false, result)
}

func TestIsNotScheduled_MinuteAfterEndTime(t *testing.T) {

	_, scope := NewUnitTestContext()
	now := time.Now().In(scope.GetLocation())

	oneMinuteFromNow := now.Add(time.Duration(1) * time.Minute)

	_, _, service := createTestScheduleService(oneMinuteFromNow)

	startTime := now.Add(time.Duration(-1) * time.Minute)

	schedule := &model.Schedule{
		StartDate: startTime,
		EndDate:   &now}

	result := service.IsScheduled(schedule, 0)
	assert.Equal(t, false, result)
}

func TestIsNotScheduled_AtEndTime(t *testing.T) {

	_, scope := NewUnitTestContext()
	now := time.Now().In(scope.GetLocation())

	_, _, service := createTestScheduleService(now)

	startTime := now.Add(time.Duration(-1) * time.Minute)

	schedule := &model.Schedule{
		StartDate: startTime,
		EndDate:   &now}

	result := service.IsScheduled(schedule, 0)
	assert.Equal(t, false, result)
}

func TestIsScheduled_YesterdayDailyAtMidnight(t *testing.T) {

	_, scope := NewUnitTestContext()

	startHour := 19
	startMin := 0

	now := time.Now()
	pointInTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, scope.GetLocation())

	_, _, service := createTestScheduleService(pointInTime)

	startDate := time.Date(now.Year(), now.Month(), now.Day()-1, startHour, startMin, 0, 0, scope.GetLocation())

	schedule := &model.Schedule{
		StartDate: startDate,
		Frequency: common.SCHEDULE_FREQUENCY_DAILY}

	result := service.IsScheduled(schedule, 43200)
	assert.Equal(t, true, result)
}

func TestIsScheduled_YesterdayDailyOneMinuteBeforeEndTime(t *testing.T) {

	_, scope := NewUnitTestContext()

	startHr := 19
	startMin := 0

	nowHr := 6
	nowMin := 59

	now := time.Now()
	pointInTime := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin, 0, 0, scope.GetLocation())

	_, _, service := createTestScheduleService(pointInTime)

	startDate := time.Date(now.Year(), now.Month(), now.Day()-1, startHr, startMin, 0, 0, scope.GetLocation())

	schedule := &model.Schedule{
		StartDate: startDate,
		Frequency: common.SCHEDULE_FREQUENCY_DAILY}

	result := service.IsScheduled(schedule, 43200)
	assert.Equal(t, true, result)
}

func TestIsScheduled_YesterdayDailyAtEndTime(t *testing.T) {

	_, scope := NewUnitTestContext()

	startHr := 19
	startMin := 0

	nowHr := 7
	nowMin := 00

	now := time.Now()
	pointInTime := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin, 0, 0, scope.GetLocation())

	_, _, service := createTestScheduleService(pointInTime)

	startDate := time.Date(now.Year(), now.Month(), now.Day()-1, startHr, startMin, 0, 0, scope.GetLocation())

	schedule := &model.Schedule{
		StartDate: startDate,
		Frequency: common.SCHEDULE_FREQUENCY_DAILY}

	result := service.IsScheduled(schedule, 43200)
	assert.Equal(t, false, result)
}

func TestIsScheduled_YesterdayDailyOneMinuteAfterEndTime(t *testing.T) {

	_, scope := NewUnitTestContext()

	startHr := 19
	startMin := 0

	nowHr := 7
	nowMin := 01

	now := time.Now()
	pointInTime := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin, 0, 0, scope.GetLocation())

	_, _, service := createTestScheduleService(pointInTime)

	startDate := time.Date(now.Year(), now.Month(), now.Day()-1, startHr, startMin, 0, 0, scope.GetLocation())

	schedule := &model.Schedule{
		StartDate: startDate,
		Frequency: common.SCHEDULE_FREQUENCY_DAILY}

	result := service.IsScheduled(schedule, 43200)
	assert.Equal(t, false, result)
}

func TestIsScheduled_YesterdayOneMinuteBeforeStartTime(t *testing.T) {

	_, scope := NewUnitTestContext()

	startHr := 19
	startMin := 0

	nowHr := 18
	nowMin := 59

	now := time.Now()
	pointInTime := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin, 0, 0, scope.GetLocation())

	_, _, service := createTestScheduleService(pointInTime)

	startDate := time.Date(now.Year(), now.Month(), now.Day()-1, startHr, startMin, 0, 0, scope.GetLocation())

	schedule := &model.Schedule{
		StartDate: startDate,
		Frequency: common.SCHEDULE_FREQUENCY_DAILY}

	result := service.IsScheduled(schedule, 43200)
	assert.Equal(t, false, result)
}

func TestIsScheduled_YesterdayAtStartTime(t *testing.T) {

	_, scope := NewUnitTestContext()

	startHr := 19
	startMin := 0

	nowHr := 19
	nowMin := 00

	now := time.Now()
	pointInTime := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin, 0, 0, scope.GetLocation())

	_, _, service := createTestScheduleService(pointInTime)

	startDate := time.Date(now.Year(), now.Month(), now.Day()-1, startHr, startMin, 0, 0, scope.GetLocation())

	schedule := &model.Schedule{
		StartDate: startDate,
		Frequency: common.SCHEDULE_FREQUENCY_DAILY}

	result := service.IsScheduled(schedule, 43200)
	assert.Equal(t, true, result)
}

func TestIsScheduled_YesterdayOneMinuteAfterStartTime(t *testing.T) {

	_, scope := NewUnitTestContext()

	startHr := 19
	startMin := 0

	nowHr := 19
	nowMin := 01

	now := time.Now()
	pointInTime := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin, 0, 0, scope.GetLocation())

	_, _, service := createTestScheduleService(pointInTime)

	startDate := time.Date(now.Year(), now.Month(), now.Day()-1, startHr, startMin, 0, 0, scope.GetLocation())

	schedule := &model.Schedule{
		StartDate: startDate,
		Frequency: common.SCHEDULE_FREQUENCY_DAILY}

	result := service.IsScheduled(schedule, 43200)
	assert.Equal(t, true, result)
}

func newTestScheduleService() (common.Scope, *MockScheduleDAO, service.ScheduleService) {
	app, scope := NewUnitTestContext()
	dao := NewMockScheduleDAO()
	mapper := mapper.NewScheduleMapper()
	configService := NewMockConfigService()
	service := service.NewScheduleService(app, dao, mapper, configService)
	return scope, dao, service
}

func createTestScheduleService(now time.Time) (common.Scope, *MockScheduleDAO, service.ScheduleService) {
	app, scope := NewUnitTestContext()
	dao := NewMockScheduleDAO()
	mapper := mapper.NewScheduleMapper()
	configService := NewMockConfigService()
	service, _ := service.CreateScheduleService(app, dao, mapper, now, configService)
	return scope, dao, service
}
