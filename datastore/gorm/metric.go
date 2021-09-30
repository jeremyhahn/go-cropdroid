package gorm

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormMetricDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.MetricDAO
}

func NewMetricDAO(logger *logging.Logger, db *gorm.DB) dao.MetricDAO {
	return &GormMetricDAO{logger: logger, db: db}
}

/*
func (metricDAO *GormMetricDAO) Create(metric config.MetricConfig) error {
	metricDAO.logger.Debugf("Creating metric record")
	return metricDAO.db.Create(metric).Error
}

func (metricDAO *GormMetricDAO) Update(metric config.MetricConfig) error {
	metricDAO.logger.Debugf("Updating metric record")
	return metricDAO.db.Update(metric).Error
}

*/
func (metricDAO *GormMetricDAO) Save(metric config.MetricConfig) error {
	metricDAO.logger.Debugf(fmt.Sprintf("Saving metric record: %+v", metric))
	return metricDAO.db.Save(metric).Error
}

func (metricDAO *GormMetricDAO) Get(metricID int) (config.MetricConfig, error) {
	metricDAO.logger.Debugf("Getting metric id %d", metricID)
	var entity config.Metric
	if err := metricDAO.db.First(&entity, metricID).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (metricDAO *GormMetricDAO) GetByDeviceID(deviceID uint64) ([]config.Metric, error) {
	metricDAO.logger.Debugf("Getting metric record for device %d", deviceID)
	var entities []config.Metric
	if err := metricDAO.db.Where("device_id = ?", deviceID).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (metricDAO *GormMetricDAO) GetByOrgUserAndDeviceID(orgID, userID, deviceID uint64) ([]config.Metric, error) {
	metricDAO.logger.Debugf("Getting metric record for org '%d'", orgID)
	var Metrics []config.Metric
	if err := metricDAO.db.Table("metrics").
		Select("metrics.*").
		Joins("JOIN devices on metrics.device_id = devices.id").
		Joins("JOIN farms on farms.id = devices.farm_id AND farms.organization_id = ?", orgID).
		Joins("JOIN permissions on farms.id = permissions.farm_id").
		Where("metrics.device_id = ? and permissions.user_id = ?", deviceID, userID).
		Find(&Metrics).Error; err != nil {
		return nil, err
	}
	return Metrics, nil
}
