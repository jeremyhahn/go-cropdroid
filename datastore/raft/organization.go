//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	logging "github.com/op/go-logging"
)

type RaftOrganizationDAO interface {
	RaftDAO[*config.OrganizationStruct]
	dao.OrganizationDAO
}

type RaftOrganization struct {
	logger    *logging.Logger
	raft      cluster.RaftNode
	serverDAO dao.ServerDAO
	GenericRaftDAO[*config.OrganizationStruct]
	RaftOrganizationDAO
}

func NewRaftOrganizationDAO(logger *logging.Logger, raftNode cluster.RaftNode,
	clusterID uint64, serverDAO dao.ServerDAO) RaftOrganizationDAO {

	return &RaftOrganization{
		logger: logger,
		raft:   raftNode,
		GenericRaftDAO: GenericRaftDAO[*config.OrganizationStruct]{
			logger:    logger,
			raft:      raftNode,
			clusterID: clusterID,
		},
		serverDAO: serverDAO}
}

func (organizationDAO *RaftOrganization) StartClusterNode(waitForClusterReady bool) error {
	return organizationDAO.GenericRaftDAO.StartClusterNode(waitForClusterReady)
}

func (organizationDAO *RaftOrganization) StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error {
	return organizationDAO.GenericRaftDAO.StartLocalCluster(localCluster, waitForClusterReady)
}

func (organizationDAO *RaftOrganization) WaitForClusterReady() {
	organizationDAO.GenericRaftDAO.WaitForClusterReady()
}

func (organizationDAO *RaftOrganization) Save(Organization *config.OrganizationStruct) error {
	return organizationDAO.GenericRaftDAO.Save(Organization)
}

func (organizationDAO *RaftOrganization) SaveWithTimeSeriesIndex(Organization *config.OrganizationStruct) error {
	return organizationDAO.GenericRaftDAO.SaveWithTimeSeriesIndex(Organization)
}

func (organizationDAO *RaftOrganization) Update(Organization *config.OrganizationStruct) error {
	return organizationDAO.GenericRaftDAO.Update(Organization)
}

func (organizationDAO *RaftOrganization) Delete(Organization *config.OrganizationStruct) error {
	return organizationDAO.GenericRaftDAO.Delete(Organization)
}

func (organizationDAO *RaftOrganization) Get(id uint64, CONSISTENCY_LEVEL int) (*config.OrganizationStruct, error) {
	return organizationDAO.GenericRaftDAO.Get(id, CONSISTENCY_LEVEL)
}

func (organizationDAO *RaftOrganization) GetPage(pageQuery query.PageQuery,
	CONSISTENCY_LEVEL int) (dao.PageResult[*config.OrganizationStruct], error) {

	return organizationDAO.GenericRaftDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
}

func (organizationDAO *RaftOrganization) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*config.OrganizationStruct], CONSISTENCY_LEVEL int) error {

	return organizationDAO.GenericRaftDAO.ForEachPage(pageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
}

func (organizationDAO *RaftOrganization) Count(CONSISTENCY_LEVEL int) (int64, error) {
	return organizationDAO.Count(CONSISTENCY_LEVEL)
}

func (organizationDAO *RaftOrganization) GetUsers(id uint64) ([]*config.UserStruct, error) {
	users := make([]*config.UserStruct, 0)
	return users, errors.New("GetUsers method not implemented")
}
