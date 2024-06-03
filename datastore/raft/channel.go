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

type RaftChannelDAO struct {
	logger  *logging.Logger
	raft    cluster.RaftNode
	farmDAO dao.FarmDAO
	dao.ChannelDAO
}

func NewRaftChannelDAO(logger *logging.Logger,
	raftNode cluster.RaftNode, farmDAO dao.FarmDAO) dao.ChannelDAO {

	return &RaftChannelDAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftChannelDAO) Save(farmID uint64, channel *config.ChannelStruct) error {
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	devices := farmConfig.GetDevices()
	for _, device := range devices {
		deviceID := device.ID
		if deviceID == channel.GetDeviceID() {
			// if channel.ID == 0 || channel.GetDeviceID() == 0 || channel.GetChannelID() == 0 {
			// 	idSetter := dao.raft.GetParams().IdSetter
			// 	idSetter.SetChannelsIds(farmID, deviceID, []*config.Channel{channel})
			// }
			device.SetChannel(channel)
			farmConfig.SetDevice(device)
			return dao.farmDAO.Save(farmConfig)
		}
	}
	return datastore.ErrRecordNotFound
}

func (dao *RaftChannelDAO) Get(orgID, farmID, channelID uint64,
	CONSISTENCY_LEVEL int) (*config.ChannelStruct, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	for _, channel := range farmConfig.GetDevices() {
		for _, channel := range channel.GetChannels() {
			if channel.ID == channelID {
				return channel, nil
			}
		}
	}
	return nil, datastore.ErrRecordNotFound
}

func (dao *RaftChannelDAO) GetByDevice(orgID, farmID, channelID uint64,
	CONSISTENCY_LEVEL int) ([]*config.ChannelStruct, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	channel, err := farmConfig.GetDeviceById(channelID)
	if err != nil {
		return nil, datastore.ErrRecordNotFound
	}
	return channel.GetChannels(), nil
}
