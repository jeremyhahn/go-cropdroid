package test

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/mock"
)

type MockControllerDAO struct {
	dao.ControllerDAO
	mock.Mock
}

func NewMockControllerDAO() *MockControllerDAO {
	return &MockControllerDAO{}
}

func (dao *MockControllerDAO) Create(controller config.ControllerConfig) error {
	args := dao.Called(controller)
	fmt.Println("Creating controller record")
	return args.Error(0)
}

func (dao *MockControllerDAO) Save(controller config.ControllerConfig) error {
	args := dao.Called(controller)
	fmt.Println("Saving controller record")
	return args.Error(0)
}

func (dao *MockControllerDAO) Update(controller config.ControllerConfig) error {
	args := dao.Called(controller)
	fmt.Println("Updating controller record")
	return args.Error(0)
}

func (dao *MockControllerDAO) Get(id int) (config.ControllerConfig, error) {
	args := dao.Called(id)
	fmt.Printf("Getting controllers for org id %d\n", id)
	return args.Get(0).(config.ControllerConfig), nil
}

func (dao *MockControllerDAO) GetByOrgId(orgId int) ([]config.Controller, error) {
	args := dao.Called(orgId)
	fmt.Printf("Getting controllers for org id %d\n", orgId)
	return args.Get(0).([]config.Controller), args.Error(1)
}

func (dao *MockControllerDAO) GetByOrgAndType(orgId int, controllerType string) ([]config.Controller, error) {
	args := dao.Called(orgId, controllerType)
	fmt.Printf("Getting controllers for org id %d\n", orgId)
	return args.Get(0).([]config.Controller), args.Error(1)
}
