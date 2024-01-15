//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"

	logging "github.com/op/go-logging"
)

type RaftMetricDAO struct {
	logger  *logging.Logger
	raft    RaftNode
	farmDAO dao.FarmDAO
	dao.MetricDAO
}

func NewRaftMetricDAO(logger *logging.Logger,
	raftNode RaftNode, farmDAO dao.FarmDAO) dao.MetricDAO {
	return &RaftMetricDAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftMetricDAO) Save(farmID uint64, metric *config.Metric) error {
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	if metric.GetID() == 0 {
		key := fmt.Sprintf("%d-%s", farmID, metric.GetName())
		id := dao.raft.GetParams().IdGenerator.NewID(key)
		metric.SetID(id)
	}
	devices := farmConfig.GetDevices()
	for _, device := range devices {
		if device.GetID() == metric.GetDeviceID() {
			device.SetMetric(metric)
			farmConfig.SetDevice(device)
			return dao.farmDAO.Save(farmConfig)
		}
	}
	return datastore.ErrNotFound
}

func (dao *RaftMetricDAO) Get(farmID, deviceID, metricID uint64,
	CONSISTENCY_LEVEL int) (*config.Metric, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == deviceID {
			for _, metric := range device.GetMetrics() {
				if metric.GetID() == metricID {
					return metric, nil
				}
			}
		}
	}
	return nil, datastore.ErrNotFound
}

func (dao *RaftMetricDAO) GetByDevice(farmID, deviceID uint64,
	CONSISTENCY_LEVEL int) ([]*config.Metric, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	metric, err := farmConfig.GetDeviceById(deviceID)
	if err != nil {
		return nil, datastore.ErrNotFound
	}
	return metric.GetMetrics(), nil
}
