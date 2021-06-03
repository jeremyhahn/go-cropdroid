// +build broken

package test

import (
	"strconv"
	"testing"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/assert"
)

func TestConfigDataStructure(t *testing.T) {

	orgDAO, controllerDAO, configDAO, metricDAO, channelDAO, scheduleDAO, configService := createConfigService()

	orgID := 1
	serverControllerID := 1
	strServerControllerID := "1"
	roomControllerID := 2
	reservoirControllerID := 3
	doserControllerID := 4
	testOrg := &entity.Organization{ID: orgID, Name: "Test Org"}
	testControllers := createTestControllers(orgID, serverControllerID, roomControllerID, reservoirControllerID, doserControllerID)
	testServerName := &entity.Config{
		ID:           1,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "name",
		Value:        "Test CropDroid Server"}
	testServerInterval := &entity.Config{
		ID:           2,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "interval",
		Value:        "61"}
	testServerTimezone := &entity.Config{
		ID:           3,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "timezone",
		Value:        "America/New_York"}
	testServerMode := &entity.Config{
		ID:           4,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "mode",
		Value:        "unittest"}
	testSmtpEnable := &entity.Config{
		ID:           5,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "smtp.enable",
		Value:        "false"}
	testSmtpHost := &entity.Config{
		ID:           6,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "smtp.host",
		Value:        "smtp.myhost.com"}
	testSmtpPort := &entity.Config{
		ID:           7,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "smtp.port",
		Value:        "587"}
	testSmtpUsername := &entity.Config{
		ID:           8,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "smtp.username",
		Value:        "test"}
	testSmtpPassword := &entity.Config{
		ID:           9,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "smtp.password",
		Value:        "test"}
	testSmtpRecipient := &entity.Config{
		ID:           10,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "smtp.recipient",
		Value:        "test"}
	testRoomEnable := &entity.Config{
		ID:           11,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: roomControllerID,
		Key:          "room.enable",
		Value:        "true"}
	testRoomNotify := &entity.Config{
		ID:           12,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: roomControllerID,
		Key:          "room.notify",
		Value:        "false"}
	testRoomURI := &entity.Config{
		ID:           13,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: serverControllerID,
		Key:          "room.uri",
		Value:        "http://myroom.local"}
	testRoomVideo := &entity.Config{
		ID:           14,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: roomControllerID,
		Key:          "room.video",
		Value:        "http://mydvr.local/cam1"}
	testReservoirEnable := &entity.Config{
		ID:           15,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: reservoirControllerID,
		Key:          "reservoir.enable",
		Value:        "true"}
	testReservoirNotify := &entity.Config{
		ID:           16,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: reservoirControllerID,
		Key:          "reservoir.notify",
		Value:        "true"}
	testReservoirURI := &entity.Config{
		ID:           17,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: reservoirControllerID,
		Key:          "reservoir.uri",
		Value:        "http://myreservoir.local"}
	testReservoirGallons := &entity.Config{
		ID:           18,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: reservoirControllerID,
		Key:          "reservoir.gallons",
		Value:        "50"}
	testReservoirWaterChangeEnable := &entity.Config{
		ID:           19,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: reservoirControllerID,
		Key:          "reservoir.waterchange.enable",
		Value:        "false"}
	testReservoirWaterChangeNotify := &entity.Config{
		ID:           20,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: reservoirControllerID,
		Key:          "reservoir.waterchange.notify",
		Value:        "false"}
	testDoserEnable := &entity.Config{
		ID:           21,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: doserControllerID,
		Key:          "doser.enable",
		Value:        "true"}
	testDoserNotify := &entity.Config{
		ID:           22,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: doserControllerID,
		Key:          "doser.notify",
		Value:        "true"}
	testDoserURI := &entity.Config{
		ID:           23,
		UserID:       0,
		OrgID:        orgID,
		ControllerID: doserControllerID,
		Key:          "doser.uri",
		Value:        "http://mydoser.local"}

	testRoomMetricEntities := createFakeRoomMetricEntities()
	testRoomChannelEntities := createFakeRoomChannelEntities()

	testReservoirMetricEntities := createFakeReservoirMetricEntities()
	testReservoirChannelEntities := createFakeReservoirChannelEntities()

	testDoserChannelEntities := createFakeDoserChannelEntities()

	testScheduleEntities := createFakeScheduleEntities()

	orgDAO.On("First").Return(testOrg, nil)
	controllerDAO.On("GetByOrgId", orgID).Return(testControllers, nil)
	configDAO.On("Get", strServerControllerID, "name").Return(testServerName, nil)
	configDAO.On("Get", strServerControllerID, "interval").Return(testServerInterval, nil)
	configDAO.On("Get", strServerControllerID, "timezone").Return(testServerTimezone, nil)
	configDAO.On("Get", strServerControllerID, "mode").Return(testServerMode, nil)

	configDAO.On("Get", strServerControllerID, "smtp.enable").Return(testSmtpEnable, nil)
	configDAO.On("Get", strServerControllerID, "smtp.host").Return(testSmtpHost, nil)
	configDAO.On("Get", strServerControllerID, "smtp.port").Return(testSmtpPort, nil)
	configDAO.On("Get", strServerControllerID, "smtp.username").Return(testSmtpUsername, nil)
	configDAO.On("Get", strServerControllerID, "smtp.password").Return(testSmtpPassword, nil)
	configDAO.On("Get", strServerControllerID, "smtp.recipient").Return(testSmtpRecipient, nil)

	configDAO.On("Get", strServerControllerID, "room.enable").Return(testRoomEnable, nil)
	configDAO.On("Get", strServerControllerID, "room.notify").Return(testRoomNotify, nil)
	configDAO.On("Get", strServerControllerID, "room.uri").Return(testRoomURI, nil)
	configDAO.On("Get", strServerControllerID, "room.video").Return(testRoomVideo, nil)

	configDAO.On("Get", strServerControllerID, "reservoir.enable").Return(testReservoirEnable, nil)
	configDAO.On("Get", strServerControllerID, "reservoir.notify").Return(testReservoirNotify, nil)
	configDAO.On("Get", strServerControllerID, "reservoir.uri").Return(testReservoirURI, nil)
	configDAO.On("Get", strServerControllerID, "reservoir.gallons").Return(testReservoirGallons, nil)
	configDAO.On("Get", strServerControllerID, "reservoir.waterchange.enable").Return(testReservoirWaterChangeEnable, nil)
	configDAO.On("Get", strServerControllerID, "reservoir.waterchange.notify").Return(testReservoirWaterChangeNotify, nil)

	configDAO.On("Get", strServerControllerID, "doser.enable").Return(testDoserEnable, nil)
	configDAO.On("Get", strServerControllerID, "doser.notify").Return(testDoserNotify, nil)
	configDAO.On("Get", strServerControllerID, "doser.uri").Return(testDoserURI, nil)

	metricDAO.On("GetByControllerID", roomControllerID).Return(testRoomMetricEntities, nil)
	channelDAO.On("GetByControllerID", roomControllerID).Return(testRoomChannelEntities, nil)

	metricDAO.On("GetByControllerID", reservoirControllerID).Return(testReservoirMetricEntities, nil)
	channelDAO.On("GetByControllerID", reservoirControllerID).Return(testReservoirChannelEntities, nil)

	channelDAO.On("GetByControllerID", doserControllerID).Return(testDoserChannelEntities, nil)

	scheduleDAO.On("GetByChannelID", 1).Return(testScheduleEntities, nil)

	expectedInterval, err := strconv.Atoi(testServerInterval.GetValue())
	assert.Nil(t, err)

	// CropDroid server configs
	config, err := configService.Build()
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, testServerName.GetValue(), config.GetName())
	assert.Equal(t, expectedInterval, config.GetInterval())
	assert.Equal(t, testServerTimezone.GetValue(), config.GetTimezone())
	assert.Equal(t, testServerMode.GetValue(), config.GetMode())

	smtpEnabled, _ := strconv.ParseBool(testSmtpEnable.GetValue())
	assert.Equal(t, smtpEnabled, config.GetSmtp().IsEnabled())
	assert.Equal(t, testSmtpHost.GetValue(), config.GetSmtp().GetHost())
	assert.Equal(t, testSmtpPort.GetValue(), config.GetSmtp().GetPort())
	assert.Equal(t, testSmtpUsername.GetValue(), config.GetSmtp().GetUsername())
	assert.Equal(t, testSmtpPassword.GetValue(), config.GetSmtp().GetPassword())
	assert.Equal(t, testSmtpRecipient.GetValue(), config.GetSmtp().GetRecipient())

	// Room controller configs
	roomEnabled, _ := strconv.ParseBool(testRoomEnable.GetValue())
	roomNotify, _ := strconv.ParseBool(testRoomNotify.GetValue())
	assert.Equal(t, roomEnabled, config.GetRoom().IsEnabled())
	assert.Equal(t, roomNotify, config.GetRoom().IsNotify())
	assert.Equal(t, testRoomURI.GetValue(), config.GetRoom().GetURI())
	//assert.Equal(t, testRoomVideo.GetValue(), config.GetRoom().GetVideoURL())
	assert.Equal(t, len(testRoomMetricEntities), len(config.GetRoom().GetMetrics()))
	assert.Equal(t, len(testRoomChannelEntities), len(config.GetRoom().GetChannels()))

	roomMetricEntity0 := testRoomMetricEntities[0]
	roomMetric0 := config.GetRoom().GetMetrics()[0]
	assert.Equal(t, roomMetricEntity0.GetID(), roomMetric0.GetID())
	assert.Equal(t, roomMetricEntity0.GetControllerID(), roomMetric0.GetControllerID())
	assert.Equal(t, roomMetricEntity0.IsEnabled(), roomMetric0.IsEnabled())
	assert.Equal(t, roomMetricEntity0.IsNotify(), roomMetric0.IsNotify())
	assert.Equal(t, roomMetricEntity0.GetKey(), roomMetric0.GetKey())
	assert.Equal(t, roomMetricEntity0.GetName(), roomMetric0.GetName())
	assert.Equal(t, roomMetricEntity0.GetUnit(), roomMetric0.GetUnit())
	assert.Equal(t, roomMetricEntity0.GetAlarmLow(), roomMetric0.GetAlarmLow())
	assert.Equal(t, roomMetricEntity0.GetAlarmHigh(), roomMetric0.GetAlarmHigh())
	assert.Equal(t, 0.0, roomMetric0.GetValue())

	roomChannelEntity0 := testRoomChannelEntities[0]
	roomChannel0 := config.GetRoom().GetChannels()[0]
	assert.Equal(t, roomChannelEntity0.GetID(), roomChannel0.GetID())
	assert.Equal(t, roomChannelEntity0.GetControllerID(), roomChannel0.GetControllerID())
	assert.Equal(t, roomChannelEntity0.GetChannelID(), roomChannel0.GetChannelID())
	assert.Equal(t, roomChannelEntity0.IsEnabled(), roomChannel0.IsEnabled())
	assert.Equal(t, roomChannelEntity0.IsNotify(), roomChannel0.IsNotify())
	assert.Equal(t, roomChannelEntity0.GetName(), roomChannel0.GetName())
	assert.Equal(t, roomChannelEntity0.GetCondition(), roomChannel0.GetCondition())
	assert.Equal(t, roomChannelEntity0.GetDuration(), roomChannel0.GetDuration())
	assert.Equal(t, roomChannelEntity0.GetDebounce(), roomChannel0.GetDebounce())
	assert.Equal(t, roomChannelEntity0.GetBackoff(), roomChannel0.GetBackoff())
	assert.Equal(t, 0, roomChannel0.GetValue())

	// Reservoir controller configs
	reservoirEnabled, _ := strconv.ParseBool(testReservoirEnable.GetValue())
	reservoirNotify, _ := strconv.ParseBool(testReservoirNotify.GetValue())
	reservoirGallons, _ := strconv.ParseInt(testReservoirGallons.GetValue(), 0, 64)
	waterChangeEnabled, _ := strconv.ParseBool(testReservoirWaterChangeEnable.GetValue())
	waterChangeNotify, _ := strconv.ParseBool(testReservoirWaterChangeNotify.GetValue())
	assert.Equal(t, reservoirEnabled, config.GetReservoir().IsEnabled())
	assert.Equal(t, reservoirNotify, config.GetReservoir().IsNotify())
	assert.Equal(t, testReservoirURI.GetValue(), config.GetReservoir().GetURI())
	assert.Equal(t, int(reservoirGallons), config.GetReservoir().GetGallons())
	assert.Equal(t, waterChangeEnabled, config.GetReservoir().GetWaterChangeConfig().IsEnabled())
	assert.Equal(t, waterChangeNotify, config.GetReservoir().GetWaterChangeConfig().IsNotify())
	assert.Equal(t, len(testReservoirMetricEntities), len(config.GetReservoir().GetMetrics()))
	assert.Equal(t, len(testReservoirChannelEntities), len(config.GetReservoir().GetChannels()))

	assert.NotNil(t, config.GetReservoir().GetWaterChangeConfig())
	assert.Equal(t, false, config.GetReservoir().GetWaterChangeConfig().IsEnabled())

	// Doser controller configs
	doserEnabled, _ := strconv.ParseBool(testDoserEnable.GetValue())
	doserNotify, _ := strconv.ParseBool(testDoserNotify.GetValue())
	assert.Equal(t, doserEnabled, config.GetDoser().IsEnabled())
	assert.Equal(t, doserNotify, config.GetDoser().IsNotify())
	assert.Equal(t, testDoserURI.GetValue(), config.GetDoser().GetURI())
	assert.Equal(t, len(testDoserChannelEntities), len(config.GetDoser().GetChannels()))
}

func createTestControllers(orgID, serverControllerID, roomControllerID, reservoirControllerID, doserControllerID int) []entity.Controller {
	return []entity.Controller{
		{
			ID:              serverControllerID,
			OrganizationID:  orgID,
			Type:            common.CONTROLLER_TYPE_SERVER,
			Description:     "Test CropDroid Server",
			HardwareVersion: "test-v1",
			FirmwareVersion: "test-v1",
			Metrics:         nil,
			Channels:        nil},
		{
			ID:              roomControllerID,
			OrganizationID:  orgID,
			Type:            common.CONTROLLER_TYPE_ROOM,
			Description:     "Test Room Controller",
			HardwareVersion: "test-v2",
			FirmwareVersion: "test-v2",
			Metrics:         nil,
			Channels:        nil},
		{
			ID:              reservoirControllerID,
			OrganizationID:  orgID,
			Type:            common.CONTROLLER_TYPE_RESERVOIR,
			Description:     "Test Reservoir Controller",
			HardwareVersion: "test-v3",
			FirmwareVersion: "test-v3",
			Metrics:         nil,
			Channels:        nil},
		{
			ID:              doserControllerID,
			OrganizationID:  orgID,
			Type:            common.CONTROLLER_TYPE_DOSER,
			Description:     "Test Doser Controller",
			HardwareVersion: "test-v4",
			FirmwareVersion: "test-v4",
			Metrics:         nil,
			Channels:        nil}}
}

func createConfigService() (*MockOrganizationDAO, *MockControllerDAO,
	*MockConfigDAO, *MockMetricDAO, *MockChannelDAO, *MockScheduleDAO, service.ConfigService) {

	ctx := NewUnitTestContext()
	organizationDAO := NewMockOrganizationDAO()
	userDAO := NewMockUserDAO()
	controllerDAO := NewMockControllerDAO()
	configDAO := NewMockConfigDAO()
	metricDAO := NewMockMetricDAO()
	channelDAO := NewMockChannelDAO()
	scheduleDAO := NewMockScheduleDAO()
	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	scheduleMapper := mapper.NewScheduleMapper()
	return organizationDAO, controllerDAO, configDAO, metricDAO, channelDAO, scheduleDAO,
		service.NewConfigService(ctx, organizationDAO, userDAO, controllerDAO,
			configDAO, metricDAO, channelDAO, scheduleDAO, metricMapper, channelMapper, scheduleMapper)
}

func createFakeRoomMetricEntities() []entity.Metric {
	metrics := createFakeRoomMetrics()
	entities := make([]entity.Metric, len(metrics))
	for i, metric := range metrics {
		_metric := mapper.NewMetricMapper().MapModelToEntity(metric)
		entities[i] = *_metric
	}
	return entities
}

func createFakeRoomMetrics() []common.Metric {
	return []common.Metric{
		&model.Metric{
			ID:        1,
			Key:       "temp0",
			Name:      "Fake Temperature Metric",
			Enable:    true,
			Notify:    true,
			AlarmLow:  65,
			AlarmHigh: 75},
		&model.Metric{
			ID:        2,
			Key:       "humidity0",
			Name:      "Fake Humidity Metric",
			Enable:    false,
			Notify:    true,
			Unit:      "°",
			AlarmLow:  10,
			AlarmHigh: 20},
		&model.Metric{
			ID:        3,
			Key:       "foo",
			Name:      "Foo Metric",
			Enable:    true,
			Notify:    false,
			Unit:      "°",
			AlarmLow:  50,
			AlarmHigh: 100,
			Value:     90}}
}

func createFakeRoomChannelEntities() []entity.Channel {
	channels := createFakeRoomChannels()
	entities := make([]entity.Channel, len(channels))
	for i, channel := range channels {
		_channel := mapper.NewChannelMapper().MapModelToEntity(channel)
		entities[i] = *_channel
	}
	return entities
}

func createFakeRoomChannels() []common.Channel {
	return []common.Channel{
		&model.Channel{
			ID:           1,
			ControllerID: 2,
			ChannelID:    0,
			Name:         "Test Channel",
			Enable:       true,
			Notify:       true,
			Condition:    "",
			Duration:     0,
			Debounce:     0,
			Backoff:      0},
		&model.Channel{
			ID:           1,
			ControllerID: 2,
			ChannelID:    1,
			Name:         "Test Channel 2",
			Enable:       true,
			Notify:       true,
			Condition:    "foo > 80",
			Duration:     0,
			Debounce:     0,
			Backoff:      0},
		&model.Channel{
			ID:           1,
			ControllerID: 2,
			ChannelID:    2,
			Name:         "Test Channel 3",
			Enable:       true,
			Notify:       true,
			Condition:    "foo > 80",
			Duration:     0,
			Debounce:     0,
			Backoff:      0}}
}

// reservoir

func createFakeReservoirMetricEntities() []entity.Metric {
	metrics := createFakeReservoirMetrics()
	entities := make([]entity.Metric, len(metrics))
	for i, metric := range metrics {
		_metric := mapper.NewMetricMapper().MapModelToEntity(metric)
		entities[i] = *_metric
	}
	return entities
}

func createFakeReservoirMetrics() []common.Metric {
	return []common.Metric{
		&model.Metric{
			ID:        1,
			Enable:    true,
			Notify:    true,
			Key:       "temp",
			Name:      "Fake Temperature Metric",
			Unit:      "°",
			AlarmLow:  65,
			AlarmHigh: 75},
		&model.Metric{
			ID:        2,
			Key:       "pH",
			Name:      "Fake Humidity Metric",
			Enable:    false,
			Notify:    true,
			AlarmLow:  5.5,
			AlarmHigh: 6.1},
		&model.Metric{
			ID:        3,
			Key:       "bar",
			Name:      "Bar Metric",
			Enable:    true,
			Notify:    false,
			AlarmLow:  50,
			AlarmHigh: 100,
			Value:     90}}
}

func createFakeReservoirChannelEntities() []entity.Channel {
	channels := createFakeReservoirChannels()
	entities := make([]entity.Channel, len(channels))
	for i, channel := range channels {
		_channel := mapper.NewChannelMapper().MapModelToEntity(channel)
		entities[i] = *_channel
	}
	return entities
}

func createFakeReservoirChannels() []common.Channel {
	return []common.Channel{
		&model.Channel{
			ID:           1,
			ControllerID: 2,
			ChannelID:    0,
			Name:         "Test Channel",
			Enable:       true,
			Notify:       true,
			Condition:    "",
			Duration:     0,
			Debounce:     0,
			Backoff:      0},
		&model.Channel{
			ID:           1,
			ControllerID: 2,
			ChannelID:    1,
			Name:         "Test Channel 2",
			Enable:       true,
			Notify:       true,
			Condition:    "foo > 80",
			Duration:     0,
			Debounce:     0,
			Backoff:      0},
		&model.Channel{
			ID:           1,
			ControllerID: 2,
			ChannelID:    2,
			Name:         "Test Channel 3",
			Enable:       true,
			Notify:       true,
			Condition:    "foo > 80",
			Duration:     0,
			Debounce:     0,
			Backoff:      0}}
}

// doser

func createFakeDoserChannelEntities() []entity.Channel {
	channels := createFakeDoserChannels()
	entities := make([]entity.Channel, len(channels))
	for i, channel := range channels {
		_channel := mapper.NewChannelMapper().MapModelToEntity(channel)
		entities[i] = *_channel
	}
	return entities
}

func createFakeDoserChannels() []common.Channel {
	return []common.Channel{
		&model.Channel{
			ID:           1,
			ControllerID: 2,
			ChannelID:    0,
			Name:         "Doser Channel 1",
			Enable:       true,
			Notify:       true,
			Condition:    "",
			Duration:     0,
			Debounce:     0,
			Backoff:      0},
		&model.Channel{
			ID:           1,
			ControllerID: 2,
			ChannelID:    1,
			Name:         "Doser Channel 2",
			Enable:       true,
			Notify:       true,
			Condition:    "foo > 80",
			Duration:     0,
			Debounce:     0,
			Backoff:      0},
		&model.Channel{
			ID:           1,
			ControllerID: 2,
			ChannelID:    2,
			Name:         "Doser Channel 3",
			Enable:       true,
			Notify:       true,
			Condition:    "foo > 80",
			Duration:     0,
			Debounce:     0,
			Backoff:      0}}
}

func createFakeScheduleEntities() []entity.Schedule {
	return []entity.Schedule{}
}
