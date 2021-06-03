// +build broken

package test

import (
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockScheduleService struct {
	scope  common.Scope
	dao    dao.ScheduleDAO
	mapper mapper.ScheduleMapper
	service.ScheduleService
	mock.Mock
}

func NewMockScheduleService(scope common.Scope, dao dao.ScheduleDAO, mapper mapper.ScheduleMapper) *MockScheduleService {
	return &MockScheduleService{
		scope:  scope,
		dao:    dao,
		mapper: mapper}
}

func (service *MockScheduleService) Parse() config.ScheduleConfig {
	args := service.Called()
	service.scope.GetLogger().Debug("MockScheduleService.Parse")
	return args.Get(0).(config.ScheduleConfig)
}

func (service *MockScheduleService) GetSchedule(channelID int) ([]config.ScheduleConfig, error) {
	args := service.Called(channelID)
	service.scope.GetLogger().Debugf("MockScheduleService.GetSchedule: channelID=%d", channelID)
	return args.Get(0).([]config.ScheduleConfig), args.Error(1)
}

func (service *MockScheduleService) GetSchedules(user common.UserAccount, controllerID int) ([]config.ScheduleConfig, error) {
	args := service.Called(user, controllerID)
	service.scope.GetLogger().Debugf("MockScheduleService.GetSchedules: controllerID=%d, user=%+v", controllerID, user)
	return args.Get(0).([]config.ScheduleConfig), args.Error(1)
}

func (service *MockScheduleService) Create(schedule config.ScheduleConfig) (config.ScheduleConfig, error) {
	args := service.Called(schedule)
	service.scope.GetLogger().Debug("MockScheduleService.Parse")
	return args.Get(0).(config.ScheduleConfig), args.Error(1)
}

func (service *MockScheduleService) Update(schedule config.ScheduleConfig) error {
	args := service.Called(schedule)
	service.scope.GetLogger().Debug("MockScheduleService.Update")
	return args.Error(1)
}

func (service *MockScheduleService) Delete(schedule config.ScheduleConfig) error {
	args := service.Called(schedule)
	service.scope.GetLogger().Debug("MockScheduleService.Delete")
	return args.Error(1)
}

func (service *MockScheduleService) IsScheduled(schedule config.ScheduleConfig, duration int) bool {
	args := service.Called(schedule, duration)
	service.scope.GetLogger().Debug("MockScheduleService.Delete duration=%d, schedule=%+v", duration, schedule)
	return args.Get(0).(bool)
}
