// +build ignore

package test

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
)

func TestManageMetrics(t *testing.T) {

	//deviceService := newDeviceService()
	//deviceService.Manage()
}

func newDeviceService() common.DeviceService {
	app, scope := NewUnitTestContext()
	scheduleDAO := NewMockScheduleDAO()
	dynamicDAO := NewMockDynamicDAO()
	scheduleMapper := mapper.NewScheduleMapper()
	metricMapper := mapper.NewMetricMapper()
	channelMapper := mapper.NewChannelMapper()
	deviceMapper := mapper.NewDeviceMapper(metricMapper, channelMapper)
	device := NewMockDevice()
	notificationService := NewMockNotificationService(scope, NewMockMailer(scope))
	eventLogService := NewMockEventLogService(scope, nil, "test")
	conditionService := NewMockConditionService()
	configService := &MockConfigService{}
	scheduleService := service.NewScheduleService(app, scheduleDAO, scheduleMapper, configService)

	service, err := service.NewMicroDeviceService(app, scope, dynamicDAO, device, deviceMapper, eventLogService, notificationService, conditionService, scheduleService)
	if err != nil {
		scope.GetLogger().Fatal(err)
	}
	return service
}
