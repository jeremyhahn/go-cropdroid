//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"

	logging "github.com/op/go-logging"
)

type RaftDeviceConfigDAODAO struct {
	logger  *logging.Logger
	raft    cluster.RaftNode
	farmDAO dao.FarmDAO
	dao.DeviceDAO
}

func NewRaftDeviceConfigDAO(logger *logging.Logger,
	raftNode cluster.RaftNode, farmDAO dao.FarmDAO) dao.DeviceDAO {

	logger.Debugf("Creating new *config.Device Raft DAO")
	return &RaftDeviceConfigDAODAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftDeviceConfigDAODAO) Save(device *config.DeviceStruct) error {
	dao.logger.Debugf("Save Raft entity *state.DeviceState: %d", device.ID)
	idSetter := dao.raft.GetParams().IdSetter
	idSetter.SetDeviceIds(device.GetFarmID(), []*config.DeviceStruct{device})
	farmConfig, err := dao.farmDAO.Get(device.GetFarmID(), common.CONSISTENCY_LOCAL)
	if err != nil {
		dao.logger.Debugf("Save error looking up farm: %d. error: %s", farmConfig.ID, err)
		return err
	}
	farmConfig.SetDevice(device)
	dao.logger.Debugf("Saving device config %+v", device)
	return dao.farmDAO.Save(farmConfig)
}

func (dao *RaftDeviceConfigDAODAO) Get(farmID, deviceID uint64, CONSISTENCY_LEVEL int) (*config.DeviceStruct, error) {
	dao.logger.Debugf("Get Raft entity *state.DeviceState for farmID: %d, deviceID: %d", farmID, deviceID)
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		dao.logger.Debugf("Get error looking up farmID: %d", farmID)
		return nil, err
	}
	device, err := farmConfig.GetDeviceById(deviceID)
	if err != nil {
		dao.logger.Debugf("Get error looking up deviceID: %d. error: ", device.ID, err)
		return nil, datastore.ErrRecordNotFound
	}
	if err := device.ParseSettings(); err != nil {
		dao.logger.Debugf("Get error parsing settings for deviceID: %d. error: %s", device.ID, err)
		return nil, err
	}
	return device, nil
}
