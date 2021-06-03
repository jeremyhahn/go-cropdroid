package test

import (
	"fmt"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/stretchr/testify/mock"
)

type MockScheduleDAO struct {
	dao.ScheduleDAO
	mock.Mock
}

func NewMockScheduleDAO() *MockScheduleDAO {
	return &MockScheduleDAO{}
}

func (dao *MockScheduleDAO) Create(schedule config.ScheduleConfig) error {
	args := dao.Called(schedule)
	fmt.Println("Creating schedule record")
	return args.Error(0)
}

func (dao *MockScheduleDAO) Save(schedule config.ScheduleConfig) error {
	args := dao.Called(schedule)
	fmt.Println("Saving schedule record")
	return args.Error(0)
}

func (dao *MockScheduleDAO) Update(schedule config.ScheduleConfig) error {
	args := dao.Called(schedule)
	fmt.Println("Updating schedule record")
	return args.Error(0)
}

func (dao *MockScheduleDAO) Get(id int) (config.ScheduleConfig, error) {
	args := dao.Called(id)
	fmt.Printf("Getting schedule id=%d\n", id)
	return args.Get(0).(config.ScheduleConfig), nil
}

func (dao *MockScheduleDAO) GetByChannelID(id int) ([]config.Schedule, error) {
	args := dao.Called(id)
	fmt.Printf("Getting schedule by channel ID: id=%d\n", id)
	return args.Get(0).([]config.Schedule), nil
}

func (dao *MockScheduleDAO) GetByUserOrgAndChannelID(user common.UserAccount, channelID int) ([]config.Schedule, error) {
	args := dao.Called(user, channelID)
	fmt.Printf("Getting schedule by user org and channel ID: channelID=%d, user=%+v\n", channelID, user)
	return args.Get(0).([]config.Schedule), nil
}
