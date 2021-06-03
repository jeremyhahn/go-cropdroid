// +build broken

package test

import (
	"testing"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/entity"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/mock"
)

func TestReservoirChannelGreaterThanConditionalActivatesSwitch(t *testing.T) {

	ctx, client, notificationService, reservoirService := createTestReservoirService()

	testReservoir := &entity.Reservoir{
		ResTemp:  70.0,
		Channels: &entity.ReservoirChannels{}}

	ctx.SetConfig(&model.Config{
		Reservoir: &model.Reservoir{
			Enable: true,
			Notify: true,
			Channels: []common.Channel{
				&model.Channel{
					ChannelID: 2,
					Name:      "Chiller",
					Enable:    true,
					Notify:    true,
					Condition: "resTemp > 62"}}}})

	ctx.GetState().SetReservoir(testReservoir)

	client.On("ReservoirStatus").Return(testReservoir, nil)
	client.On("Switch", 2, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	reservoirService.Poll()
	reservoirService.ManageChannels()

	client.AssertCalled(t, "Switch", 2, 1)

	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "ReservoirStatus", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestReservoirMetricAlarmLow(t *testing.T) {

	ctx, client, notificationService, roomService := createTestReservoirService()

	testReservoir := &entity.Reservoir{ResTemp: 90}

	ctx.SetConfig(&model.Config{
		Reservoir: &model.Reservoir{
			Enable: true,
			Notify: true,
			Metrics: []common.Metric{
				&model.Metric{
					ID:        1,
					Key:       "resTemp",
					Name:      "Water Temperature",
					Enable:    true,
					Notify:    true,
					AlarmLow:  50,
					AlarmHigh: 75}}}})
	ctx.GetState().SetReservoir(testReservoir)

	client.On("ReservoirStatus").Return(testReservoir, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageMetrics()

	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)
	notificationService.AssertCalled(t, "Enqueue", mock.Anything)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func createTestReservoirService() (common.Context, *MockReservoirController, *MockNotificationService, *service.Reservoir) {
	ctx := NewUnitTestContext()
	dao := NewMockReservoirDAO(ctx)
	scheduleDAO := NewMockScheduleDAO()
	client := NewMockReservoirController()
	mailer := NewMockMailer(ctx)
	notificationService := NewMockNotificationService(ctx, mailer)
	eventLogService := NewMockEventLogService(ctx, nil, "test")
	scheduleMapper := mapper.NewScheduleMapper()
	scheduleService := NewMockScheduleService(ctx, scheduleDAO, scheduleMapper)
	return ctx, client, notificationService,
		service.NewReservoirService(ctx, dao, client, eventLogService, notificationService, scheduleService).(*service.Reservoir)
}
