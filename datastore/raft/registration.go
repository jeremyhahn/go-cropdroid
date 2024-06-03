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
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/statemachine"
	logging "github.com/op/go-logging"
)

type RaftRegistrationDAO struct {
	logger    *logging.Logger
	raft      cluster.RaftNode
	clusterID uint64
	dao.RegistrationDAO
}

func NewRaftRegistrationDAO(logger *logging.Logger, raftNode cluster.RaftNode,
	clusterID uint64) dao.RegistrationDAO {

	return &RaftRegistrationDAO{
		logger:    logger,
		raft:      raftNode,
		clusterID: clusterID}
}

func (registrationDAO *RaftRegistrationDAO) StartCluster() {
	params := registrationDAO.raft.GetParams()
	clusterID := params.RaftOptions.RegistrationClusterID
	sm := statemachine.NewGenericOnDiskStateMachine[*config.Registration](registrationDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, params.NodeID)
	err := registrationDAO.raft.CreateOnDiskCluster(clusterID, params.Join, sm.CreateOnDiskStateMachine)
	if err != nil {
		registrationDAO.logger.Fatal(err)
	}
	registrationDAO.raft.WaitForClusterReady(clusterID)
}

func (registrationDAO *RaftRegistrationDAO) Get(registrationID uint64,
	CONSISTENCY_LEVEL int) (*config.RegistrationStruct, error) {

	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = registrationDAO.raft.ReadLocal(registrationDAO.clusterID, registrationID)
		if err != nil {
			registrationDAO.logger.Errorf("Error (registrationID=%d): %s", registrationID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = registrationDAO.raft.SyncRead(registrationDAO.clusterID, registrationID)
		if err != nil {
			registrationDAO.logger.Errorf("Error (registrationID=%d): %s", registrationID, err)
			return nil, err
		}
	}
	if result != nil {
		return result.(*config.RegistrationStruct), nil
	}
	return nil, datastore.ErrRecordNotFound
}

func (registrationDAO *RaftRegistrationDAO) Save(registration *config.RegistrationStruct) error {

	if registration.ID == 0 {
		id := registrationDAO.raft.GetParams().IdGenerator.NewStringID(registration.GetEmail())
		registration.SetID(id)
	}

	registrationDAO.logger.Debugf("Saving registration: %+v", registration)

	perm, err := json.Marshal(registration)
	if err != nil {
		registrationDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		registrationDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := registrationDAO.raft.SyncPropose(registrationDAO.clusterID, proposal); err != nil {
		registrationDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

func (registrationDAO *RaftRegistrationDAO) Delete(registration *config.RegistrationStruct) error {
	registrationDAO.logger.Debugf(fmt.Sprintf("Deleting registration record: %+v", registration))
	perm, err := json.Marshal(registration)
	if err != nil {
		registrationDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, perm).Serialize()
	if err != nil {
		registrationDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := registrationDAO.raft.SyncPropose(registrationDAO.clusterID, proposal); err != nil {
		registrationDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}
