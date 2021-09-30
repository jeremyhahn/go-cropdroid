// +build cluster

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
	GetListView(session Session, channelID uint64) ([]*viewmodel.Condition, error)
	GetConditions(session Session, channelID uint64) ([]config.ConditionConfig, error)
	Create(session Session, condition config.ConditionConfig) (config.ConditionConfig, error)
	Update(session Session, condition config.ConditionConfig) error
	Delete(session Session, condition config.ConditionConfig) error
	IsTrue(condition config.ConditionConfig, value float64) (bool, error)
}

type DefaultConditionService struct {
	logger *logging.Logger
	dao    dao.ConditionDAO
	mapper mapper.ConditionMapper
	ConditionService
}

// NewConditionService creates a new default ConditionService instance using the current time for calculations
func NewConditionService(logger *logging.Logger, conditionDAO dao.ConditionDAO,
	conditionMapper mapper.ConditionMapper) ConditionService {
	return &DefaultConditionService{
		logger: logger,
		dao:    conditionDAO,
		mapper: conditionMapper}
}

// GetConditions retrieves a list of condition entries from the database
func (service *DefaultConditionService) GetListView(session Session, channelID uint64) ([]*viewmodel.Condition, error) {

	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	orgID := farmConfig.GetOrganizationID()
	userID := session.GetUser().GetID()

	session.GetLogger().Debugf("[ConditionService.GetConditions] orgID=%d, userID=%d, channelID=%d", orgID, userID, channelID)

	for _, device := range farmConfig.GetDevices() {
		for _, channel := range device.GetChannels() {
			if channel.GetID() == channelID {
				_conditions := channel.GetConditions()
				conditions := make([]*viewmodel.Condition, len(_conditions))
				for i, condition := range _conditions {
					err := func(condition config.Condition) error {
						for _, device := range farmConfig.GetDevices() {
							for _, metric := range device.GetMetrics() {
								if metric.GetID() == condition.GetMetricID() {
									conditions[i] = service.mapper.MapEntityToView(&condition, device.GetType(), &metric, channelID)
									return nil
								}
							}
						}
						return fmt.Errorf("Device for channel id %d not found", channelID)
					}(condition)

					if err != nil {
						return nil, err
					}
				}
				return conditions, nil
			}
		}
	}
	return nil, ErrChannelNotFound
}

// GetConditions retrieves a list of condition entries from the database
func (service *DefaultConditionService) GetConditions(session Session, channelID uint64) ([]config.ConditionConfig, error) {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	orgID := farmConfig.GetOrganizationID()
	userID := session.GetUser().GetID()
	session.GetLogger().Debugf("[ConditionService.GetConditions] orgID=%d, userID=%d, channelID=%d", orgID, userID, channelID)
	for _, device := range farmConfig.GetDevices() {
		for _, channel := range device.GetChannels() {
			if channel.GetID() == channelID {
				_conditions := channel.GetConditions()
				conditions := make([]config.ConditionConfig, len(_conditions))
				for i, condition := range _conditions {
					conditions[i] = service.mapper.MapEntityToModel(&condition)
				}
				return conditions, nil
			}
		}
	}
	return nil, ErrChannelNotFound
}

// Create a new condition entry
func (service *DefaultConditionService) Create(session Session, condition config.ConditionConfig) (config.ConditionConfig, error) {
	service.logger.Debugf("Creating condition config: %+v", condition)
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		for _, channel := range device.GetChannels() {
			if channel.GetID() == condition.GetChannelID() {
				condition.SetID(condition.Hash())
				channel.AddCondition(condition)
				device.SetChannel(&channel)
				farmConfig.SetDevice(&device)
				return condition, farmService.SetConfig(farmConfig)
			}
		}
	}
	return nil, ErrConditionNotFound
}

// Update an existing condition entry in the database
func (service *DefaultConditionService) Update(session Session, condition config.ConditionConfig) error {
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
