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

type RaftRoleDAO interface {
	RaftDAO[*config.RoleStruct]
	dao.RoleDAO
}

type RaftRole struct {
	logger *logging.Logger
	raft   cluster.RaftNode
	GenericRaftDAO[*config.RoleStruct]
}

func NewRaftRoleDAO(logger *logging.Logger, raftNode cluster.RaftNode,
	clusterID uint64) RaftRoleDAO {

	return &RaftRole{
		logger: logger,
		raft:   raftNode,
		GenericRaftDAO: GenericRaftDAO[*config.RoleStruct]{
			logger:    logger,
			raft:      raftNode,
			clusterID: clusterID}}
}

func (roleDAO *RaftRole) StartClusterNode(waitForClusterReady bool) error {
	return roleDAO.GenericRaftDAO.StartClusterNode(waitForClusterReady)
}

func (roleDAO *RaftRole) StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error {
	return roleDAO.GenericRaftDAO.StartLocalCluster(localCluster, waitForClusterReady)
}

func (roleDAO *RaftRole) WaitForClusterReady() {
	roleDAO.GenericRaftDAO.WaitForClusterReady()
}

func (roleDAO *RaftRole) Save(Role *config.RoleStruct) error {
	return roleDAO.GenericRaftDAO.Save(Role)
}

func (roleDAO *RaftRole) SaveWithTimeSeriesIndex(Role *config.RoleStruct) error {
	return roleDAO.GenericRaftDAO.SaveWithTimeSeriesIndex(Role)
}

func (roleDAO *RaftRole) Update(Role *config.RoleStruct) error {
	return roleDAO.GenericRaftDAO.Update(Role)
}

func (roleDAO *RaftRole) Delete(Role *config.RoleStruct) error {
	return roleDAO.GenericRaftDAO.Delete(Role)
}

func (roleDAO *RaftRole) Get(id uint64, CONSISTENCY_LEVEL int) (*config.RoleStruct, error) {
	return roleDAO.GenericRaftDAO.Get(id, CONSISTENCY_LEVEL)
}

func (roleDAO *RaftRole) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.RoleStruct], error) {
	return roleDAO.GenericRaftDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
}

func (roleDAO *RaftRole) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*config.RoleStruct], CONSISTENCY_LEVEL int) error {

	return roleDAO.GenericRaftDAO.ForEachPage(pageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
}

func (roleDAO *RaftRole) GetByName(roleName string, CONSISTENCY_LEVEL int) (*config.RoleStruct, error) {
	roleID := roleDAO.raft.GetParams().IdGenerator.NewStringID(roleName)
	return roleDAO.Get(roleID, CONSISTENCY_LEVEL)
}
