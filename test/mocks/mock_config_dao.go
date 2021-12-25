//go:build ignore
// +build ignore

package mocks

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/mock"
)

type MockConfigDAO struct {
	dao.ConfigDAO
	mock.Mock
}

func NewMockConfigDAO() *MockConfigDAO {
	return &MockConfigDAO{}
}

func (dao *MockConfigDAO) Create(device config.DeviceConfigConfig) error {
	args := dao.Called(device)
	fmt.Println("Creating device record")
	return args.Error(0)
}

func (dao *MockConfigDAO) Save(device config.DeviceConfigConfig) error {
	args := dao.Called(device)
	fmt.Println("Saving device record")
	return args.Error(0)
}

func (dao *MockConfigDAO) Update(device config.DeviceConfigConfig) error {
	args := dao.Called(device)
	fmt.Println("Updating device record")
	return args.Error(0)
}

func (dao *MockConfigDAO) Get(deviceID, name string) (config.DeviceConfigConfig, error) {
	args := dao.Called(deviceID, name)
	fmt.Printf("Getting config for deviceID=%s and name=%s\n", deviceID, name)
	return args.Get(0).(*config.DeviceConfigItem), nil
}
