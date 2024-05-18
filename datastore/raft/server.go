//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	logging "github.com/op/go-logging"
)

type RaftServer interface {
	RaftDAO[*config.Server]
	dao.ServerDAO
}

type RaftServerDAO struct {
	logger *logging.Logger
	raft   cluster.RaftNode
	GenericRaftDAO[*config.Server]
	RaftServer
}

func NewRaftServerDAO(logger *logging.Logger, raftNode cluster.RaftNode,
	clusterID uint64) RaftServer {

	return &RaftServerDAO{
		logger: logger,
		raft:   raftNode,
		GenericRaftDAO: GenericRaftDAO[*config.Server]{
			logger:    logger,
			raft:      raftNode,
			clusterID: clusterID}}
}

func (serverDAO *RaftServerDAO) StartClusterNode(waitForClusterReady bool) error {
	return serverDAO.GenericRaftDAO.StartClusterNode(waitForClusterReady)
}

func (serverDAO *RaftServerDAO) StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error {
	return serverDAO.GenericRaftDAO.StartLocalCluster(localCluster, waitForClusterReady)
}

func (serverDAO *RaftServerDAO) WaitForClusterReady() {
	serverDAO.GenericRaftDAO.WaitForClusterReady()
}

func (serverDAO *RaftServerDAO) Save(Server *config.Server) error {
	return serverDAO.GenericRaftDAO.Save(Server)
}

func (serverDAO *RaftServerDAO) SaveWithTimeSeriesIndex(Server *config.Server) error {
	return serverDAO.GenericRaftDAO.SaveWithTimeSeriesIndex(Server)
}

func (serverDAO *RaftServerDAO) Update(Server *config.Server) error {
	return serverDAO.GenericRaftDAO.Update(Server)
}

func (serverDAO *RaftServerDAO) Delete(Server *config.Server) error {
	return serverDAO.GenericRaftDAO.Delete(Server)
}

func (serverDAO *RaftServerDAO) Get(id uint64, CONSISTENCY_LEVEL int) (*config.Server, error) {
	return serverDAO.GenericRaftDAO.Get(id, CONSISTENCY_LEVEL)
}

func (serverDAO *RaftServerDAO) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.Server], error) {
	return serverDAO.GenericRaftDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
}

func (serverDAO *RaftServerDAO) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*config.Server], CONSISTENCY_LEVEL int) error {

	return serverDAO.GenericRaftDAO.ForEachPage(pageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
}

func (serverDAO *RaftServerDAO) Count(CONSISTENCY_LEVEL int) (int64, error) {
	return serverDAO.GenericRaftDAO.Count(CONSISTENCY_LEVEL)
}
