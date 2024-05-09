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

type RaftDeviceSettingDAO struct {
	logger    *logging.Logger
	raft      RaftNode
	deviceDAO dao.DeviceDAO
	dao.DeviceSettingDAO
}

func NewRaftDeviceSettingDAO(logger *logging.Logger,
	raftNode RaftNode,
	deviceDAO dao.DeviceDAO) dao.DeviceSettingDAO {

	return &RaftDeviceSettingDAO{
		logger:    logger,
		raft:      raftNode,
		deviceDAO: deviceDAO}
}

func (dao *RaftDeviceSettingDAO) Save(farmID uint64,
	setting *config.DeviceSetting) error {

	if setting.GetID() == 0 {
		key := fmt.Sprintf("%d-%d-%s", farmID, setting.GetDeviceID(),
			setting.GetKey())
		id := dao.raft.GetParams().IdGenerator.NewStringID(key)
		setting.SetID(id)
	}
	dao.logger.Debugf("Saving farm %d device setting %+v",
		farmID, setting)

	device, err := dao.deviceDAO.Get(farmID,
		setting.GetDeviceID(), common.CONSISTENCY_LOCAL)
	if err != nil {
		return datastore.ErrNotFound
	}
	device.SetSetting(setting)
	return dao.deviceDAO.Save(device)
}

func (dao *RaftDeviceSettingDAO) Get(farmID, deviceID uint64,
	name string, CONSISTENCY_LEVEL int) (*config.DeviceSetting, error) {

	device, err := dao.deviceDAO.Get(farmID, deviceID,
		common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, datastore.ErrNotFound
	}
	for _, setting := range device.GetSettings() {
		if setting.GetKey() == name {
			return setting, nil
		}
	}
	return nil, datastore.ErrNotFound
}

// type RaftDeviceSettingDAO struct {
// 	logger *logging.Logger
// 	raft   RaftNode
// 	dao.DeviceDAO
// }

// func NewRaftDeviceSettingDAO(logger *logging.Logger,
// 	raftNode RaftNode, clusterID uint64) dao.DeviceDAO {

// 	return &RaftDeviceSettingDAO{
// 		logger: logger,
// 		raft:   raftNode}
// }

// func (dao *RaftDeviceSettingDAO) Save(device config.DeviceSetting) error {
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

// func (dao *RaftDeviceSettingDAO) Get(id uint64, CONSISTENCY_LEVEL int) (config.DeviceSetting, error) {
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
// 		records := result.([]config.DeviceSetting)
// 		if len(records) > 0 {
// 			return records[0], nil
// 		}
// 		return nil, nil
// 	}
// 	return nil, datastore.ErrNotFound
// }
