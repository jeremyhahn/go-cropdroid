package test

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/mock"
)

type MockConditionDAO struct {
	dao.ConditionDAO
	mock.Mock
}

func NewMockConditionDAO() *MockConditionDAO {
	return &MockConditionDAO{}
}

func (dao *MockConditionDAO) Create(condition config.ConditionConfig) error {
	args := dao.Called(condition)
	fmt.Println("Creating condition record")
	return args.Error(0)
}

func (dao *MockConditionDAO) Save(condition config.ConditionConfig) error {
	args := dao.Called(condition)
	fmt.Println("Saving condition record")
	return args.Error(0)
}

func (dao *MockConditionDAO) Update(condition config.ConditionConfig) error {
	args := dao.Called(condition)
	fmt.Println("Updating condition record")
	return args.Error(0)
}

func (dao *MockConditionDAO) Get(id int) (config.ConditionConfig, error) {
	args := dao.Called(id)
	fmt.Printf("Getting condition for id=%d\n", id)
	return args.Get(0).(*config.Condition), nil
}

func (dao *MockConditionDAO) GetByChannelID(id int) ([]config.Condition, error) {
	args := dao.Called(id)
	fmt.Printf("Getting condition by channel id=%d\n", id)
	return args.Get(0).([]config.Condition), nil
}

func (dao *MockConditionDAO) GetByUserOrgAndControllerID(orgID, controllerID int) ([]config.Condition, error) {
	args := dao.Called(orgID, controllerID)
	fmt.Printf("Getting condition by controller id=%d and org id=%d\n", controllerID, orgID)
	return args.Get(0).([]config.Condition), nil
}
