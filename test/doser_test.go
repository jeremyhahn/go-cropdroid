// +build broken

package test

import (
	"testing"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/mock"
)

func TestDoserChannelGreaterThanConditionalActivatesSwitch(t *testing.T) {

	ctx, client, notificationService, DoserService := createTestDoserService()

	reservoirState := &entity.Reservoir{
		PH:       6.1,
		Channels: &entity.ReservoirChannels{}}

	doserState := &entity.Doser{
		Channels: &entity.DoserChannels{Channel0: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:        0,
			Name:      "phDOWN",
			Enable:    true,
			Notify:    true,
			Condition: "reservoir.ph > 5.9"}}

	ctx.SetConfig(&model.Config{
		Doser: &model.Doser{
			Enable:   true,
			Notify:   true,
			Channels: channels}})

	ctx.GetState().SetReservoir(reservoirState)
	ctx.GetState().SetDoser(doserState)

	client.On("DoserStatus").Return(doserState, nil)
	client.On("Switch", 0, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	DoserService.Poll()
	DoserService.ManageChannels()

	client.AssertCalled(t, "Switch", 0, 1)

	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "DoserStatus", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestDoserChannelLessThanConditionalActivatesSwitch(t *testing.T) {

	ctx, client, notificationService, DoserService := createTestDoserService()

	reservoirState := &entity.Reservoir{
		ORP:      280,
		Channels: &entity.ReservoirChannels{}}

	doserState := &entity.Doser{
		Channels: &entity.DoserChannels{Channel0: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:        0,
			Name:      "Oxidizer",
			Enable:    true,
			Notify:    true,
			Condition: "reservoir.orp < 300"}}

	ctx.SetConfig(&model.Config{
		Doser: &model.Doser{
			Enable:   true,
			Notify:   true,
			Channels: channels}})

	ctx.GetState().SetReservoir(reservoirState)
	ctx.GetState().SetDoser(doserState)

	client.On("DoserStatus").Return(doserState, nil)
	client.On("Switch", 0, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	DoserService.Poll()
	DoserService.ManageChannels()

	client.AssertCalled(t, "Switch", 0, 1)

	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "DoserStatus", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func createTestDoserService() (common.Context, *MockDoserController, *MockNotificationService, *service.Doser) {
	ctx := NewUnitTestContext()
	dao := NewMockDoserDAO(ctx)
	scheduleDAO := NewMockScheduleDAO()
	client := new(MockDoserController)
	mailer := NewMockMailer(ctx)
	channelMapper := mapper.NewChannelMapper()
	doserMapper := mapper.NewDoserMapper(channelMapper)
	scheduleMapper := mapper.NewScheduleMapper()
	eventLogService := NewMockEventLogService(ctx, nil, "test")
	notificationService := NewMockNotificationService(ctx, mailer)
	scheduleService := NewMockScheduleService(ctx, scheduleDAO, scheduleMapper)
	return ctx, client, notificationService,
		service.NewDoserService(ctx, dao, doserMapper, client, eventLogService, notificationService, scheduleService).(*service.Doser)
}
