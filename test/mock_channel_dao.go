package test

import (
	"fmt"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/stretchr/testify/mock"
)

type MockChannelDAO struct {
	dao.ChannelDAO
	mock.Mock
}

func NewMockChannelDAO() *MockChannelDAO {
	return &MockChannelDAO{}
}

func (dao *MockChannelDAO) Create(channel config.ChannelConfig) error {
	args := dao.Called(channel)
	fmt.Printf("Creating channel record")
	return args.Error(0)
}

func (dao *MockChannelDAO) Save(channel config.ChannelConfig) error {
	args := dao.Called(channel)
	fmt.Printf("Saving channel record")
	return args.Error(0)
}

func (dao *MockChannelDAO) Update(channel config.ChannelConfig) error {
	args := dao.Called(channel)
	fmt.Printf("Updating channel record")
	return args.Error(0)
}

func (dao *MockChannelDAO) Get(channelID int) (config.ChannelConfig, error) {
	args := dao.Called(channelID)
	fmt.Printf("Getting channel channelID=%d", channelID)
	return args.Get(0).(config.ChannelConfig), nil
}

func (dao *MockChannelDAO) GetByControllerNameAndID(controllerID int, name string) (config.ChannelConfig, error) {
	args := dao.Called(controllerID, name)
	fmt.Printf("Getting channel by controller name and ID: name=%s, controllerID=%d", name, controllerID)
	return args.Get(0).(config.ChannelConfig), nil
}

func (dao *MockChannelDAO) GetByControllerID(controllerID int) ([]config.Channel, error) {
	args := dao.Called(controllerID)
	fmt.Printf("Getting channel by controller ID: controllerID=%d", controllerID)
	return args.Get(0).([]config.Channel), nil
}

func (dao *MockChannelDAO) GetByUserOrgAndControllerID(user common.UserAccount, controllerID int) ([]config.Channel, error) {
	args := dao.Called(user, controllerID)
	fmt.Printf("Getting channel by user org and controller ID: controllerID=%d, user=%+v", controllerID, user)
	return args.Get(0).([]config.Channel), nil
}
