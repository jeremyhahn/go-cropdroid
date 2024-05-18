package mocks

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/stretchr/testify/mock"
)

type MockDeviceDAO struct {
	dao.DeviceDAO
	mock.Mock
}

func NewMockDeviceDAO() *MockDeviceDAO {
	return &MockDeviceDAO{}
}

func (dao *MockDeviceDAO) Create(device config.Device) error {
	args := dao.Called(device)
	fmt.Println("Creating device record")
	return args.Error(0)
}

func (dao *MockDeviceDAO) Save(device config.Device) error {
	args := dao.Called(device)
	fmt.Println("Saving device record")
	return args.Error(0)
}

func (dao *MockDeviceDAO) Update(device config.Device) error {
	args := dao.Called(device)
	fmt.Println("Updating device record")
	return args.Error(0)
}

func (dao *MockDeviceDAO) Get(id int) (config.Device, error) {
	args := dao.Called(id)
	fmt.Printf("Getting devices for org id %d\n", id)
	return args.Get(0).(config.Device), nil
}

func (dao *MockDeviceDAO) GetByOrgId(orgId int) ([]config.Device, error) {
	args := dao.Called(orgId)
	fmt.Printf("Getting devices for org id %d\n", orgId)
	return args.Get(0).([]config.Device), args.Error(1)
}

func (dao *MockDeviceDAO) GetByOrgAndType(orgId int, deviceType string) ([]config.Device, error) {
	args := dao.Called(orgId, deviceType)
	fmt.Printf("Getting devices for org id %d\n", orgId)
	return args.Get(0).([]config.Device), args.Error(1)
}
