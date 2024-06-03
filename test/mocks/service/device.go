package service

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
	"github.com/stretchr/testify/mock"
)

type MockDeviceService struct {
	deviceView viewmodel.DeviceView
	mock.Mock
	service.DeviceServicer
}

func NewMockDeviceService() service.DeviceServicer {
	return &MockDeviceService{}
}

func (service *MockDeviceService) DeviceType() string {
	service.Called()
	return "test-device"
}

func (service *MockDeviceService) View() (viewmodel.DeviceView, error) {
	service.Called()
	deviceView := service.ExpectedView()
	return deviceView, nil
}

// Creates a singleton device view to preserve the timestamp for assertions
func (service *MockDeviceService) ExpectedView() viewmodel.DeviceView {
	if service.deviceView != nil {
		return service.deviceView
	}
	now := time.Now()
	service.deviceView = &viewmodel.DeviceViewModel{
		Metrics: []model.Metric{
			&model.MetricStruct{
				ID:        uint64(1),
				DeviceID:  uint64(1),
				DataType:  1,
				Enable:    true,
				Notify:    true,
				Name:      "Test Metric 1",
				Key:       "tm1",
				Unit:      "%",
				AlarmLow:  5,
				AlarmHigh: 10,
				Value:     8,
				Timestamp: &now,
			},
		},
		Channels: []model.Channel{
			&model.ChannelStruct{
				ID:       uint64(1),
				DeviceID: uint64(1),
				Enable:   true,
				Notify:   true,
				Name:     "Test Channel 1",
				Value:    5},
		},
		Timestamp: time.Now(),
	}
	return service.deviceView
}
