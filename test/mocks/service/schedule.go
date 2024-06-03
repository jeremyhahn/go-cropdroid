package service

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockScheduleService struct {
	session service.Session
	dao     dao.ScheduleDAO
	service.ScheduleService
	mock.Mock
}

func NewMockScheduleService(session service.Session, dao dao.ScheduleDAO) *MockScheduleService {
	return &MockScheduleService{
		session: session,
		dao:     dao}
}

func (service *MockScheduleService) Parse() *config.ScheduleStruct {
	args := service.Called()
	service.session.GetLogger().Debug("MockScheduleService.Parse")
	return args.Get(0).(*config.ScheduleStruct)
}

func (service *MockScheduleService) GetSchedule(channelID int) ([]*config.ScheduleStruct, error) {
	args := service.Called(channelID)
	service.session.GetLogger().Debugf("MockScheduleService.GetSchedule: channelID=%d", channelID)
	return args.Get(0).([]*config.ScheduleStruct), args.Error(1)
}

func (service *MockScheduleService) GetSchedules(user model.User, deviceID int) ([]*config.ScheduleStruct, error) {
	args := service.Called(user, deviceID)
	service.session.GetLogger().Debugf("MockScheduleService.GetSchedules: deviceID=%d, user=%+v", deviceID, user)
	return args.Get(0).([]*config.ScheduleStruct), args.Error(1)
}

func (service *MockScheduleService) Create(schedule *config.ScheduleStruct) (*config.ScheduleStruct, error) {
	args := service.Called(schedule)
	service.session.GetLogger().Debug("MockScheduleService.Parse")
	return args.Get(0).(*config.ScheduleStruct), args.Error(1)
}

func (service *MockScheduleService) Update(schedule *config.ScheduleStruct) error {
	args := service.Called(schedule)
	service.session.GetLogger().Debug("MockScheduleService.Update")
	return args.Error(1)
}

func (service *MockScheduleService) Delete(schedule *config.ScheduleStruct) error {
	args := service.Called(schedule)
	service.session.GetLogger().Debug("MockScheduleService.Delete")
	return args.Error(1)
}

func (service *MockScheduleService) IsScheduled(schedule *config.ScheduleStruct, duration int) bool {
	args := service.Called(schedule, duration)
	service.session.GetLogger().Debug("MockScheduleService.Delete duration=%d, schedule=%+v", duration, schedule)
	return args.Get(0).(bool)
}
