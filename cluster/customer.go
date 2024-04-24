//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	logging "github.com/op/go-logging"
)

type RaftCustomerDAO struct {
	logger    *logging.Logger
	raft      RaftNode
	clusterID uint64
	dao.CustomerDAO
}

func NewRaftCustomerDAO(logger *logging.Logger, raftNode RaftNode,
	clusterID uint64) dao.CustomerDAO {

	return &RaftCustomerDAO{
		logger:    logger,
		raft:      raftNode,
		clusterID: clusterID}
}

func (customerDAO *RaftCustomerDAO) StartCluster() {
	params := customerDAO.raft.GetParams()
	clusterID := params.RaftOptions.CustomerClusterID
	sm := statemachine.NewCustomerConfigMachine(customerDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, params.NodeID)
	err := customerDAO.raft.CreateOnDiskCluster(clusterID, params.Join, sm.CreateCustomerConfigMachine)
	if err != nil {
		customerDAO.logger.Fatal(err)
	}
	//customerDAO.raft.WaitForClusterReady(clusterID)
}

func (customerDAO *RaftCustomerDAO) Get(customerID uint64, CONSISTENCY_LEVEL int) (*config.Customer, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = customerDAO.raft.ReadLocal(customerDAO.clusterID, customerID)
		if err != nil {
			customerDAO.logger.Errorf("Error (customerClusterID=%d, customerID=%d): %s",
				customerDAO.clusterID, customerID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = customerDAO.raft.SyncRead(customerDAO.clusterID, customerID)
		if err != nil {
			customerDAO.logger.Errorf("Error (customerClusterID=%d, customerID=%d): %s",
				customerDAO.clusterID, customerID, err)
			return nil, err
		}
	}
	if result != nil {
		return result.(*config.Customer), nil
	}
	return nil, datastore.ErrNotFound
}

func (customerDAO *RaftCustomerDAO) GetByEmail(customerName string, CONSISTENCY_LEVEL int) (*config.Customer, error) {
	customerID := customerDAO.raft.GetParams().IdGenerator.NewID(customerName)
	return customerDAO.Get(customerID, CONSISTENCY_LEVEL)
}

func (customerDAO *RaftCustomerDAO) Save(customer *config.Customer) error {

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
	if err := customerDAO.raft.SyncPropose(customerDAO.clusterID, proposal); err != nil {
		customerDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

func (customerDAO *RaftCustomerDAO) Delete(customer *config.Customer) error {
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
	if err := customerDAO.raft.SyncPropose(customerDAO.clusterID, proposal); err != nil {
		customerDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}
