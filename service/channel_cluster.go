// +build cluster

package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type ChannelService interface {
	Get(id int) (common.Channel, error)
	GetAll(session Session, deviceID uint64) ([]common.Channel, error)
	Update(session Session, viewModel common.Channel) error
}

type DefaultChannelService struct {
	dao    dao.ChannelDAO
	mapper mapper.ChannelMapper
	ChannelService
}

func NewChannelService(dao dao.ChannelDAO, mapper mapper.ChannelMapper) ChannelService {
	return &DefaultChannelService{
		dao:    dao,
		mapper: mapper}
}

func (service *DefaultChannelService) Get(id int) (common.Channel, error) {
	entity, err := service.dao.Get(id)
	if err != nil {
		return nil, err
	}
	return service.mapper.MapEntityToModel(entity), nil
}

func (service *DefaultChannelService) GetAll(session Session, deviceID uint64) ([]common.Channel, error) {
	orgID := session.GetFarmService().GetConfig().GetOrgID()
	userID := session.GetUser().GetID()
	entities, err := service.dao.GetByOrgUserAndDeviceID(orgID, userID, deviceID)
	if err != nil {
		return nil, err
	}
	channelViews := make([]common.Channel, len(entities))
	for i, entity := range entities {
		model := service.mapper.MapEntityToModel(&entity)
		channelViews[i] = model
	}
	return channelViews, nil
}

func (service *DefaultChannelService) Update(session Session, viewModel common.Channel) error {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	channelConfig := service.mapper.MapModelToConfig(viewModel)
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == viewModel.GetDeviceID() {
			for i, channel := range device.GetChannels() {
				if channel.GetID() == viewModel.GetID() {
					c := channelConfig.(*config.Channel)
					device.Channels[i] = *c
					return farmService.SetConfig(farmConfig)
				}
			}
		}
	}
	// fmt.Errorf("Farm config not found in service: farm.id=%d", farmService.GetFarmID())
	return ErrFarmConfigNotFound
}
