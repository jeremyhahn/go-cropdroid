package mocks

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

func (dao *MockConditionDAO) Save(farmID, deviceID uint64,
	condition *config.Condition) error {

	args := dao.Called(condition)
	fmt.Println("Saving condition record")
	return args.Error(0)
}

func (dao *MockConditionDAO) Get(farmID, deviceID, channelID,
	conditionID uint64, CONSISTENCY_LEVEL int) (*config.Condition, error) {

	args := dao.Called(farmID, deviceID, channelID, conditionID)
	fmt.Printf("Getting condition for id=%d\n", conditionID)
	return args.Get(0).(*config.Condition), nil
}

func (dao *MockConditionDAO) GetByChannel(farmID, deviceID,
	channelID uint64) ([]*config.Condition, error) {

	args := dao.Called(farmID, deviceID, channelID)
	fmt.Printf("Getting condition by channel id=%d\n", channelID)
	return args.Get(0).([]*config.Condition), nil
}

// func (dao *MockConditionDAO) GetByUserOrgAndDeviceID(orgID, deviceID uint64) ([]config.Condition, error) {
// 	args := dao.Called(orgID, deviceID)
// 	fmt.Printf("Getting condition by device id=%d and org id=%d\n", deviceID, orgID)
// 	return args.Get(0).([]config.Condition), nil
// }
