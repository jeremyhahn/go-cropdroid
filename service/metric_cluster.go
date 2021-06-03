// +build cluster

package service

import (
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jeremyhahn/cropdroid/mapper"
)

type MetricService interface {
	Get(id int) (common.Metric, error)
	GetAll(session Session, controllerID int) ([]common.Metric, error)
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

func (service *DefaultMetricService) GetAll(session Session, controllerID int) ([]common.Metric, error) {
	orgID := session.GetFarmService().GetConfig().GetOrgID()
	userID := session.GetUser().GetID()
	entities, err := service.dao.GetByOrgUserAndControllerID(orgID, userID, controllerID)
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
	farmConfig := farmService.GetConfig()
	metricConfig := service.mapper.MapModelToConfig(metric)
	for _, controller := range farmConfig.GetControllers() {
		if controller.GetID() == metric.GetControllerID() {
			for i, m := range controller.GetMetrics() {
				if m.GetID() == metric.GetID() {
					controller.Metrics[i] = metricConfig
					return farmService.SetConfig(farmConfig)
				}
			}
		}
	}
	//return fmt.Errorf("Farm config not found in service: farm.id=%d", farmService.GetFarmID())
	return ErrFarmConfigNotFound
}
