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

type RaftScheduleDAO struct {
	logger  *logging.Logger
	raft    RaftNode
	farmDAO dao.FarmDAO
	dao.ScheduleDAO
}

func NewRaftScheduleDAO(logger *logging.Logger,
	raftNode RaftNode, farmDAO dao.FarmDAO) dao.ScheduleDAO {
	return &RaftScheduleDAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftScheduleDAO) Save(farmID, deviceID uint64,
	schedule *config.Schedule) error {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == deviceID {
			for _, channel := range device.GetChannels() {
				// if channel.GetID() == schedule.GetChannelID() {
				// 	if schedule.GetID() == 0 || schedule.GetWorkflowID() == 0 && schedule.GetChannelID() == 0 {
				// 		idSetter := dao.raft.GetParams().IdSetter
				// 		idSetter.SetScheduleIds(farmID, deviceID, []*config.Schedule{schedule})
				// 	}
				// 	channel.SetScheduleItem(schedule)
				// 	return dao.farmDAO.Save(farmConfig)
				// }
				channel.SetScheduleItem(schedule)
				return dao.farmDAO.Save(&farmConfig)
			}
		}
	}
	return datastore.ErrNotFound
}

func (dao *RaftScheduleDAO) Get(farmID, deviceID, channelID,
	scheduleID uint64, CONSISTENCY_LEVEL int) (*config.Schedule, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == deviceID {
			for _, channel := range device.GetChannels() {
				if channel.GetID() == channelID {
					for _, schedule := range channel.GetSchedule() {
						if schedule.GetID() == scheduleID {
							return schedule, nil
						}
					}
				}
			}
		}
	}
	return nil, datastore.ErrNotFound
}

func (dao *RaftScheduleDAO) Delete(farmID, deviceID uint64, schedule *config.Schedule) error {
	dao.logger.Debugf(fmt.Sprintf("Deleting schedule record: %+v", schedule))
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	newScheduleList := make([]*config.Schedule, 0)
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == deviceID {
			for _, channel := range device.GetChannels() {
				if channel.GetID() == schedule.GetChannelID() {
					for _, cond := range channel.GetSchedule() {
						if schedule.GetID() == cond.GetID() {
							continue
						}
						newScheduleList = append(newScheduleList, schedule)
					}
					if len(newScheduleList) > 0 {
						channel.SetSchedule(newScheduleList)
						device.SetChannel(channel)
						farmConfig.SetDevice(device)
						if err := dao.farmDAO.Save(&farmConfig); err != nil {
							return err
						}
						return nil
					}
				}
			}
		}
	}
	return datastore.ErrNotFound
}

func (dao *RaftScheduleDAO) GetByChannelID(farmID, deviceID,
	channelID uint64, CONSISTENCY_LEVEL int) ([]*config.Schedule, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	device, err := farmConfig.GetDeviceById(deviceID)
	if err != nil {
		return nil, datastore.ErrNotFound
	}
	for _, channel := range device.GetChannels() {
		if channel.GetID() == channelID {
			return channel.GetSchedule(), nil
		}
	}
	return nil, datastore.ErrNotFound
}
