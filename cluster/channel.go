//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"

	logging "github.com/op/go-logging"
)

type RaftChannelDAO struct {
	logger  *logging.Logger
	raft    RaftNode
	farmDAO dao.FarmDAO
	dao.ChannelDAO
}

func NewRaftChannelDAO(logger *logging.Logger,
	raftNode RaftNode, farmDAO dao.FarmDAO) dao.ChannelDAO {
	return &RaftChannelDAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftChannelDAO) Save(farmID uint64, channel *config.Channel) error {
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	devices := farmConfig.GetDevices()
	for _, device := range devices {
		deviceID := device.GetID()
		if deviceID == channel.GetDeviceID() {
			// if channel.GetID() == 0 || channel.GetDeviceID() == 0 || channel.GetChannelID() == 0 {
			// 	idSetter := dao.raft.GetParams().IdSetter
			// 	idSetter.SetChannelsIds(farmID, deviceID, []*config.Channel{channel})
			// }
			device.SetChannel(channel)
			farmConfig.SetDevice(device)
			return dao.farmDAO.Save(farmConfig)
		}
	}
	return datastore.ErrNotFound
}

func (dao *RaftChannelDAO) Get(orgID, farmID, channelID uint64,
	CONSISTENCY_LEVEL int) (*config.Channel, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	for _, channel := range farmConfig.GetDevices() {
		for _, channel := range channel.GetChannels() {
			if channel.GetID() == channelID {
				return channel, nil
			}
		}
	}
	return nil, datastore.ErrNotFound
}

func (dao *RaftChannelDAO) GetByDevice(orgID, farmID, channelID uint64,
	CONSISTENCY_LEVEL int) ([]*config.Channel, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	channel, err := farmConfig.GetDeviceById(channelID)
	if err != nil {
		return nil, datastore.ErrNotFound
	}
	return channel.GetChannels(), nil
}
