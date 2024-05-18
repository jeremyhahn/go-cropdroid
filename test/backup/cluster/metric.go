//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"

	logging "github.com/op/go-logging"
)

type RaftMetricDAO struct {
	logger  *logging.Logger
	raft    cluster.RaftNode
	farmDAO dao.FarmDAO
	dao.MetricDAO
}

func NewRaftMetricDAO(logger *logging.Logger,
	raftNode cluster.RaftNode, farmDAO dao.FarmDAO) dao.MetricDAO {
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

	raftParams := dao.raft.GetParams()
	idSetter := raftParams.IdSetter

	devices := farmConfig.GetDevices()
	for _, device := range devices {
		deviceID := device.ID
		if deviceID == metric.GetDeviceID() {
			if metric.ID == 0 || metric.GetDeviceID() == 0 {
				idSetter.SetMetricIds(deviceID, []*config.Metric{metric})
			}
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
		if device.ID == deviceID {
			for _, metric := range device.GetMetrics() {
				if metric.ID == metricID {
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
