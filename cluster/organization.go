//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"

	logging "github.com/op/go-logging"
)

type RaftOrganizationDAO struct {
	logger    *logging.Logger
	raft      RaftNode
	clusterID uint64
	serverDAO ServerDAO
	dao.OrganizationDAO
}

func NewRaftOrganizationDAO(logger *logging.Logger,
	raftNode RaftNode, clusterID uint64,
	serverDAO ServerDAO) dao.OrganizationDAO {

	return &RaftOrganizationDAO{
		logger:    logger,
		raft:      raftNode,
		clusterID: clusterID,
		serverDAO: serverDAO}
}

func (dao *RaftOrganizationDAO) StartCluster() {
	params := dao.raft.GetParams()
	clusterID := params.RaftOptions.OrganizationClusterID
	sm := statemachine.NewOrganizationConfigMachine(dao.logger,
		params.IdGenerator, params.DataDir, clusterID, params.NodeID)
	err := dao.raft.CreateOnDiskCluster(clusterID, params.Join, sm.CreateOrganizationConfigMachine)
	if err != nil {
		dao.logger.Fatal(err)
	}
	dao.raft.WaitForClusterReady(clusterID)
}

func (dao *RaftOrganizationDAO) Save(organization *config.Organization) error {
	if organization.GetID() == 0 {
		id := dao.raft.GetParams().IdGenerator.NewID(organization.GetName())
		organization.SetID(id)
	}
	dao.logger.Debugf("Saving organization: %+v", organization)
	org, err := json.Marshal(organization)
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, org).Serialize()
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}
	if err := dao.raft.SyncPropose(dao.clusterID, proposal); err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}
	// Update server config organization refs
	serverConfig, err := dao.serverDAO.GetConfig(common.CONSISTENCY_LOCAL)
	if errors.Is(err, datastore.ErrNotFound) {
		serverConfig = config.NewServer()
	} else if err != nil {
		return err
	}
	serverConfig.AddOrganizationRef(organization.GetID())
	if err := dao.serverDAO.Save(serverConfig); err != nil {
		return err
	}
	return nil
}

func (dao *RaftOrganizationDAO) Get(id uint64, CONSISTENCY_LEVEL int) (*config.Organization, error) {
	dao.logger.Debugf("Fetching organization ID: %d", id)
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = dao.raft.ReadLocal(dao.clusterID, id)
		if err != nil {
			dao.logger.Errorf("Error (orgID=%d): %s", id, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = dao.raft.SyncRead(dao.clusterID, id)
		if err != nil {
			dao.logger.Errorf("Error (orgID=%d): %s", id, err)
			return nil, err
		}
	}
	if result != nil {
		return result.(*config.Organization), nil
	}
	return nil, datastore.ErrNotFound
}

func (dao *RaftOrganizationDAO) GetAll(CONSISTENCY_LEVEL int) ([]*config.Organization, error) {
	dao.logger.Debugf("Fetching all organizations")
	var result interface{}
	var err error
	query := []byte(fmt.Sprint(statemachine.QUERY_TYPE_WILDCARD))
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = dao.raft.ReadLocal(dao.clusterID, query)
		if err != nil {
			dao.logger.Errorf("Error (orgID=%d): %s", dao.clusterID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = dao.raft.SyncRead(dao.clusterID, query)
		if err != nil {
			dao.logger.Errorf("Error (orgID=%d): %s", dao.clusterID, err)
			return nil, err
		}
	}

	// Lookup method always returns an array
	// Lookup(nil) = get current state
	// Lookup("*") = get state history
	// Lookup("start:end") = get slice ranging from "start" to "end" (0 based index)

	switch v := result.(type) {
	case []*config.Organization:
		return v, nil
	default:
		dao.logger.Errorf("unexpected query type %T", v)
		return []*config.Organization{}, nil
	}
}

func (dao *RaftOrganizationDAO) Delete(organization *config.Organization) error {
	dao.logger.Debugf("Deleting organization: %+v", organization)
	org, err := json.Marshal(organization)
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, org).Serialize()
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}
	if err := dao.raft.SyncPropose(dao.clusterID, proposal); err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}
	// Delete server config organization refs
	serverConfig, err := dao.serverDAO.GetConfig(common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	serverConfig.RemoveOrganizationRef(organization.GetID())
	if err := dao.serverDAO.Save(serverConfig); err != nil {
		return err
	}
	return nil
}

// This method returns a minimal depth organization with its associated farms and users.
// No device or workflows are returned.
// func (dao *RaftOrganizationDAO) GetByUserID(userID uint64, shallow bool) ([]config.OrganizationConfig, error) {
// 	//	dao.logger.Debugf("Getting organizations for user.id %d", userID)
// }

// func (dao *RaftOrganizationDAO) GetUsers(id uint64) ([]config.UserConfig, error) {
// 	//	dao.logger.Debugf("Fetching users for organization ID: %d", id)
// }
