package gorm

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/datastore"
	"github.com/jeremyhahn/cropdroid/state"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormControllerStateDAO struct {
	logger   *logging.Logger
	db       *gorm.DB
	dbtype   string
	location *time.Location
	datastore.ControllerStateDAO
}

func NewControllerStateDAO(logger *logging.Logger, db *gorm.DB, dbtype string, location *time.Location) datastore.ControllerStateDAO {
	return &GormControllerStateDAO{logger: logger, db: db, dbtype: dbtype, location: location}
}

func (dao *GormControllerStateDAO) parseMetricKeysAndValues(data map[string]float64) ([]string, []float64) {
	keys := make([]string, len(data))
	values := make([]float64, len(data))
	i := 0
	for k, v := range data {
		keys[i] = k
		values[i] = v
		i++
	}
	return keys, values
}

func (dao *GormControllerStateDAO) parseChannelKeysAndValues(data []int) ([]string, []int) {
	keys := make([]string, len(data))
	values := make([]int, len(data))
	for i, v := range data {
		keys[i] = fmt.Sprintf("c%d", i)
		values[i] = v
	}
	return keys, values
}

func (dao *GormControllerStateDAO) Save(controllerID int, controllerState state.ControllerStateMap) error {
	tableName := fmt.Sprintf("state_%d", controllerID)
	metrics := controllerState.GetMetrics()
	metricKeys, metricValues := dao.parseMetricKeysAndValues(metrics)
	stringMetricValues := make([]string, len(metrics))
	for i, value := range metricValues {
		stringMetricValues[i] = strconv.FormatFloat(value, 'f', -1, 64)
	}

	channels := controllerState.GetChannels()
	channelKeys, channelValues := dao.parseChannelKeysAndValues(channels)
	stringChannelValues := make([]string, len(channels))
	for i, value := range channelValues {
		if value == 0 {
			stringChannelValues[i] = "0"
		} else {
			stringChannelValues[i] = "1"
		}
	}

	timestamp := time.Now().In(dao.location).Round(time.Microsecond).Format(common.TIME_FORMAT_LOCAL) // cockaroach doesnt like time zone included

	noSuchTableError := fmt.Sprintf("no such table: %s", tableName)
	doesNotExistError := fmt.Sprintf("pq: relation \"%s\" does not exist", tableName)
	alreadyExistsError := fmt.Sprintf("table \"%s\" already exists", tableName)

	insertSQL := fmt.Sprintf("INSERT INTO \"%s\"(controller_id, %s, %s, timestamp) VALUES (%d, %s, %s, '%s')",
		tableName, strings.Join(metricKeys, ","), strings.Join(channelKeys, ","),
		controllerID, strings.Join(stringMetricValues, ","), strings.Join(stringChannelValues, ","), timestamp)

	if err := dao.db.Exec(insertSQL).Error; err != nil {
		if err.Error() == noSuchTableError || err.Error() == doesNotExistError {

			dao.logger.Warningf("[MetricDAO.Save] noSuchTableError || doesNotExistError ERROR CAUGHT! :: err=%s", err)

			if err2 := dao.createTable(tableName, controllerState); err2 != nil {
				dao.logger.Errorf("Unable to create table %s: %s", tableName, err2)
				return err2
			}
			if err3 := dao.Save(controllerID, controllerState); err3 != nil {
				dao.logger.Errorf("Unable to save controller state for table %s: %s", tableName, err3)
				return err3
			}
			err = nil
			return nil
		}
		if err.Error() == alreadyExistsError {

			dao.logger.Warningf("[MetricDAO.Save] ALREADY EXISTS ERROR CAUGHT! :: err=%s", err)

			return nil
		}
		dao.logger.Errorf("[MetricDAO.Save] Error:%s", err.Error())
		return err
	}
	return nil
}

func (dao *GormControllerStateDAO) GetLast30Days(controllerID int, metric string) ([]float64, error) {
	dao.logger.Debugf("Getting metric history for controller: %d", controllerID)
	tableName := fmt.Sprintf("state_%d", controllerID)
	var metricValues []float64
	sinceDate := time.Now().AddDate(0, -30, 0)
	rows, err := dao.db.Table(tableName).Select(metric).Where("timestamp >= ?", sinceDate).Limit(500).Rows()
	if err != nil {
		return nil, err
	}
	var metricValue float64
	for rows.Next() {
		rows.Scan(&metricValue)
		metricValues = append(metricValues, metricValue)
	}
	return metricValues, nil
}

func (dao *GormControllerStateDAO) createTable(tableName string, controllerState state.ControllerStateMap) error {

	var columnSQL string
	metrics := controllerState.GetMetrics()
	metricKeys, _ := dao.parseMetricKeysAndValues(metrics)
	sort.Strings(metricKeys)

	channels := controllerState.GetChannels()
	channelKeys, channelValues := dao.parseChannelKeysAndValues(channels)
	stringChannelValues := make([]string, len(channels))
	for i, value := range channelValues {
		if value == 0 {
			stringChannelValues[i] = "0"
		} else {
			stringChannelValues[i] = "1"
		}
	}

	pgSQL := fmt.Sprintf("CREATE TABLE \"%s\" (id bigserial,controller_id int,%s NUMERIC,%s INTEGER,\"timestamp\" timestamp without time zone not null default (current_timestamp at time zone '%s'), primary key (id))",
		tableName, strings.Join(metricKeys, " NUMERIC,"), strings.Join(channelKeys, " INTEGER,"), dao.location.String())

	sqliteSQL := fmt.Sprintf("CREATE TABLE %s(id INTEGER primary key,controller_id INTEGER,%s,%s ", tableName, strings.Join(metricKeys, " REAL,"), strings.Join(channelKeys, " INTEGER,"))
	sqliteSQL = fmt.Sprintf("%s INTEGER,\"timestamp\" datetime)", sqliteSQL)

	if dao.dbtype == "sqlite" || dao.dbtype == "memory" {
		columnSQL = sqliteSQL
	} else {
		columnSQL = pgSQL
	}

	dao.logger.Errorf("columnSQL: %s", columnSQL)

	if err := dao.db.Exec(columnSQL).Error; err != nil {
		dao.logger.Errorf("[MetricDAO.CreateTable] Error:%s", err.Error())
		return err
	}
	return nil
}
