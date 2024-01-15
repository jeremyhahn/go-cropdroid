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
	if channel.GetID() == 0 {
		key := fmt.Sprintf("%d-%s", farmID, channel.GetName())
		id := dao.raft.GetParams().IdGenerator.NewID(key)
		channel.SetID(id)
	}
	devices := farmConfig.GetDevices()
	for _, device := range devices {
		if device.GetID() == channel.GetDeviceID() {
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
