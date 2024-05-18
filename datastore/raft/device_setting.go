//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"

	logging "github.com/op/go-logging"
)

type RaftDeviceSettingDAO struct {
	logger    *logging.Logger
	raft      cluster.RaftNode
	deviceDAO dao.DeviceDAO
	dao.DeviceSettingDAO
}

func NewRaftDeviceSettingDAO(logger *logging.Logger,
	raftNode cluster.RaftNode, deviceDAO dao.DeviceDAO) dao.DeviceSettingDAO {

	return &RaftDeviceSettingDAO{
		logger:    logger,
		raft:      raftNode,
		deviceDAO: deviceDAO}
}

func (dao *RaftDeviceSettingDAO) Save(farmID uint64,
	setting *config.DeviceSetting) error {

	if setting.ID == 0 {
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
