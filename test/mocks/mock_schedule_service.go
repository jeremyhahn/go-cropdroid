package mocks

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockScheduleService struct {
	session service.Session
	dao     dao.ScheduleDAO
	mapper  mapper.ScheduleMapper
	service.ScheduleService
	mock.Mock
}

func NewMockScheduleService(session service.Session, dao dao.ScheduleDAO, mapper mapper.ScheduleMapper) *MockScheduleService {
	return &MockScheduleService{
		session: session,
		dao:     dao,
		mapper:  mapper}
}

func (service *MockScheduleService) Parse() config.Schedule {
	args := service.Called()
	service.session.GetLogger().Debug("MockScheduleService.Parse")
	return args.Get(0).(config.Schedule)
}

func (service *MockScheduleService) GetSchedule(channelID int) ([]config.Schedule, error) {
	args := service.Called(channelID)
	service.session.GetLogger().Debugf("MockScheduleService.GetSchedule: channelID=%d", channelID)
	return args.Get(0).([]config.Schedule), args.Error(1)
}

func (service *MockScheduleService) GetSchedules(user common.UserAccount, deviceID int) ([]config.Schedule, error) {
	args := service.Called(user, deviceID)
	service.session.GetLogger().Debugf("MockScheduleService.GetSchedules: deviceID=%d, user=%+v", deviceID, user)
	return args.Get(0).([]config.Schedule), args.Error(1)
}

func (service *MockScheduleService) Create(schedule config.Schedule) (config.Schedule, error) {
	args := service.Called(schedule)
	service.session.GetLogger().Debug("MockScheduleService.Parse")
	return args.Get(0).(config.Schedule), args.Error(1)
}

func (service *MockScheduleService) Update(schedule config.Schedule) error {
	args := service.Called(schedule)
	service.session.GetLogger().Debug("MockScheduleService.Update")
	return args.Error(1)
}

func (service *MockScheduleService) Delete(schedule config.Schedule) error {
	args := service.Called(schedule)
	service.session.GetLogger().Debug("MockScheduleService.Delete")
	return args.Error(1)
}

func (service *MockScheduleService) IsScheduled(schedule config.Schedule, duration int) bool {
	args := service.Called(schedule, duration)
	service.session.GetLogger().Debug("MockScheduleService.Delete duration=%d, schedule=%+v", duration, schedule)
	return args.Get(0).(bool)
}
