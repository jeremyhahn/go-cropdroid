// +build !cluster

package service

import (
	"errors"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
	logging "github.com/op/go-logging"
)

type ConditionService interface {
	GetListView(session Session, channelID int) ([]*viewmodel.Condition, error)
	GetConditions(session Session, deviceID int) ([]config.ConditionConfig, error)
	Create(session Session, condition config.ConditionConfig) (config.ConditionConfig, error)
	Update(session Session, condition config.ConditionConfig) error
	Delete(session Session, condition config.ConditionConfig) error
	IsTrue(condition config.ConditionConfig, value float64) (bool, error)
}

type DefaultConditionService struct {
	logger        *logging.Logger
	dao           dao.ConditionDAO
	mapper        mapper.ConditionMapper
	configService ConfigService
	ConditionService
}

// NewConditionService creates a new default ConditionService instance using the current time for calculations
func NewConditionService(logger *logging.Logger, conditionDAO dao.ConditionDAO,
	conditionMapper mapper.ConditionMapper, configService ConfigService) ConditionService {
	return &DefaultConditionService{
		logger:        logger,
		dao:           conditionDAO,
		mapper:        conditionMapper,
		configService: configService}
}

// GetConditions retrieves a list of condition entries from the database
func (service *DefaultConditionService) GetListView(session Session, channelID int) ([]*viewmodel.Condition, error) {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	orgID := farmConfig.GetOrgID()
	userID := session.GetUser().GetID()
	session.GetLogger().Debugf("orgID=%d, userID=%d, channelID=%d", orgID, userID, channelID)
	conditionEntities, err := service.dao.GetByOrgUserAndChannelID(orgID, userID, channelID)
	if err != nil {
		return nil, err
	}
	conditions := make([]*viewmodel.Condition, len(conditionEntities))
	for i, conditionEntity := range conditionEntities {
		err = func(conditionEntity config.Condition) error {
			for _, device := range farmConfig.GetDevices() {
				for _, metric := range device.GetMetrics() {
					if metric.GetID() == conditionEntity.GetMetricID() {
						conditions[i] = service.mapper.MapEntityToView(&conditionEntity, device.GetType(), &metric, channelID)
						return nil
					}
				}
			}
			return fmt.Errorf("Device for channel id %d not found", channelID)
		}(conditionEntity)

		if err != nil {
			return nil, err
		}
	}
	return conditions, nil
}

/*
// GetCondition retrieves a specific condition entry from the database
func (service *DefaultConditionService) GetCondition(channelID int) ([]config.ConditionConfig, error) {
	entities, err := service.dao.GetByChannelID(channelID)
	if err != nil {
		return nil, err
	}
	conditions := make([]config.ConditionConfig, len(entities))
	for i, entity := range entities {
		conditions[i] = service.mapper.MapEntityToModel(&entity)
	}
	return conditions, nil
}*/

// GetConditions retrieves a list of condition entries from the database
func (service *DefaultConditionService) GetConditions(session Session, deviceID int) ([]config.ConditionConfig, error) {
	farmService := session.GetFarmService()
	orgID := farmService.GetConfig().GetOrgID()
	userID := session.GetUser().GetID()
	session.GetLogger().Debugf("orgID=%d, userID=%d, deviceID=%d", orgID, userID, deviceID)
	entities, err := service.dao.GetByOrgUserAndChannelID(orgID, userID, deviceID)
	if err != nil {
		return nil, err
	}
	conditions := make([]config.ConditionConfig, len(entities))
	for i, entity := range entities {
		conditions[i] = service.mapper.MapEntityToModel(&entity)
	}
	return conditions, nil
}

// Create a new condition entry
func (service *DefaultConditionService) Create(session Session, condition config.ConditionConfig) (config.ConditionConfig, error) {

	service.logger.Debugf("Creating condition config: %+v", condition)
	entity := service.mapper.MapModelToEntity(condition)
	if err := service.dao.Create(entity); err != nil {
		return nil, err
	}
	condition.SetID(entity.GetID())
	//service.configService.Reload() // TODO: etcd
	//return condition, nil
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		for _, channel := range device.GetChannels() {
			if channel.GetID() == condition.GetChannelID() {
				channel.AddCondition(condition)
				device.SetChannel(&channel)
				farmConfig.SetDevice(&device)
				if err := farmService.SetConfig(farmConfig); err != nil {
					return nil, err
				}
				return condition, nil
			}
		}
	}
	return nil, ErrConditionNotFound
}

// Update an existing condition entry in the database
func (service *DefaultConditionService) Update(session Session, condition config.ConditionConfig) error {
	service.logger.Debugf("Updating condition config: %+v", condition)
	entity := service.mapper.MapModelToEntity(condition)
	if err := service.dao.Save(entity); err != nil {
		return err
	}
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		for _, channel := range device.GetChannels() {
			if channel.GetID() == condition.GetChannelID() {
				channel.SetCondition(condition)
				device.SetChannel(&channel)
				farmConfig.SetDevice(&device)
				farmService.SetConfig(farmConfig)
				return nil
			}
		}
	}
	return ErrConditionNotFound
}

// Delete a condition entry from the database
func (service *DefaultConditionService) Delete(session Session, condition config.ConditionConfig) error {
	service.logger.Debugf("Deleting condition config: %+v", condition)
	entity := service.mapper.MapModelToEntity(condition)
	if err := service.dao.Delete(entity); err != nil {
		return err
	}
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		for _, channel := range device.GetChannels() {
			for i, _condition := range channel.GetConditions() {
				if _condition.GetID() == condition.GetID() {
					channel.Conditions = append(channel.Conditions[:i], channel.Conditions[i+1:]...)
					device.SetChannel(&channel)
					farmConfig.SetDevice(&device)
					farmService.SetConfig(farmConfig)
					return nil
				}
			}
		}
	}
	return ErrConditionNotFound
}

func (service *DefaultConditionService) IsTrue(condition config.ConditionConfig, value float64) (bool, error) {
	service.logger.Debugf("condition=%+v, value=%.2f", condition, value)
	switch condition.GetComparator() {
	case ">":
		if value > condition.GetThreshold() {
			return true, nil
		}
		return false, nil
	case "<":
		if value < condition.GetThreshold() {
			return true, nil
		}
		return false, nil
	case ">=":
		if value >= condition.GetThreshold() {
			return true, nil
		}
		return false, nil
	case "<=":
		if value <= condition.GetThreshold() {
			return true, nil
		}
		return false, nil
	case "=":
		if value == condition.GetThreshold() {
			return true, nil
		}
		return false, nil
	}
	return false, errors.New(fmt.Sprintf("Unsupported comparison operator: %s", condition.GetComparator()))
}
