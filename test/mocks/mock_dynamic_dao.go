// +build broken

package mocks

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/stretchr/testify/mock"
)

type MockDynamicDAO struct {
	datastore.MetricDataDAO
	mock.Mock
}

func NewMockDynamicDAO() *MockDynamicDAO {
	return &MockDynamicDAO{}
}

func (dao *MockDynamicDAO) CreateTable(tableName string, rows map[string]float64) error {
	args := dao.Called(tableName, rows)
	fmt.Printf("Creating table for dynamic device: device=%s, %+v\n", tableName, rows)
	return args.Error(0)
}

func (dao *MockDynamicDAO) Save(tableName string, rows map[string]float64) error {
	args := dao.Called(tableName, rows)
	fmt.Printf("Saving dynamic device state; device=%s, rows=%+v\n", tableName, rows)
	return args.Error(0)
}

func (dao *MockDynamicDAO) GetLast30Days(tableName, metric string) ([]float64, error) {
	args := dao.Called(tableName, metric)
	fmt.Printf("Creating dynamic device state; device=%s, metric=%s\n", tableName, metric)
	return args.Get(0).([]float64), args.Error(0)
}
