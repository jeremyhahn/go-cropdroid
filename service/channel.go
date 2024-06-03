package service

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type ChannelServicer interface {
	Get(session Session, id uint64) (model.Channel, error)
	GetByDeviceID(session Session, deviceID uint64) ([]model.Channel, error)
	Update(session Session, channel model.Channel) error
}

type ChannelService struct {
	dao    dao.ChannelDAO
	mapper mapper.ChannelMapper
	ChannelServicer
}

func NewChannelService(
	dao dao.ChannelDAO,
	mapper mapper.ChannelMapper) ChannelServicer {

	return &ChannelService{
		dao:    dao,
		mapper: mapper}
}

func (service *ChannelService) Get(session Session, id uint64) (model.Channel, error) {
	orgID := session.GetRequestedOrganizationID()
	farmID := session.GetRequestedFarmID()
	consistencyLevel := session.GetFarmService().GetConsistencyLevel()
	entity, err := service.dao.Get(orgID, farmID, id, consistencyLevel)
	if err != nil {
		return nil, err
	}
	return service.mapper.MapConfigToModel(entity), nil
}

func (service *ChannelService) GetByDeviceID(session Session, deviceID uint64) ([]model.Channel, error) {
	farmService := session.GetFarmService()
	orgID := session.GetRequestedOrganizationID()
	farmID := farmService.GetFarmID()
	consistencyLevel := farmService.GetConsistencyLevel()
	entities, err := service.dao.GetByDevice(orgID, farmID, deviceID, consistencyLevel)
	if err != nil {
		return nil, err
	}
	channels := make([]model.Channel, len(entities))
	for i, entity := range entities {
		channels[i] = service.mapper.MapConfigToModel(entity)
	}
	return channels, nil
}

func (service *ChannelService) Update(session Session, channel model.Channel) error {
	orgID := session.GetRequestedOrganizationID()
	farmID := session.GetRequestedFarmID()
	consistencyLevel := session.GetFarmService().GetConsistencyLevel()
	persisted, err := service.dao.Get(orgID, farmID, channel.Identifier(), consistencyLevel)
	if err != nil {
		return err
	}
	channel.SetDeviceID(persisted.GetDeviceID())
	channel.SetBoardID(persisted.GetBoardID())
	channelStruct := service.mapper.MapModelToConfig(channel).(*config.ChannelStruct)
	if err = service.dao.Save(farmID, channelStruct); err != nil {
		return err
	}
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		if device.ID == channel.GetDeviceID() {
			for _, ch := range device.GetChannels() {
				if ch.ID == channel.Identifier() {
					// conditions and schedules not sent by android client
					// android ui bug?
					channelStruct.SetConditions(ch.GetConditions())
					channelStruct.SetSchedule(ch.GetSchedule())
					break
				}
			}
			device.SetChannel(channelStruct)
			return farmService.SetDeviceConfig(device)
		}
	}
	return ErrChannelNotFound
}
