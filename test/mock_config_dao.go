package test

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

func (dao *MockConfigDAO) Create(controller config.ControllerConfigConfig) error {
	args := dao.Called(controller)
	fmt.Println("Creating controller record")
	return args.Error(0)
}

func (dao *MockConfigDAO) Save(controller config.ControllerConfigConfig) error {
	args := dao.Called(controller)
	fmt.Println("Saving controller record")
	return args.Error(0)
}

func (dao *MockConfigDAO) Update(controller config.ControllerConfigConfig) error {
	args := dao.Called(controller)
	fmt.Println("Updating controller record")
	return args.Error(0)
}

func (dao *MockConfigDAO) Get(controllerID, name string) (config.ControllerConfigConfig, error) {
	args := dao.Called(controllerID, name)
	fmt.Printf("Getting config for controllerID=%s and name=%s\n", controllerID, name)
	return args.Get(0).(*config.ControllerConfigItem), nil
}
