package test

/*
func TestRoomChannelGreaterThanConditionalActivatesSwitch(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 0

	roomState := &entity.Room{
		Humidity0: 75.0,
		Channels:  &entity.RoomChannels{}}

	channels := []common.Channel{
		&model.Channel{
			ID:           1,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "test",
			Enable:       true,
			Notify:       true,
			Condition:    "humidity0 > 55"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", testChannelID, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", testChannelID, 1)

	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestRoomChannelGreaterThanConditionalDoesntActivateSwitch(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 0

	roomState := &entity.Room{
		Humidity0: 56.0,
		Channels: &entity.RoomChannels{
			Channel0: 1}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Dehumidifier",
			Enable:       true,
			Notify:       true,
			Condition:    "humidity0 > 55"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertNumberOfCalls(t, "RoomStatus", 1)
	client.AssertNumberOfCalls(t, "Switch", 0)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 0)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestRoomChannelLessThanConditionalActivatesSwitch(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 60,
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Test Heater",
			Enable:       true,
			Notify:       false,
			Condition:    "tempF0 < 70"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", testChannelID, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", testChannelID, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestRoomChannelLessThanConditionalDoesntActivateSwitch(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	roomState := &entity.Room{
		TempF0: 90,
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    0,
			Name:         "Test Heater",
			Enable:       true,
			Notify:       false,
			Condition:    "tempF0 < 70"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertNumberOfCalls(t, "Switch", 0)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestRoomChannelGreaterThanOrEqualToConditionalActivatesSwitchWhenEqual(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 70,
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Test Heater",
			Enable:       true,
			Notify:       false,
			Condition:    "tempF0 >= 70"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", 5, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", 5, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestRoomChannelGreaterThanOrEqualToConditionalActivatesSwitchWhenGreaterThan(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 71,
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Test Channel",
			Enable:       true,
			Notify:       false,
			Condition:    "tempF0 >= 70"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", testChannelID, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", testChannelID, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestRoomChannelLessThanOrEqualToConditionalActivatesSwitchWhenLessThan(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 60,
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Test Channel",
			Enable:       true,
			Notify:       false,
			Condition:    "tempF0 <= 70"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", 5, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", 5, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestRoomChannelLessThanOrEqualToConditionalActivatesSwitchWhenEqual(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 70,
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Test Channel",
			Enable:       true,
			Notify:       false,
			Condition:    "tempF0 <= 70"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", 5, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", 5, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestRoomChannelEqualsConditionalActivatesSwitch(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 70,
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Test Channel",
			Enable:       true,
			Notify:       false,
			Condition:    "tempF0 = 70"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", 5, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", 5, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestRoomChannelEqualsConditionalDoesntActivateSwitch(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 90,
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Test Channel",
			Enable:       true,
			Notify:       false,
			Condition:    "tempF0 = 70"}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertNumberOfCalls(t, "Switch", 0)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestLightsSwitchOnWhenScheduledNow(t *testing.T) {

	ctx, client, notificationService, _, scheduleService, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 90,
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	endDate := scheduleService.GetNow().AddDate(0, 0, 1)
	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Lights",
			Enable:       true,
			Notify:       true,
			Schedule: []common.ScheduleConfig{
				&model.Schedule{
					StartDate: *scheduleService.GetNow(),
					EndDate:   &endDate}}}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", testChannelID, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", testChannelID, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)

	schedule := channels[0].GetSchedule()[0]
	assert.Equal(t, true, scheduleService.IsScheduled(schedule, 0))
}

func TestLightsSwitchOffWhenTimerExpires(t *testing.T) {

	ctx, client, notificationService, _, scheduleService, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 90,
		Channels: &entity.RoomChannels{
			Channel5: 1}}

	now := scheduleService.GetNow()
	startDate := now.Add(time.Duration(-1) * time.Minute)

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Lights",
			Enable:       true,
			Notify:       true,
			Duration:     1,
			Schedule: []common.ScheduleConfig{
				&model.Schedule{
					StartDate: startDate}}}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", 5, 0).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", 5, 0)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)

	schedule := channels[0].GetSchedule()[0]
	assert.Equal(t, false, scheduleService.IsScheduled(schedule, 60))
}

func TestLightsSwitchOffWhenEndDateExpires(t *testing.T) {

	ctx, client, notificationService, _, scheduleService, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		TempF0: 90,
		Channels: &entity.RoomChannels{
			Channel5: 1}}

	endDate := scheduleService.GetNow()
	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Lights",
			Enable:       true,
			Notify:       true,
			Duration:     0,
			Schedule: []common.ScheduleConfig{
				&model.Schedule{
					StartDate: scheduleService.GetNow().Add(time.Duration(-1) * time.Second),
					EndDate:   endDate}}}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", 5, 0).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", 5, 0)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)

	schedule := channels[0].GetSchedule()[0]
	assert.Equal(t, false, scheduleService.IsScheduled(schedule, 0))
}


func TestLightsSwitchStaysOnWhenScheduled(t *testing.T) {

	//ctx := NewUnitTestContext()
	//now := scheduleService.GetNow().In(ctx.GetLocation())
	//client, _, _, scheduleService, roomService := createTestRoomServiceWithSchedule(ctx, now)

	ctx, client, _, _, scheduleService, roomService := createTestRoomService()

	now := scheduleService.GetNow()
	startDate := now.Add(time.Duration(-1) * time.Second)
	endDate := now.Add(time.Duration(1) * time.Second)

	testChannelID := 5

	roomState := &entity.Room{
		Channels: &entity.RoomChannels{
			Channel5: 1}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Lights",
			Enable:       true,
			Notify:       true,
			Schedule: []common.ScheduleConfig{
				&model.Schedule{
					StartDate: startDate,
					EndDate:   &endDate}}}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertNumberOfCalls(t, "Switch", 0)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)

	schedule := channels[0].GetSchedule()[0]
	assert.Equal(t, true, scheduleService.IsScheduled(schedule, 0))
}

func TestLightsSwitchStaysOffWhenTimeBeforeScheduled(t *testing.T) {

	ctx, client, _, _, scheduleService, roomService := createTestRoomService()

	testChannelID := 5

	now := scheduleService.GetNow()
	startDate := now.Add(time.Duration(1) * time.Minute)
	endDate := now.Add(time.Duration(2) * time.Minute)

	roomState := &entity.Room{
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Lights",
			Enable:       true,
			Notify:       true,
			Schedule: []common.ScheduleConfig{
				&model.Schedule{
					StartDate: startDate,
					EndDate:   &endDate}}}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertNumberOfCalls(t, "Switch", 0)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)

	schedule := channels[0].GetSchedule()[0]
	assert.Equal(t, false, scheduleService.IsScheduled(schedule, 0))
}

func TestLightsSwitchStaysOffWhenTimeAfterScheduled(t *testing.T) {

	ctx, client, _, _, scheduleService, roomService := createTestRoomService()

	testChannelID := 5

	now := scheduleService.GetNow()
	startDate := now.Add(time.Duration(-2) * time.Minute)
	endDate := now.Add(time.Duration(-1) * time.Minute)

	roomState := &entity.Room{
		Channels: &entity.RoomChannels{
			Channel5: 0}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Lights",
			Enable:       true,
			Notify:       true,
			Schedule: []common.ScheduleConfig{
				&model.Schedule{
					StartDate: startDate,
					EndDate:   &endDate}}}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertNumberOfCalls(t, "Switch", 0)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)

	schedule := channels[0].GetSchedule()[0]
	assert.Equal(t, false, scheduleService.IsScheduled(schedule, 0))
}

func TestChannelDebounceHighWorks(t *testing.T) {

	ctx, client, _, channelService, _, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		Channels: &entity.RoomChannels{}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Dehumidifier",
			Enable:       true,
			Notify:       true,
			Condition:    "humidity0 > 55.00",
			Debounce:     10}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertNumberOfCalls(t, "Switch", 0)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
}

func TestChannelDebounceLowDoesntActivate(t *testing.T) {

	ctx, client, _, channelService, _, roomService := createTestRoomService()

	testChannelID := 5

	roomState := &entity.Room{
		Humidity0: 50.0,
		Channels:  &entity.RoomChannels{}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Dehumidifier",
			Enable:       true,
			Notify:       true,
			Condition:    "humidity0 > 55.00",
			Debounce:     10}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertNumberOfCalls(t, "Switch", 0)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
}

func TestChannelDebounceExceededHighActivates(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 3

	roomState := &entity.Room{
		Humidity0: 66.0,
		Channels:  &entity.RoomChannels{}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Dehumidifier",
			Enable:       true,
			Notify:       true,
			Condition:    "humidity0 > 55.00",
			Debounce:     10}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", testChannelID, 1).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", testChannelID, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestChannelDebounceLowTurnsOffWhenAlreadyOn(t *testing.T) {

	ctx, client, notificationService, channelService, _, roomService := createTestRoomService()

	testChannelID := 3

	roomState := &entity.Room{
		Humidity0: 20.0,
		Channels: &entity.RoomChannels{
			Channel3: 1}}

	channels := []common.Channel{
		&model.Channel{
			ID:           0,
			ControllerID: common.CONTROLLER_TYPE_ID_ROOM,
			ChannelID:    testChannelID,
			Name:         "Fake Dehumidifier",
			Enable:       true,
			Notify:       true,
			Condition:    "humidity0 > 55.00",
			Debounce:     10}}

	config := &model.Config{
		Room: &model.Room{
			Enable:   true,
			Notify:   true,
			Channels: channels}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	channelService.On("GetAll", ctx.GetUser(), common.CONTROLLER_TYPE_ID_ROOM).Return(channels, nil)
	client.On("RoomStatus").Return(roomState, nil)
	client.On("Switch", testChannelID, 0).Return(&entity.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageChannels()

	client.AssertCalled(t, "Switch", testChannelID, 0)
	client.AssertNumberOfCalls(t, "Switch", 1)
	client.AssertNumberOfCalls(t, "RoomStatus", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestMetric(t *testing.T) {

	ctx, client, notificationService, _, _, roomService := createTestRoomService()

	roomState := &entity.Room{Photo: 300}

	metrics := []common.Metric{
		&model.Metric{
			ID:           1,
			ControllerID: 2,
			Key:          "photo",
			Name:         "Light Sensor",
			Enable:       true,
			Notify:       true,
			AlarmLow:     500,
			AlarmHigh:    1000}}

	config := &model.Config{
		Room: &model.Room{
			Enable:  true,
			Notify:  true,
			Metrics: metrics}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageMetrics()

	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)
	notificationService.AssertCalled(t, "Enqueue", mock.Anything)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestMetricAlarmLow(t *testing.T) {

	ctx, client, notificationService, _, _, roomService := createTestRoomService()

	roomState := &entity.Room{Photo: 300}

	metrics := []common.Metric{
		&model.Metric{
			Key:       "photo",
			Name:      "Light Sensor",
			Enable:    true,
			Notify:    true,
			AlarmLow:  500,
			AlarmHigh: 1000}}

	config := &model.Config{
		Room: &model.Room{
			Enable:  true,
			Notify:  true,
			Metrics: metrics}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageMetrics()

	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)
	notificationService.AssertCalled(t, "Enqueue", mock.Anything)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestMetricAlarmHigh(t *testing.T) {

	ctx, client, notificationService, _, _, roomService := createTestRoomService()

	roomState := &entity.Room{Photo: 1100}

	metrics := []common.Metric{
		&model.Metric{
			ID:        0,
			Key:       "photo",
			Name:      "Light Sensor",
			Enable:    true,
			Notify:    true,
			AlarmLow:  500,
			AlarmHigh: 1000}}

	config := &model.Config{
		Room: &model.Room{
			Enable:  true,
			Notify:  true,
			Metrics: metrics}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	roomService.Poll()
	roomService.ManageMetrics()

	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)
	notificationService.AssertCalled(t, "Enqueue", mock.Anything)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestMetricAlarmNormal(t *testing.T) {

	ctx, client, notificationService, _, _, roomService := createTestRoomService()

	roomState := &entity.Room{Photo: 800}

	metrics := []common.Metric{
		&model.Metric{
			ID:        0,
			Key:       "photo",
			Name:      "Light Sensor",
			Enable:    true,
			Notify:    true,
			AlarmLow:  500,
			AlarmHigh: 1000}}

	config := &model.Config{
		Room: &model.Room{
			Enable:  true,
			Notify:  true,
			Metrics: metrics}}

	ctx.SetConfig(config)
	ctx.GetState().SetRoom(roomState)

	client.On("RoomStatus").Return(roomState, nil)

	roomService.Poll()
	roomService.ManageMetrics()

	notificationService.AssertNumberOfCalls(t, "Enqueue", 0)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func createTestRoomService() (common.Context, *MockController, *MockNotificationService, *MockChannelService,
	service.ScheduleService, common.ControllerService) {

	ctx := NewUnitTestContext()
	dao := NewMockDynamicDAO()
	scheduleDAO := NewMockScheduleDAO()
	client := NewMockController()
	mailer := NewMockMailer(ctx)
	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	scheduleMapper := mapper.NewScheduleMapper()
	controllerMapper := mapper.NewControllerMapper(metricMapper, channelMapper)
	notificationService := NewMockNotificationService(ctx, mailer)
	eventLogService := NewMockEventLogService(ctx, nil, "test")
	channelService := NewMockChannelService()
	configService := NewMockConfigService()
	conditionService := NewMockConditionService()
	scheduleService := service.NewScheduleService(ctx, scheduleDAO, scheduleMapper, configService)
	microcontrollerService, _ := service.NewMicroControllerService(ctx, dao, client, controllerMapper,
		eventLogService, notificationService, conditionService, scheduleService)

	return ctx, client, notificationService, channelService, scheduleService, microcontrollerService
}

func createTestRoomServiceWithSchedule(ctx common.Context, now time.Time) (*MockController, *MockNotificationService,
	*MockChannelService, service.ScheduleService, common.ControllerService) {

	dao := NewMockDynamicDAO()
	scheduleDAO := NewMockScheduleDAO()
	client := NewMockController()
	mailer := NewMockMailer(ctx)
	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	scheduleMapper := mapper.NewScheduleMapper()
	controllerMapper := mapper.NewControllerMapper(metricMapper, channelMapper)
	notificationService := NewMockNotificationService(ctx, mailer)
	eventLogService := NewMockEventLogService(ctx, nil, "test")
	channelService := NewMockChannelService()
	configService := NewMockConfigService()
	conditionService := NewMockConditionService()
	scheduleService, _ := service.CreateScheduleService(ctx, scheduleDAO, scheduleMapper, now, configService)
	microcontrollerService, _ := service.NewMicroControllerService(ctx, dao, client, controllerMapper, eventLogService,
		notificationService, conditionService, scheduleService)
	return client, notificationService, channelService, scheduleService, microcontrollerService
}
*/
