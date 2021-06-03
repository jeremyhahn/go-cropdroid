// +build !cluster

package service

import (
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jeremyhahn/cropdroid/mapper"
)

type ChannelService interface {
	Get(id int) (common.Channel, error)
	GetAll(session Session, controllerID int) ([]common.Channel, error)
	Update(session Session, viewModel common.Channel) error
}

type DefaultChannelService struct {
	dao           dao.ChannelDAO
	mapper        mapper.ChannelMapper
	configService ConfigService
	ChannelService
}

func NewChannelService(dao dao.ChannelDAO, mapper mapper.ChannelMapper, configService ConfigService) ChannelService {
	return &DefaultChannelService{
		dao:           dao,
		mapper:        mapper,
		configService: configService}
}

func (service *DefaultChannelService) Get(id int) (common.Channel, error) {
	entity, err := service.dao.Get(id)
	if err != nil {
		return nil, err
	}
	return service.mapper.MapEntityToModel(entity), nil
}

func (service *DefaultChannelService) GetAll(session Session, controllerID int) ([]common.Channel, error) {
	orgID := session.GetFarmService().GetConfig().GetOrgID()
	userID := session.GetUser().GetID()
	entities, err := service.dao.GetByOrgUserAndControllerID(orgID, userID, controllerID)
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
	channelConfig.SetControllerID(persisted.GetControllerID())
	channelConfig.SetChannelID(persisted.GetChannelID())
	if err = service.dao.Save(channelConfig); err != nil {
		return err
	}

	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, controller := range farmConfig.GetControllers() {
		if controller.GetID() == channelConfig.GetControllerID() {
			controller.SetChannel(channelConfig)
			farmConfig.SetController(&controller)
			return farmService.SetConfig(farmConfig)
		}
	}
	return ErrChannelNotFound
}
