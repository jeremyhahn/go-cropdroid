// +build ignore

package test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/mock"
)

func TestChannelGreaterThanConditionalActivatesSwitch(t *testing.T) {

	scope, client, notificationService, _, _, _, microdeviceService := createTestMicrodeviceService()

	testFarmID := 0
	testChannelID := 0
	deviceType := "testdevice"
	farmConfig, _ := createTestFarm(testFarmID, deviceType)

	scope.SetConfig(farmConfig)

	client.On("GetType").Return(deviceType, nil)
	client.On("Switch", testChannelID, 1).Return(&common.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	channels := farmConfig.GetDevices()[0].GetChannels()
	microdeviceService.ManageChannels(channels)

	client.AssertCalled(t, "Switch", testChannelID, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestChannelLessThanConditionalDeactivatesSwitch(t *testing.T) {

	scope, client, notificationService, _, _, _, microdeviceService := createTestMicrodeviceService()

	testChannelID := 0
	deviceType := "testdevice"
	fakeValue := 54.0

	metricMap := make(map[string]float64, 0)
	metricMap["humidity0"] = fakeValue
	deviceState := common.CreateDeviceStateMap(metricMap, []int{1})

	scope.GetStateMap().SetDevice(deviceType, deviceState)

	metrics := []config.MetricConfig{
		&config.Metric{
			ID:  1,
			Key: "humidity0"}}

	scope.GetConfig().GetDevices()[0].SetMetrics(metrics)

	client.On("GetType").Return(deviceType, nil)
	client.On("Switch", testChannelID, 0).Return(&common.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	channels := scope.GetConfig().GetDevices()[0].GetChannels()
	microdeviceService.ManageChannels(channels)

	client.AssertCalled(t, "Switch", testChannelID, 0)
	client.AssertNumberOfCalls(t, "Switch", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestChannelLessThanConditionalDeactivatesSwitchOnlyIfAlreadyOn(t *testing.T) {

	scope, client, notificationService, _, _, _, microdeviceService := createTestMicrodeviceService()

	deviceType := "testdevice"
	fakeValue := 54.0

	metricMap := make(map[string]float64, 0)
	metricMap["humidity0"] = fakeValue
	deviceState := common.CreateDeviceStateMap(metricMap, []int{0})

	scope.GetStateMap().SetDevice(deviceType, deviceState)

	metrics := []config.MetricConfig{
		&config.Metric{
			ID:  1,
			Key: "humidity0"}}

	scope.GetConfig().GetDevices()[0].SetMetrics(metrics)
	channels := scope.GetConfig().GetDevices()[0].GetChannels()

	client.On("GetType").Return(deviceType, nil)

	microdeviceService.ManageChannels(channels)

	client.AssertNumberOfCalls(t, "Switch", 0)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 0)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestChannelLessThanConditionalActivatesSwitchOnlyIfAlreadyOff(t *testing.T) {

	scope, client, notificationService, _, _, _, microdeviceService := createTestMicrodeviceService()

	deviceType := "testdevice"
	fakeValue := 56.0

	metricMap := make(map[string]float64, 0)
	metricMap["humidity0"] = fakeValue
	deviceState := common.CreateDeviceStateMap(metricMap, []int{1})

	scope.GetStateMap().SetDevice(deviceType, deviceState)

	metrics := []config.MetricConfig{
		&config.Metric{
			ID:  1,
			Key: "humidity0"}}

	scope.GetConfig().GetDevices()[0].SetMetrics(metrics)
	channels := scope.GetConfig().GetDevices()[0].GetChannels()

	client.On("GetType").Return(deviceType, nil)

	microdeviceService.ManageChannels(channels)

	client.AssertNumberOfCalls(t, "Switch", 0)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 0)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestPhAlgorithmActivatesSwitch(t *testing.T) {

	scope, client, notificationService, _, _, _, microdeviceService := createTestMicrodeviceService()

	testChannelID := 0
	deviceType := "testdevice"
	fakeValue := 6.1
	expectedAutoDoseSize := 3

	metricMap := make(map[string]float64, 0)
	metricMap["ph"] = fakeValue

	deviceState := common.CreateDeviceStateMap(metricMap, []int{0})
	scope.GetStateMap().SetDevice(deviceType, deviceState)

	deviceConfigMap := make(map[string]string, 0)
	deviceConfigMap[fmt.Sprintf("%s.notify", deviceType)] = "true"
	deviceConfigMap[fmt.Sprintf("%s.gallons", deviceType)] = "60"

	metrics := []config.MetricConfig{
		&config.Metric{
			ID:  1,
			Key: "ph"}}

	condition := &config.Condition{
		ID:         1,
		ChannelID:  1,
		MetricID:   1,
		Comparator: ">",
		Threshold:  6.0}

	channels := []config.ChannelConfig{
		&config.Channel{
			ID:           1,
			DeviceID: 1,
			ChannelID:    testChannelID,
			Name:         "channel",
			Enable:       true,
			Notify:       true,
			Conditions:   []config.ConditionConfig{condition},
			AlgorithmID:  common.ALGORITHM_PH_ID}}

	device := scope.GetConfig().GetDevices()[0]
	device.SetConfigs(deviceConfigMap)
	device.SetMetrics(metrics)
	device.SetChannels(channels)

	client.On("GetType").Return(deviceType, nil)
	client.On("TimerSwitch", testChannelID, expectedAutoDoseSize).Return(&common.ChannelTimerEvent{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	microdeviceService.ManageChannels(channels)

	client.AssertCalled(t, "TimerSwitch", testChannelID, expectedAutoDoseSize)
	client.AssertNumberOfCalls(t, "TimerSwitch", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func createTestMicrodeviceService() (common.Scope, *MockDevice, *MockNotificationService, *MockChannelService,
	service.ScheduleService, service.ConditionService, common.DeviceService) {

	app, scope := NewUnitTestContext()
	dao := NewMockDynamicDAO()
	conditionDAO := NewMockConditionDAO()
	scheduleDAO := NewMockScheduleDAO()
	client := NewMockDevice()
	mailer := NewMockMailer(scope)
	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	scheduleMapper := mapper.NewScheduleMapper()
	conditionMapper := mapper.NewConditionMapper()
	deviceMapper := mapper.NewDeviceMapper(metricMapper, channelMapper)
	notificationService := NewMockNotificationService(scope, mailer)
	eventLogService := NewMockEventLogService(scope, nil, "test")
	channelService := NewMockChannelService()
	configService := NewMockConfigService()
	conditionService := service.NewConditionService(scope, conditionDAO, conditionMapper, configService)
	scheduleService := service.NewScheduleService(app, scheduleDAO, scheduleMapper, configService)
	microdeviceService, err := service.NewMicroDeviceService(app, scope, dao, client, deviceMapper,
		eventLogService, notificationService, conditionService, scheduleService)

	if err != nil {
		log.Fatal("[MicroDeviceTest.createTestMicrodeviceService] Error: ", err)
	}

	return scope, client, notificationService, channelService, scheduleService, conditionService, microdeviceService
}

func createTestMicrodeviceServiceWithSchedule(scope common.Scope, now time.Time) (*MockDevice, *MockNotificationService,
	*MockChannelService, service.ScheduleService, service.ConditionService, common.DeviceService) {

	location, _ := time.LoadLocation("America/New_York")
	_app := &app.App{
		Location: location,
	}
	dao := NewMockDynamicDAO()
	conditionDAO := NewMockConditionDAO()
	scheduleDAO := NewMockScheduleDAO()
	client := NewMockDevice()
	mailer := NewMockMailer(scope)
	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	scheduleMapper := mapper.NewScheduleMapper()
	conditionMapper := mapper.NewConditionMapper()
	deviceMapper := mapper.NewDeviceMapper(metricMapper, channelMapper)
	notificationService := NewMockNotificationService(scope, mailer)
	eventLogService := NewMockEventLogService(scope, nil, "test")
	channelService := NewMockChannelService()
	configService := NewMockConfigService()
	conditionService := service.NewConditionService(scope, conditionDAO, conditionMapper, configService)
	scheduleService, _ := service.CreateScheduleService(_app, scheduleDAO, scheduleMapper, now, configService)
	microdeviceService, err := service.NewMicroDeviceService(_app, scope, dao, client, deviceMapper, eventLogService,
		notificationService, conditionService, scheduleService)

	if err != nil {
		log.Fatal("[MicroDeviceTest.createTestMicrodeviceServiceWithSchedule] Error: ", err)
	}

	return client, notificationService, channelService, scheduleService, conditionService, microdeviceService
}
