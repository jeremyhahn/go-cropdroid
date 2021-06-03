package test

import (
	"fmt"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockChannelService struct {
	service.ChannelService
	mock.Mock
}

func NewMockChannelService() *MockChannelService {
	return &MockChannelService{}
}

func (mcs *MockChannelService) Get(id int) (common.Channel, error) {
	args := mcs.Called(id)
	fmt.Printf("[MockChannelService.Get] id=%d", id)
	return args.Get(0).(common.Channel), args.Error(1)
}

func (mcs *MockChannelService) GetAll(user common.UserAccount, controllerID int) ([]common.Channel, error) {
	args := mcs.Called(user, controllerID)
	fmt.Printf("[MockChannelService.GetAll] controllerID=%d, user=%+v", controllerID, user)
	return args.Get(0).([]common.Channel), args.Error(1)
}

func (mcs *MockChannelService) Update(model common.Channel) error {
	args := mcs.Called(model)
	fmt.Printf("[MockChannelService.Get] model=%+v", model)
	return args.Error(1)
}
