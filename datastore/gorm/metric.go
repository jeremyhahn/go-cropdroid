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

func (metricDAO *GormMetricDAO) Save(farmID uint64, metric *config.Metric) error {
	metricDAO.logger.Debugf(fmt.Sprintf("Saving metric record: %+v", metric))
	return metricDAO.db.Save(metric).Error
}

func (metricDAO *GormMetricDAO) Get(farmID, deviceID,
	metricID uint64, CONSISTENCY_LEVEL int) (*config.Metric, error) {

	metricDAO.logger.Debugf("Getting metric id %d", metricID)
	var entity config.Metric
	if err := metricDAO.db.First(&entity, metricID).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (metricDAO *GormMetricDAO) GetByDevice(farmID, deviceID uint64,
	CONSISTENCY_LEVEL int) ([]*config.Metric, error) {

	metricDAO.logger.Debugf("Getting metric record for farm %d", farmID)
	var metrics []*config.Metric
	// if err := metricDAO.db.Table("metrics").
	// 	Select("metrics.*").
	// 	Joins("JOIN devices on metrics.device_id = devices.id").
	// 	Joins("JOIN farms on farms.id = devices.farm_id AND farms.organization_id = ?", orgID).
	// 	Joins("JOIN permissions on farms.id = permissions.farm_id").
	// 	Where("metrics.device_id = ? and permissions.user_id = ?", deviceID, userID).
	// 	Find(&metrics).Error; err != nil {
	// 	return nil, err
	// }
	if err := metricDAO.db.Table("metrics").
		Select("metrics.*").
		Joins("JOIN devices on metrics.device_id = devices.id").
		Joins("JOIN farms on farms.id = devices.farm_id").
		Where("devices.id = ? and farms.id = ?", deviceID, farmID).
		Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}
