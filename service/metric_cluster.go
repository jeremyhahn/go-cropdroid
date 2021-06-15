// +build cluster

package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/config/store"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type MetricService interface {
	Get(id int) (common.Metric, error)
	GetAll(session Session, deviceID int) ([]common.Metric, error)
	Update(session Session, metric common.Metric) error
}

type DefaultMetricService struct {
	dao           dao.MetricDAO
	mapper        mapper.MetricMapper
	configService ConfigService
	MetricService
}

func NewMetricService(dao dao.MetricDAO, mapper mapper.MetricMapper, configService ConfigService) MetricService {
	return &DefaultMetricService{
		dao:           dao,
		mapper:        mapper,
		configService: configService}
}

func (service *DefaultMetricService) Get(id int) (common.Metric, error) {
	entity, err := service.dao.Get(id)
	if err != nil {
		return nil, err
	}
	return service.mapper.MapEntityToModel(entity), nil
}

func (service *DefaultMetricService) GetAll(session Session, deviceID int) ([]common.Metric, error) {
	orgID := session.GetFarmService().GetConfig(store.READ_COMMITTED).GetOrgID()
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
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig(store.READ_COMMITTED)
	metricConfig := service.mapper.MapModelToConfig(metric)
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == metric.GetDeviceID() {
			for i, m := range device.GetMetrics() {
				if m.GetID() == metric.GetID() {
					device.Metrics[i] = metricConfig
					return farmService.SetConfig(farmConfig)
				}
			}
		}
	}
	//return fmt.Errorf("Farm config not found in service: farm.id=%d", farmService.GetFarmID())
	return ErrFarmConfigNotFound
}
