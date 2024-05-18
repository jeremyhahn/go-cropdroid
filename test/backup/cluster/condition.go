//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"

	logging "github.com/op/go-logging"
)

type RaftConditionDAO struct {
	logger  *logging.Logger
	raft    cluster.RaftNode
	farmDAO dao.FarmDAO
	dao.ConditionDAO
}

func NewRaftConditionDAO(logger *logging.Logger,
	raftNode cluster.RaftNode, farmDAO dao.FarmDAO) dao.ConditionDAO {
	return &RaftConditionDAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftConditionDAO) Save(farmID, deviceID uint64,
	condition *config.Condition) error {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	for _, device := range farmConfig.GetDevices() {
		if device.ID == deviceID {
			for _, channel := range device.GetChannels() {
				if channel.ID == condition.GetChannelID() {
					// if condition.GetWorkflowID() == 0 || condition.GetChannelID() == 0 {
					// 	idSetter := dao.raft.GetParams().IdSetter
					// 	idSetter.SetConditionIds(deviceID, []*config.Condition{condition})
					// }
					channel.SetCondition(condition)
					return dao.farmDAO.Save(farmConfig)
				}
			}
		}
	}
	return datastore.ErrNotFound
}

func (dao *RaftConditionDAO) Get(farmID, deviceID, channelID,
	conditionID uint64, CONSISTENCY_LEVEL int) (*config.Condition, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	for _, device := range farmConfig.GetDevices() {
		if device.ID == deviceID {
			for _, channel := range device.GetChannels() {
				if channel.ID == channelID {
					for _, condition := range channel.GetConditions() {
						if condition.ID == conditionID {
							return condition, nil
						}
					}
				}
			}
		}
	}
	return nil, datastore.ErrNotFound
}

func (dao *RaftConditionDAO) Delete(farmID, deviceID uint64, condition *config.Condition) error {
	dao.logger.Debugf(fmt.Sprintf("Deleting condition record: %+v", condition))
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	newConditionList := make([]*config.Condition, 0)
	for _, device := range farmConfig.GetDevices() {
		if device.ID == deviceID {
			for _, channel := range device.GetChannels() {
				if channel.ID == condition.GetChannelID() {
					for _, cond := range channel.GetConditions() {
						if condition.ID == cond.ID {
							continue
						}
						newConditionList = append(newConditionList, condition)
					}
					channel.SetConditions(newConditionList)
					device.SetChannel(channel)
					farmConfig.SetDevice(device)
					if err := dao.farmDAO.Save(farmConfig); err != nil {
						return err
					}
					return nil
				}
			}
		}
	}
	return datastore.ErrNotFound
}

func (dao *RaftConditionDAO) GetByChannelID(farmID, deviceID,
	channelID uint64, CONSISTENCY_LEVEL int) ([]*config.Condition, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, CONSISTENCY_LEVEL)
	if err != nil {
		return nil, err
	}
	device, err := farmConfig.GetDeviceById(deviceID)
	if err != nil {
		return nil, datastore.ErrNotFound
	}
	for _, channel := range device.GetChannels() {
		if channel.ID == channelID {
			return channel.GetConditions(), nil
		}
	}
	return nil, datastore.ErrNotFound
}
