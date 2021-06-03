// +build broken

package test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/mock"
)

func TestChannelGreaterThanConditionalActivatesSwitch(t *testing.T) {

	scope, client, notificationService, _, _, _, microcontrollerService := createTestMicrocontrollerService()

	testFarmID := 0
	testChannelID := 0
	controllerType := "testcontroller"
	farmConfig, _ := createTestFarm(testFarmID, controllerType)

	scope.SetConfig(farmConfig)

	client.On("GetType").Return(controllerType, nil)
	client.On("Switch", testChannelID, 1).Return(&common.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	channels := farmConfig.GetControllers()[0].GetChannels()
	microcontrollerService.ManageChannels(channels)

	client.AssertCalled(t, "Switch", testChannelID, 1)
	client.AssertNumberOfCalls(t, "Switch", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestChannelLessThanConditionalDeactivatesSwitch(t *testing.T) {

	scope, client, notificationService, _, _, _, microcontrollerService := createTestMicrocontrollerService()

	testChannelID := 0
	controllerType := "testcontroller"
	fakeValue := 54.0

	metricMap := make(map[string]float64, 0)
	metricMap["humidity0"] = fakeValue
	controllerState := common.CreateControllerStateMap(metricMap, []int{1})

	scope.GetStateMap().SetController(controllerType, controllerState)

	metrics := []config.MetricConfig{
		&config.Metric{
			ID:  1,
			Key: "humidity0"}}

	scope.GetConfig().GetControllers()[0].SetMetrics(metrics)

	client.On("GetType").Return(controllerType, nil)
	client.On("Switch", testChannelID, 0).Return(&common.Switch{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	channels := scope.GetConfig().GetControllers()[0].GetChannels()
	microcontrollerService.ManageChannels(channels)

	client.AssertCalled(t, "Switch", testChannelID, 0)
	client.AssertNumberOfCalls(t, "Switch", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestChannelLessThanConditionalDeactivatesSwitchOnlyIfAlreadyOn(t *testing.T) {

	scope, client, notificationService, _, _, _, microcontrollerService := createTestMicrocontrollerService()

	controllerType := "testcontroller"
	fakeValue := 54.0

	metricMap := make(map[string]float64, 0)
	metricMap["humidity0"] = fakeValue
	controllerState := common.CreateControllerStateMap(metricMap, []int{0})

	scope.GetStateMap().SetController(controllerType, controllerState)

	metrics := []config.MetricConfig{
		&config.Metric{
			ID:  1,
			Key: "humidity0"}}

	scope.GetConfig().GetControllers()[0].SetMetrics(metrics)
	channels := scope.GetConfig().GetControllers()[0].GetChannels()

	client.On("GetType").Return(controllerType, nil)

	microcontrollerService.ManageChannels(channels)

	client.AssertNumberOfCalls(t, "Switch", 0)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 0)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestChannelLessThanConditionalActivatesSwitchOnlyIfAlreadyOff(t *testing.T) {

	scope, client, notificationService, _, _, _, microcontrollerService := createTestMicrocontrollerService()

	controllerType := "testcontroller"
	fakeValue := 56.0

	metricMap := make(map[string]float64, 0)
	metricMap["humidity0"] = fakeValue
	controllerState := common.CreateControllerStateMap(metricMap, []int{1})

	scope.GetStateMap().SetController(controllerType, controllerState)

	metrics := []config.MetricConfig{
		&config.Metric{
			ID:  1,
			Key: "humidity0"}}

	scope.GetConfig().GetControllers()[0].SetMetrics(metrics)
	channels := scope.GetConfig().GetControllers()[0].GetChannels()

	client.On("GetType").Return(controllerType, nil)

	microcontrollerService.ManageChannels(channels)

	client.AssertNumberOfCalls(t, "Switch", 0)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 0)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func TestPhAlgorithmActivatesSwitch(t *testing.T) {

	scope, client, notificationService, _, _, _, microcontrollerService := createTestMicrocontrollerService()

	testChannelID := 0
	controllerType := "testcontroller"
	fakeValue := 6.1
	expectedAutoDoseSize := 3

	metricMap := make(map[string]float64, 0)
	metricMap["ph"] = fakeValue

	controllerState := common.CreateControllerStateMap(metricMap, []int{0})
	scope.GetStateMap().SetController(controllerType, controllerState)

	controllerConfigMap := make(map[string]string, 0)
	controllerConfigMap[fmt.Sprintf("%s.notify", controllerType)] = "true"
	controllerConfigMap[fmt.Sprintf("%s.gallons", controllerType)] = "60"

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
			ControllerID: 1,
			ChannelID:    testChannelID,
			Name:         "channel",
			Enable:       true,
			Notify:       true,
			Conditions:   []config.ConditionConfig{condition},
			AlgorithmID:  common.ALGORITHM_PH_ID}}

	controller := scope.GetConfig().GetControllers()[0]
	controller.SetConfigs(controllerConfigMap)
	controller.SetMetrics(metrics)
	controller.SetChannels(channels)

	client.On("GetType").Return(controllerType, nil)
	client.On("TimerSwitch", testChannelID, expectedAutoDoseSize).Return(&common.ChannelTimerEvent{}, nil)
	notificationService.On("Enqueue", mock.Anything).Return(nil)

	microcontrollerService.ManageChannels(channels)

	client.AssertCalled(t, "TimerSwitch", testChannelID, expectedAutoDoseSize)
	client.AssertNumberOfCalls(t, "TimerSwitch", 1)
	notificationService.AssertNumberOfCalls(t, "Enqueue", 1)

	client.AssertExpectations(t)
	notificationService.AssertExpectations(t)
}

func createTestMicrocontrollerService() (common.Scope, *MockController, *MockNotificationService, *MockChannelService,
	service.ScheduleService, service.ConditionService, common.ControllerService) {

	app, scope := NewUnitTestContext()
	dao := NewMockDynamicDAO()
	conditionDAO := NewMockConditionDAO()
	scheduleDAO := NewMockScheduleDAO()
	client := NewMockController()
	mailer := NewMockMailer(scope)
	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	scheduleMapper := mapper.NewScheduleMapper()
	conditionMapper := mapper.NewConditionMapper()
	controllerMapper := mapper.NewControllerMapper(metricMapper, channelMapper)
	notificationService := NewMockNotificationService(scope, mailer)
	eventLogService := NewMockEventLogService(scope, nil, "test")
	channelService := NewMockChannelService()
	configService := NewMockConfigService()
	conditionService := service.NewConditionService(scope, conditionDAO, conditionMapper, configService)
	scheduleService := service.NewScheduleService(app, scheduleDAO, scheduleMapper, configService)
	microcontrollerService, err := service.NewMicroControllerService(app, scope, dao, client, controllerMapper,
		eventLogService, notificationService, conditionService, scheduleService)

	if err != nil {
		log.Fatal("[MicroControllerTest.createTestMicrocontrollerService] Error: ", err)
	}

	return scope, client, notificationService, channelService, scheduleService, conditionService, microcontrollerService
}

func createTestMicrocontrollerServiceWithSchedule(scope common.Scope, now time.Time) (*MockController, *MockNotificationService,
	*MockChannelService, service.ScheduleService, service.ConditionService, common.ControllerService) {

	location, _ := time.LoadLocation("America/New_York")
	_app := &app.App{
		Location: location,
	}
	dao := NewMockDynamicDAO()
	conditionDAO := NewMockConditionDAO()
	scheduleDAO := NewMockScheduleDAO()
	client := NewMockController()
	mailer := NewMockMailer(scope)
	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	scheduleMapper := mapper.NewScheduleMapper()
	conditionMapper := mapper.NewConditionMapper()
	controllerMapper := mapper.NewControllerMapper(metricMapper, channelMapper)
	notificationService := NewMockNotificationService(scope, mailer)
	eventLogService := NewMockEventLogService(scope, nil, "test")
	channelService := NewMockChannelService()
	configService := NewMockConfigService()
	conditionService := service.NewConditionService(scope, conditionDAO, conditionMapper, configService)
	scheduleService, _ := service.CreateScheduleService(_app, scheduleDAO, scheduleMapper, now, configService)
	microcontrollerService, err := service.NewMicroControllerService(_app, scope, dao, client, controllerMapper, eventLogService,
		notificationService, conditionService, scheduleService)

	if err != nil {
		log.Fatal("[MicroControllerTest.createTestMicrocontrollerServiceWithSchedule] Error: ", err)
	}

	return client, notificationService, channelService, scheduleService, conditionService, microcontrollerService
}
