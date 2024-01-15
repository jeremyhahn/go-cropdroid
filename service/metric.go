package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type MetricService interface {
	Get(session Session, deviceID, metricID uint64) (common.Metric, error)
	GetAll(session Session, deviceID uint64) ([]common.Metric, error)
	Update(session Session, metric common.Metric) error
}

type DefaultMetricService struct {
	dao    dao.MetricDAO
	mapper mapper.MetricMapper
	MetricService
}

func NewMetricService(dao dao.MetricDAO, mapper mapper.MetricMapper) MetricService {
	return &DefaultMetricService{
		dao:    dao,
		mapper: mapper}
}

func (service *DefaultMetricService) Get(session Session, deviceID,
	metricID uint64) (common.Metric, error) {
	farmID := session.GetRequestedFarmID()
	consistencyLevel := session.GetFarmService().GetConsistencyLevel()
	entity, err := service.dao.Get(farmID, deviceID, metricID, consistencyLevel)
	if err != nil {
		return nil, err
	}
	return service.mapper.MapConfigToModel(entity), nil
}

func (service *DefaultMetricService) GetAll(session Session, deviceID uint64) ([]common.Metric, error) {
	farmID := session.GetRequestedFarmID()
	consistencyLevel := session.GetFarmService().GetConsistencyLevel()
	entities, err := service.dao.GetByDevice(farmID, deviceID, consistencyLevel)
	if err != nil {
		return nil, err
	}
	metricViews := make([]common.Metric, len(entities))
	for i, entity := range entities {
		model := service.mapper.MapConfigToModel(entity)
		metricViews[i] = model
	}
	return metricViews, nil
}

func (service *DefaultMetricService) Update(session Session, metric common.Metric) error {
	farmID := session.GetRequestedFarmID()
	consistencyLevel := session.GetFarmService().GetConsistencyLevel()
	entity := service.mapper.MapModelToConfig(metric)
	persisted, err := service.dao.Get(farmID, metric.GetDeviceID(),
		entity.GetID(), consistencyLevel)
	if err != nil {
		return err
	}
	entity.SetDeviceID(persisted.GetDeviceID())
	if err = service.dao.Save(farmID, entity); err != nil {
		return err
	}

	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == metric.GetDeviceID() {
			metricConfig := service.mapper.MapModelToConfig(metric)
			device.SetMetric(metricConfig)
			return farmService.SetDeviceConfig(device)
		}
	}
	return ErrMetricNotFound
}
