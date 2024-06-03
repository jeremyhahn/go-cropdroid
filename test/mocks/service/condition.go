package service

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
	"github.com/stretchr/testify/mock"
)

type MockConditionService struct {
	mock.Mock
	service.ConditionServicer
}

func NewMockConditionService() service.ConditionServicer {
	return &MockConditionService{}
}

func (service *MockConditionService) ListView(session service.Session, channelID uint64) ([]*viewmodel.Condition, error) {
	args := service.Called(session, channelID)
	models := []*viewmodel.Condition{
		{
			ID:         uint64(1),
			DeviceType: "Test Device",
			MetricID:   uint64(1),
			MetricName: "Test Metric",
			ChannelID:  uint64(1),
			Comparator: ">",
			Threshold:  5,
			Text:       "Test metric is greater than 5",
		},
	}
	return models, args.Error(1)
}

func (service *MockConditionService) Create(session service.Session, condition config.Condition) (config.Condition, error) {
	args := service.Called(session, condition)
	condition.SetID(uint64(1))
	return condition, args.Error(0)
}

func (service *MockConditionService) Update(session service.Session, condition config.Condition) error {
	args := service.Called(session, condition)
	return args.Error(0)
}

func (service *MockConditionService) Delete(session service.Session, condition config.Condition) error {
	args := service.Called(session, condition)
	return args.Error(0)
}
