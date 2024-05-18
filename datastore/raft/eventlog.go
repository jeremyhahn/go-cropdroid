//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/entity"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	logging "github.com/op/go-logging"
)

type RaftEventLogDAO interface {
	RaftDAO[*entity.EventLog]
	dao.EventLogDAO
	ClusterID() uint64
}

type RaftEventLog struct {
	logger *logging.Logger
	raft   cluster.RaftNode
	dao.EventLogDAO
	GenericRaftDAO[*entity.EventLog]
}

func NewRaftEventLogDAO(logger *logging.Logger, raftNode cluster.RaftNode, farmID uint64) RaftEventLogDAO {

	eventLogClusterID := raftNode.GetParams().
		IdGenerator.CreateEventLogClusterID(farmID)

	return &RaftEventLog{
		logger: logger,
		raft:   raftNode,
		GenericRaftDAO: GenericRaftDAO[*entity.EventLog]{
			logger:    logger,
			raft:      raftNode,
			clusterID: eventLogClusterID,
		}}
}

func (dao *RaftEventLog) ClusterID() uint64 {
	return dao.GenericRaftDAO.clusterID
}

func (dao *RaftEventLog) StartClusterNode(waitForClusterReady bool) error {
	return dao.GenericRaftDAO.StartClusterNode(waitForClusterReady)
}

func (dao *RaftEventLog) StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error {
	return dao.GenericRaftDAO.StartLocalCluster(localCluster, waitForClusterReady)
}

func (dao *RaftEventLog) WaitForClusterReady() {
	dao.GenericRaftDAO.WaitForClusterReady()
}

func (dao *RaftEventLog) Save(eventLog *entity.EventLog) error {
	return dao.GenericRaftDAO.Save(eventLog)
}

func (dao *RaftEventLog) Update(eventLog *entity.EventLog) error {
	return dao.GenericRaftDAO.Update(eventLog)
}

func (dao *RaftEventLog) Delete(eventLog *entity.EventLog) error {
	return dao.GenericRaftDAO.Delete(eventLog)
}

func (dao *RaftEventLog) Get(id uint64, CONSISTENCY_LEVEL int) (*entity.EventLog, error) {
	return dao.GenericRaftDAO.Get(id, CONSISTENCY_LEVEL)
}

func (dao *RaftEventLog) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*entity.EventLog], error) {
	return dao.GenericRaftDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
}

func (customerDAO *RaftEventLog) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*entity.EventLog], CONSISTENCY_LEVEL int) error {

	return customerDAO.GenericRaftDAO.ForEachPage(pageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
}

func (dao *RaftEventLog) Count(CONSISTENCY_LEVEL int) (int64, error) {
	return dao.GenericRaftDAO.Count(CONSISTENCY_LEVEL)
}
