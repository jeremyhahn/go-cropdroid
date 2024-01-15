package mocks

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockConditionService struct {
	service.ConditionService
	mock.Mock
}

func NewMockConditionService() *MockConditionService {
	return &MockConditionService{}
}

func (service *MockConditionService) IsTrue(condition config.Condition, value float64) (bool, error) {
	args := service.Called(condition, value)
	fmt.Printf("Evaluating condition: %+v, value=%.2f", condition, value)
	return args.Get(0).(bool), args.Error(0)
}
