package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type ChannelService interface {
	Get(id uint64) (common.Channel, error)
	GetAll(session Session, deviceID uint64) ([]common.Channel, error)
	Update(session Session, viewModel common.Channel) error
}

type DefaultChannelService struct {
	dao         dao.ChannelDAO
	mapper      mapper.ChannelMapper
	consistency int
	ChannelService
}

func NewChannelService(dao dao.ChannelDAO, mapper mapper.ChannelMapper) ChannelService {
	return &DefaultChannelService{
		dao:    dao,
		mapper: mapper}
}

func (service *DefaultChannelService) Get(id uint64) (common.Channel, error) {
	entity, err := service.dao.Get(id)
	if err != nil {
		return nil, err
	}
	return service.mapper.MapEntityToModel(entity), nil
}

func (service *DefaultChannelService) GetAll(session Session, deviceID uint64) ([]common.Channel, error) {
	orgID := session.GetFarmService().GetConfig().GetOrganizationID()
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
	channelConfig := service.mapper.MapModelToConfig(viewModel)
	persisted, err := service.dao.Get(channelConfig.GetID())
	if err != nil {
		return err
	}
	channelConfig.SetDeviceID(persisted.GetDeviceID())
	channelConfig.SetChannelID(persisted.GetChannelID())
	if err = service.dao.Save(channelConfig); err != nil {
		return err
	}

	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == channelConfig.GetDeviceID() {
			for _, ch := range device.GetChannels() {
				if ch.GetID() == channelConfig.GetID() {
					// conditions and schedules not sent by android client
					// android ui bug?
					channelConfig.SetConditions(ch.GetConditions())
					channelConfig.SetSchedule(ch.GetSchedule())
					break
				}
			}
			device.SetChannel(channelConfig)
			return farmService.SetDeviceConfig(&device)
		}
	}
	return ErrChannelNotFound
}
