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

type RaftDeviceConfigDAO struct {
	logger  *logging.Logger
	raft    cluster.RaftNode
	farmDAO dao.FarmDAO
	dao.DeviceDAO
}

func NewRaftDeviceConfigDAO(logger *logging.Logger,
	raftNode cluster.RaftNode, farmDAO dao.FarmDAO) dao.DeviceDAO {
	return &RaftDeviceConfigDAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftDeviceConfigDAO) Save(device *config.Device) error {

	idSetter := dao.raft.GetParams().IdSetter
	idSetter.SetDeviceIds(device.GetFarmID(), []*config.Device{device})

	farmConfig, err := dao.farmDAO.Get(device.GetFarmID(), common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	farmConfig.SetDevice(device)
	dao.logger.Debugf("Saving device config %+v", device)
	return dao.farmDAO.Save(farmConfig)
}

func (dao *RaftDeviceConfigDAO) Get(farmID, deviceID uint64, CONSISTENCY_LEVEL int) (*config.Device, error) {
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	device, err := farmConfig.GetDeviceById(deviceID)
	if err != nil {
		return nil, datastore.ErrNotFound
	}
	if err := device.ParseSettings(); err != nil {
		return nil, err
	}
	return device, nil
}

// type RaftDeviceConfigDAO struct {
// 	logger *logging.Logger
// 	raft   cluster.RaftNode
// 	dao.DeviceDAO
// }

// func NewRaftDeviceConfigDAO(logger *logging.Logger,
// 	raftNode cluster.RaftNode, clusterID uint64) dao.DeviceDAO {

// 	return &RaftDeviceConfigDAO{
// 		logger: logger,
// 		raft:   raftNode}
// }

// func (dao *RaftDeviceConfigDAO) Save(device *config.Device) error {
// 	dao.logger.Debugf("Saving device: %+v", device)
// 	org, err := json.Marshal(*device.(*config.Device))
// 	if err != nil {
// 		dao.logger.Errorf("Error: %s", err)
// 		return err
// 	}
// 	proposal, err := statemachine.CreateProposal(
// 		statemachine.QUERY_TYPE_UPDATE, org).Serialize()
// 	if err != nil {
// 		dao.logger.Errorf("Error: %s", err)
// 		return err
// 	}
// 	if err := dao.raft.SyncPropose(device.GetID(), proposal); err != nil {
// 		dao.logger.Errorf("Error: %s", err)
// 		return err
// 	}
// 	return nil
// }

// func (dao *RaftDeviceConfigDAO) Get(id uint64, CONSISTENCY_LEVEL int) (*config.Device, error) {
// 	dao.logger.Debugf("Fetching device ID: %d", id)
// 	var result interface{}
// 	var err error
// 	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
// 		result, err = dao.raft.ReadLocal(id, nil)
// 		if err != nil {
// 			dao.logger.Errorf("Error (orgID=%d): %s", id, err)
// 			return nil, err
// 		}
// 	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
// 		result, err = dao.raft.SyncRead(id, nil)
// 		if err != nil {
// 			dao.logger.Errorf("Error (orgID=%d): %s", id, err)
// 			return nil, err
// 		}
// 	}
// 	if result != nil {
// 		records := result.([]*config.Device)
// 		if len(records) > 0 {
// 			return records[0], nil
// 		}
// 		return nil, nil
// 	}
// 	return nil, datastore.ErrNotFound
// }
