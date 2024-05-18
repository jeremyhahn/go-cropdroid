package mocks

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/stretchr/testify/mock"
)

type MockChannelDAO struct {
	dao.ChannelDAO
	mock.Mock
}

func NewMockChannelDAO() *MockChannelDAO {
	return &MockChannelDAO{}
}

func (dao *MockChannelDAO) Create(channel config.Channel) error {
	args := dao.Called(channel)
	fmt.Printf("Creating channel record")
	return args.Error(0)
}

func (dao *MockChannelDAO) Save(channel config.Channel) error {
	args := dao.Called(channel)
	fmt.Printf("Saving channel record")
	return args.Error(0)
}

func (dao *MockChannelDAO) Update(channel config.Channel) error {
	args := dao.Called(channel)
	fmt.Printf("Updating channel record")
	return args.Error(0)
}

func (dao *MockChannelDAO) Get(channelID int) (config.Channel, error) {
	args := dao.Called(channelID)
	fmt.Printf("Getting channel channelID=%d", channelID)
	return args.Get(0).(config.Channel), nil
}

func (dao *MockChannelDAO) GetByDeviceNameAndID(deviceID int, name string) (config.Channel, error) {
	args := dao.Called(deviceID, name)
	fmt.Printf("Getting channel by device name and ID: name=%s, deviceID=%d", name, deviceID)
	return args.Get(0).(config.Channel), nil
}

func (dao *MockChannelDAO) GetByDeviceID(deviceID int) ([]config.Channel, error) {
	args := dao.Called(deviceID)
	fmt.Printf("Getting channel by device ID: deviceID=%d", deviceID)
	return args.Get(0).([]config.Channel), nil
}

func (dao *MockChannelDAO) GetByUserOrgAndDeviceID(user common.UserAccount, deviceID int) ([]config.Channel, error) {
	args := dao.Called(user, deviceID)
	fmt.Printf("Getting channel by user org and device ID: deviceID=%d, user=%+v", deviceID, user)
	return args.Get(0).([]config.Channel), nil
}
