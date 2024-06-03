package dao

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/stretchr/testify/mock"
)

type MockMetricDAO struct {
	dao.MetricDAO
	mock.Mock
}

func NewMockMetricDAO() *MockMetricDAO {
	return &MockMetricDAO{}
}

func (dao *MockMetricDAO) Create(metric *config.MetricStruct) error {
	args := dao.Called(metric)
	fmt.Println("Creating metric record")
	return args.Error(0)
}

func (dao *MockMetricDAO) Save(metric *config.MetricStruct) error {
	args := dao.Called(metric)
	fmt.Println("Saving metric record")
	return args.Error(0)
}

func (dao *MockMetricDAO) Update(metric *config.MetricStruct) error {
	args := dao.Called(metric)
	fmt.Println("Updating metric record")
	return args.Error(0)
}

func (dao *MockMetricDAO) Get(metricID int) (*config.MetricStruct, error) {
	args := dao.Called(metricID)
	fmt.Printf("Getting metric metricID=%d]n", metricID)
	return args.Get(0).(*config.MetricStruct), nil
}

func (dao *MockMetricDAO) GetByDeviceID(deviceID int) ([]*config.MetricStruct, error) {
	args := dao.Called(deviceID)
	fmt.Printf("Getting metric by device ID: deviceID=%d\n", deviceID)
	return args.Get(0).([]*config.MetricStruct), nil
}

func (dao *MockMetricDAO) GetByUserOrgAndDeviceID(user model.User, deviceID int) ([]*config.MetricStruct, error) {
	args := dao.Called(user, deviceID)
	fmt.Printf("Getting metric by user org and device ID: deviceID=%d, user=%+v\n", deviceID, user)
	return args.Get(0).([]*config.MetricStruct), nil
}
