// +build !cluster

package service

import (
	"errors"
	"fmt"

	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/viewmodel"
	logging "github.com/op/go-logging"
)

type ConditionService interface {
	GetListView(session Session, channelID int) ([]*viewmodel.Condition, error)
	GetConditions(session Session, controllerID int) ([]config.ConditionConfig, error)
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
	session.GetLogger().Debugf("[ConditionService.GetListView] orgID=%d, userID=%d, channelID=%d", orgID, userID, channelID)
	conditionEntities, err := service.dao.GetByOrgUserAndChannelID(orgID, userID, channelID)
	if err != nil {
		return nil, err
	}
	conditions := make([]*viewmodel.Condition, len(conditionEntities))
	for i, conditionEntity := range conditionEntities {
		err = func(conditionEntity config.Condition) error {
			for _, controller := range farmConfig.GetControllers() {
				for _, metric := range controller.GetMetrics() {
					if metric.GetID() == conditionEntity.GetMetricID() {
						conditions[i] = service.mapper.MapEntityToView(&conditionEntity, controller.GetType(), &metric, channelID)
						return nil
					}
				}
			}
			return fmt.Errorf("Controller for channel id %d not found", channelID)
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
func (service *DefaultConditionService) GetConditions(session Session, controllerID int) ([]config.ConditionConfig, error) {
	farmService := session.GetFarmService()
	orgID := farmService.GetConfig().GetOrgID()
	userID := session.GetUser().GetID()
	session.GetLogger().Debugf("[ConditionService.GetConditions] orgID=%d, userID=%d, controllerID=%d", orgID, userID, controllerID)
	entities, err := service.dao.GetByOrgUserAndChannelID(orgID, userID, controllerID)
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
	for _, controller := range farmConfig.GetControllers() {
		for _, channel := range controller.GetChannels() {
			if channel.GetID() == condition.GetChannelID() {
				channel.AddCondition(condition)
				controller.SetChannel(&channel)
				farmConfig.SetController(&controller)
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
	for _, controller := range farmConfig.GetControllers() {
		for _, channel := range controller.GetChannels() {
			if channel.GetID() == condition.GetChannelID() {
				channel.SetCondition(condition)
				controller.SetChannel(&channel)
				farmConfig.SetController(&controller)
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
	for _, controller := range farmConfig.GetControllers() {
		for _, channel := range controller.GetChannels() {
			for i, _condition := range channel.GetConditions() {
				if _condition.GetID() == condition.GetID() {
					channel.Conditions = append(channel.Conditions[:i], channel.Conditions[i+1:]...)
					controller.SetChannel(&channel)
					farmConfig.SetController(&controller)
					farmService.SetConfig(farmConfig)
					return nil
				}
			}
		}
	}
	return ErrConditionNotFound
}

func (service *DefaultConditionService) IsTrue(condition config.ConditionConfig, value float64) (bool, error) {
	service.logger.Debugf("[DefaultConditionService.IsTrue] condition=%+v, value=%.2f", condition, value)
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
