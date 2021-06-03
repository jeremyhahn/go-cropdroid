package test

import (
	"fmt"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/stretchr/testify/mock"
)

type MockMetricDAO struct {
	dao.MetricDAO
	mock.Mock
}

func NewMockMetricDAO() *MockMetricDAO {
	return &MockMetricDAO{}
}

func (dao *MockMetricDAO) Create(metric config.MetricConfig) error {
	args := dao.Called(metric)
	fmt.Println("Creating metric record")
	return args.Error(0)
}

func (dao *MockMetricDAO) Save(metric config.MetricConfig) error {
	args := dao.Called(metric)
	fmt.Println("Saving metric record")
	return args.Error(0)
}

func (dao *MockMetricDAO) Update(metric config.MetricConfig) error {
	args := dao.Called(metric)
	fmt.Println("Updating metric record")
	return args.Error(0)
}

func (dao *MockMetricDAO) Get(metricID int) (config.MetricConfig, error) {
	args := dao.Called(metricID)
	fmt.Printf("Getting metric metricID=%d]n", metricID)
	return args.Get(0).(config.MetricConfig), nil
}

func (dao *MockMetricDAO) GetByControllerID(controllerID int) ([]config.Metric, error) {
	args := dao.Called(controllerID)
	fmt.Printf("Getting metric by controller ID: controllerID=%d\n", controllerID)
	return args.Get(0).([]config.Metric), nil
}

func (dao *MockMetricDAO) GetByUserOrgAndControllerID(user common.UserAccount, controllerID int) ([]config.Metric, error) {
	args := dao.Called(user, controllerID)
	fmt.Printf("Getting metric by user org and controller ID: controllerID=%d, user=%+v\n", controllerID, user)
	return args.Get(0).([]config.Metric), nil
}
