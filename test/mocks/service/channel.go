package service

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockChannelService struct {
	mock.Mock
	service.ChannelServicer
}

func NewMockChannelService() service.ChannelServicer {
	return &MockChannelService{}
}

func (service *MockChannelService) Get(session service.Session, id uint64) (model.Channel, error) {
	args := service.Called(session, id)
	return &model.ChannelStruct{ID: id}, args.Error(1)
}

func (service *MockChannelService) GetByDeviceID(session service.Session, deviceID uint64) ([]model.Channel, error) {
	pageQuery := query.NewPageQuery()
	models := make([]model.Channel, pageQuery.PageSize)
	for i := 0; i < pageQuery.PageSize; i++ {
		models[i] = &model.ChannelStruct{
			ID:   uint64(i),
			Name: fmt.Sprintf("Test Channel %d", i)}
	}
	args := service.Called(session, deviceID)
	return models, args.Error(1)
}

func (service *MockChannelService) Update(session service.Session, channel model.Channel) error {
	args := service.Called(session, channel)
	return args.Error(0)
}
