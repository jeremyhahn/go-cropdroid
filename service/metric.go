package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type MetricService interface {
	Get(id int) (common.Metric, error)
	GetAll(session Session, deviceID int) ([]common.Metric, error)
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

func (service *DefaultMetricService) Get(id int) (common.Metric, error) {
	entity, err := service.dao.Get(id)
	if err != nil {
		return nil, err
	}
	return service.mapper.MapEntityToModel(entity), nil
}

func (service *DefaultMetricService) GetAll(session Session, deviceID int) ([]common.Metric, error) {
	orgID := session.GetFarmService().GetConfig().GetOrgID()
	userID := session.GetUser().GetID()
	entities, err := service.dao.GetByOrgUserAndDeviceID(orgID, userID, deviceID)
	if err != nil {
		return nil, err
	}
	metricViews := make([]common.Metric, len(entities))
	for i, entity := range entities {
		model := service.mapper.MapEntityToModel(&entity)
		metricViews[i] = model
	}
	return metricViews, nil
}

func (service *DefaultMetricService) Update(session Session, metric common.Metric) error {
	entity := service.mapper.MapModelToConfig(metric)
	persisted, err := service.dao.Get(entity.GetID())
	if err != nil {
		return err
	}
	entity.SetDeviceID(persisted.GetDeviceID())
	if err = service.dao.Save(&entity); err != nil {
		return err
	}

	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == metric.GetDeviceID() {
			metricConfig := service.mapper.MapModelToConfig(metric)
			device.SetMetric(&metricConfig)
			return farmService.SetDeviceConfig(&device)
		}
	}
	return ErrMetricNotFound
}
