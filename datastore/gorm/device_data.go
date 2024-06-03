package gorm

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/state"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormDeviceStore struct {
	logger   *logging.Logger
	db       *gorm.DB
	dbtype   string
	location *time.Location
	datastore.DeviceDataStore
}

func NewGormDeviceDataStore(logger *logging.Logger, db *gorm.DB, dbtype string,
	location *time.Location) datastore.DeviceDataStore {

	logger.Info("Using GormDeviceStore for device data store")
	return &GormDeviceStore{logger: logger, db: db, dbtype: dbtype, location: location}
}

func (gds *GormDeviceStore) Save(deviceID uint64, deviceState state.DeviceStateMap) error {
	tableName := fmt.Sprintf("state_%d", deviceID)
	metrics := deviceState.GetMetrics()
	metricKeys, metricValues := gds.parseMetricKeysAndValues(metrics)
	stringMetricValues := make([]string, len(metrics))
	for i, value := range metricValues {
		stringMetricValues[i] = strconv.FormatFloat(value, 'f', -1, 64)
	}

	channels := deviceState.GetChannels()
	channelKeys, channelValues := gds.parseChannelKeysAndValues(channels)
	stringChannelValues := make([]string, len(channels))
	for i, value := range channelValues {
		if value == 0 {
			stringChannelValues[i] = "0"
		} else {
			stringChannelValues[i] = "1"
		}
	}

	timestamp := time.Now().In(gds.location).Round(time.Microsecond).Format(common.TIME_FORMAT_LOCAL) // cockaroach doesnt like time zone included

	noSuchTableError := fmt.Sprintf("no such table: %s", tableName)
	doesNotExistError := fmt.Sprintf("pq: relation \"%s\" does not exist", tableName)
	alreadyExistsError := fmt.Sprintf("table \"%s\" already exists", tableName)

	insertSQL := fmt.Sprintf("INSERT INTO \"%s\"(device_id, %s, %s, timestamp) VALUES (%d, %s, %s, '%s')",
		tableName, strings.Join(metricKeys, ","), strings.Join(channelKeys, ","),
		deviceID, strings.Join(stringMetricValues, ","), strings.Join(stringChannelValues, ","), timestamp)

	if err := gds.db.Exec(insertSQL).Error; err != nil {
		if err.Error() == noSuchTableError || err.Error() == doesNotExistError {
			if err2 := gds.createTable(tableName, deviceState); err2 != nil {
				gds.logger.Errorf("Unable to create table %s: %s", tableName, err2)
				return err2
			}
			if err3 := gds.Save(deviceID, deviceState); err3 != nil {
				gds.logger.Errorf("Unable to save device state for table %s: %s", tableName, err3)
				return err3
			}
			err = nil
			return nil
		}
		if err.Error() == alreadyExistsError {
			return nil
		}
		gds.logger.Errorf("[GormDeviceStore.Save] Error:%s", err.Error())
		return err
	}
	return nil
}

func (gds *GormDeviceStore) GetLast30Days(deviceID uint64, metric string) ([]float64, error) {
	gds.logger.Debugf("Getting metric history for device: %d", deviceID)
	tableName := fmt.Sprintf("state_%d", deviceID)
	var metricValues []float64
	sinceDate := time.Now().AddDate(0, -30, 0)
	rows, err := gds.db.Table(tableName).Select(metric).Where("timestamp >= ?", sinceDate).Limit(500).Rows()
	if err != nil {
		gds.logger.Error(err)
		return nil, err
	}
	var metricValue float64
	for rows.Next() {
		rows.Scan(&metricValue)
		metricValues = append(metricValues, metricValue)
	}
	return metricValues, nil
}

func (gds *GormDeviceStore) createTable(tableName string, deviceState state.DeviceStateMap) error {

	var columnSQL string
	metrics := deviceState.GetMetrics()
	metricKeys, _ := gds.parseMetricKeysAndValues(metrics)
	sort.Strings(metricKeys)

	channels := deviceState.GetChannels()
	channelKeys, channelValues := gds.parseChannelKeysAndValues(channels)
	stringChannelValues := make([]string, len(channels))
	for i, value := range channelValues {
		if value == 0 {
			stringChannelValues[i] = "0"
		} else {
			stringChannelValues[i] = "1"
		}
	}

	pgSQL := fmt.Sprintf("CREATE TABLE \"%s\" (id bigserial,device_id int,%s NUMERIC,%s INTEGER,\"timestamp\" timestamp without time zone not null default (current_timestamp at time zone '%s'), primary key (id))",
		tableName, strings.Join(metricKeys, " NUMERIC,"), strings.Join(channelKeys, " INTEGER,"), gds.location.String())

	sqliteSQL := fmt.Sprintf("CREATE TABLE %s(id INTEGER primary key,device_id INTEGER,%s,%s ", tableName, strings.Join(metricKeys, " REAL,"), strings.Join(channelKeys, " INTEGER,"))
	sqliteSQL = fmt.Sprintf("%s INTEGER,\"timestamp\" datetime)", sqliteSQL)

	if gds.dbtype == "sqlite" || gds.dbtype == "memory" {
		columnSQL = sqliteSQL
	} else {
		columnSQL = pgSQL
	}

	gds.logger.Errorf("columnSQL: %s", columnSQL)

	if err := gds.db.Exec(columnSQL).Error; err != nil {
		gds.logger.Errorf("[GormDeviceStore.CreateTable] Error:%s", err.Error())
		return err
	}
	return nil
}

func (gds *GormDeviceStore) parseMetricKeysAndValues(data map[string]float64) ([]string, []float64) {
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

func (gds *GormDeviceStore) parseChannelKeysAndValues(data []int) ([]string, []int) {
	keys := make([]string, len(data))
	values := make([]int, len(data))
	for i, v := range data {
		keys[i] = fmt.Sprintf("c%d", i)
		values[i] = v
	}
	return keys, values
}
