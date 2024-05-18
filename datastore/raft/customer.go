//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"encoding/json"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/statemachine"
	logging "github.com/op/go-logging"
)

type RaftCustomerDAO interface {
	RaftDAO[*config.Customer]
	dao.CustomerDAO
}

type RaftCustomer struct {
	logger *logging.Logger
	raft   cluster.RaftNode
	GenericRaftDAO[*config.Customer]
	RaftCustomerDAO
}

func NewRaftCustomerDAO(logger *logging.Logger, raftNode cluster.RaftNode,
	clusterID uint64) RaftCustomerDAO {

	return &RaftCustomer{
		logger: logger,
		raft:   raftNode,
		GenericRaftDAO: GenericRaftDAO[*config.Customer]{
			logger:    logger,
			raft:      raftNode,
			clusterID: clusterID,
		}}
}

func (customerDAO *RaftCustomer) StartCluster() {
	params := customerDAO.raft.GetParams()
	clusterID := params.RaftOptions.CustomerClusterID
	sm := statemachine.NewGenericOnDiskStateMachine[*config.Customer](customerDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, params.NodeID)
	err := customerDAO.raft.CreateOnDiskCluster(clusterID, params.Join, sm.CreateOnDiskStateMachine)
	if err != nil {
		customerDAO.logger.Fatal(err)
	}
	//customerDAO.raft.WaitForClusterReady(clusterID)
}

func (customerDAO *RaftCustomer) StartClusterNode(waitForClusterReady bool) error {
	return customerDAO.GenericRaftDAO.StartClusterNode(waitForClusterReady)
}

func (customerDAO *RaftCustomer) StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error {
	return customerDAO.GenericRaftDAO.StartLocalCluster(localCluster, waitForClusterReady)
}

func (customerDAO *RaftCustomer) WaitForClusterReady() {
	customerDAO.GenericRaftDAO.WaitForClusterReady()
}

func (customerDAO *RaftCustomer) SaveWithTimeSeriesIndex(CustomerConfig *config.Customer) error {
	return customerDAO.GenericRaftDAO.SaveWithTimeSeriesIndex(CustomerConfig)
}

func (customerDAO *RaftCustomer) Update(CustomerConfig *config.Customer) error {
	return customerDAO.GenericRaftDAO.Update(CustomerConfig)
}

func (customerDAO *RaftCustomer) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.Customer], error) {
	return customerDAO.GenericRaftDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
}

func (customerDAO *RaftCustomer) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*config.Customer], CONSISTENCY_LEVEL int) error {

	pageResult, err := customerDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
	if err != nil {
		return nil
	}
	if err = pagerProcFunc(pageResult.Entities); err != nil {
		return err
	}
	if pageResult.HasMore {
		nextPageQuery := query.PageQuery{
			Page:      pageQuery.Page + 1,
			PageSize:  pageQuery.PageSize,
			SortOrder: pageQuery.SortOrder}
		return customerDAO.ForEachPage(nextPageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
	}
	return nil
}

func (customerDAO *RaftCustomer) Count(CONSISTENCY_LEVEL int) (int64, error) {
	return customerDAO.Count(CONSISTENCY_LEVEL)
}

func (customerDAO *RaftCustomer) Get(customerID uint64, CONSISTENCY_LEVEL int) (*config.Customer, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = customerDAO.raft.ReadLocal(customerDAO.GenericRaftDAO.clusterID, customerID)
		if err != nil {
			customerDAO.logger.Errorf("Error (customerClusterID=%d, customerID=%d): %s",
				customerDAO.GenericRaftDAO.clusterID, customerID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = customerDAO.raft.SyncRead(customerDAO.GenericRaftDAO.clusterID, customerID)
		if err != nil {
			customerDAO.logger.Errorf("Error (customerClusterID=%d, customerID=%d): %s",
				customerDAO.GenericRaftDAO.clusterID, customerID, err)
			return nil, err
		}
	}
	if result != nil {
		return result.(*config.Customer), nil
	}
	return nil, datastore.ErrNotFound
}

func (customerDAO *RaftCustomer) GetByEmail(customerEmail string, CONSISTENCY_LEVEL int) (*config.Customer, error) {
	customerID := customerDAO.raft.GetParams().IdGenerator.NewStringID(customerEmail)
	return customerDAO.Get(customerID, CONSISTENCY_LEVEL)
}

func (customerDAO *RaftCustomer) Save(customer *config.Customer) error {

	if customer.ID == 0 {
		idSetter := customerDAO.raft.GetParams().IdSetter
		idSetter.SetCustomerIds(customer)
	}

	customerDAO.logger.Debugf("Saving customer: %+v", customer)

	perm, err := json.Marshal(customer)
	if err != nil {
		customerDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		customerDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := customerDAO.raft.SyncPropose(customerDAO.GenericRaftDAO.clusterID, proposal); err != nil {
		customerDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

func (customerDAO *RaftCustomer) Delete(customer *config.Customer) error {
	customerDAO.logger.Debugf(fmt.Sprintf("Deleting customer record: %+v", customer))
	perm, err := json.Marshal(customer)
	if err != nil {
		customerDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, perm).Serialize()
	if err != nil {
		customerDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := customerDAO.raft.SyncPropose(customerDAO.GenericRaftDAO.clusterID, proposal); err != nil {
		customerDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}
