package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type ChannelService interface {
	Get(session Session, id uint64) (common.Channel, error)
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

func (service *DefaultChannelService) Get(session Session, id uint64) (common.Channel, error) {
	orgID := session.GetRequestedOrganizationID()
	farmID := session.GetRequestedFarmID()
	consistencyLevel := session.GetFarmService().GetConsistencyLevel()
	entity, err := service.dao.Get(orgID, farmID, id, consistencyLevel)
	if err != nil {
		return nil, err
	}
	return service.mapper.MapConfigToModel(entity), nil
}

func (service *DefaultChannelService) GetAll(session Session, deviceID uint64) ([]common.Channel, error) {
	farmService := session.GetFarmService()
	orgID := session.GetRequestedOrganizationID()
	//userID := session.GetUser().ID
	farmID := farmService.GetFarmID()
	consistencyLevel := farmService.GetConsistencyLevel()
	entities, err := service.dao.GetByDevice(orgID, farmID, deviceID, consistencyLevel)
	if err != nil {
		return nil, err
	}
	channelViews := make([]common.Channel, len(entities))
	for i, entity := range entities {
		model := service.mapper.MapConfigToModel(entity)
		channelViews[i] = model
	}
	return channelViews, nil
}

func (service *DefaultChannelService) Update(session Session, viewModel common.Channel) error {
	channelConfig := service.mapper.MapModelToConfig(viewModel)
	orgID := session.GetRequestedOrganizationID()
	farmID := session.GetRequestedFarmID()
	consistencyLevel := session.GetFarmService().GetConsistencyLevel()
	persisted, err := service.dao.Get(orgID, farmID, channelConfig.ID, consistencyLevel)
	if err != nil {
		return err
	}
	channelConfig.SetDeviceID(persisted.GetDeviceID())
	channelConfig.SetChannelID(persisted.GetChannelID())
	if err = service.dao.Save(farmID, channelConfig); err != nil {
		return err
	}
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		if device.ID == channelConfig.GetDeviceID() {
			for _, ch := range device.GetChannels() {
				if ch.ID == channelConfig.ID {
					// conditions and schedules not sent by android client
					// android ui bug?
					channelConfig.SetConditions(ch.GetConditions())
					channelConfig.SetSchedule(ch.GetSchedule())
					break
				}
			}
			device.SetChannel(channelConfig)
			return farmService.SetDeviceConfig(device)
		}
	}
	return ErrChannelNotFound
}
