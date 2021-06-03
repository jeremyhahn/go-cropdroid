// +build broken

package test

import (
	"testing"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/service"
)

func TestManageMetrics(t *testing.T) {

	//controllerService := newControllerService()
	//controllerService.Manage()
}

func newControllerService() common.ControllerService {
	app, scope := NewUnitTestContext()
	scheduleDAO := NewMockScheduleDAO()
	dynamicDAO := NewMockDynamicDAO()
	scheduleMapper := mapper.NewScheduleMapper()
	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	controllerMapper := mapper.NewControllerMapper(metricMapper, channelMapper)
	controller := NewMockController()
	notificationService := NewMockNotificationService(scope, NewMockMailer(scope))
	eventLogService := NewMockEventLogService(scope, nil, "test")
	conditionService := NewMockConditionService()
	configService := &MockConfigService{}
	scheduleService := service.NewScheduleService(app, scheduleDAO, scheduleMapper, configService)

	service, err := service.NewMicroControllerService(app, scope, dynamicDAO, controller, controllerMapper, eventLogService, notificationService, conditionService, scheduleService)
	if err != nil {
		scope.GetLogger().Fatal(err)
	}
	return service
}
