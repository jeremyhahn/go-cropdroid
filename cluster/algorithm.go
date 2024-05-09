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

type RaftAlgorithmDAO struct {
	logger    *logging.Logger
	raft      RaftNode
	clusterID uint64
	dao.AlgorithmDAO
}

func NewRaftAlgorithmDAO(logger *logging.Logger, raftNode RaftNode,
	clusterID uint64) dao.AlgorithmDAO {

	return &RaftAlgorithmDAO{
		logger:    logger,
		raft:      raftNode,
		clusterID: clusterID}
}

func (algorithmDAO *RaftAlgorithmDAO) StartCluster() {
	params := algorithmDAO.raft.GetParams()
	clusterID := params.RaftOptions.AlgorithmClusterID
	nodeID := params.GetNodeID()
	join := false
	sm := statemachine.NewAlgorithmConfigMachine(algorithmDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, nodeID)
	err := algorithmDAO.raft.CreateOnDiskCluster(clusterID, join, sm.CreateAlgorithmConfigMachine)
	if err != nil {
		algorithmDAO.logger.Fatal(err)
	}
}

func (algorithmDAO *RaftAlgorithmDAO) Get(algorithmID uint64, CONSISTENCY_LEVEL int) (*config.Algorithm, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = algorithmDAO.raft.ReadLocal(algorithmDAO.clusterID, algorithmID)
		if err != nil {
			algorithmDAO.logger.Errorf("Error (algorithmID=%d): %s", algorithmID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = algorithmDAO.raft.SyncRead(algorithmDAO.clusterID, algorithmID)
		if err != nil {
			algorithmDAO.logger.Errorf("Error (algorithmID=%d): %s", algorithmID, err)
			return nil, err
		}
	}
	if result != nil {
		return result.(*config.Algorithm), nil
	}
	return nil, datastore.ErrNotFound
}

func (algorithmDAO *RaftAlgorithmDAO) GetAll(CONSISTENCY_LEVEL int) ([]*config.Algorithm, error) {
	algorithmDAO.logger.Debugf("Fetching supported algorithms")
	var result interface{}
	var err error
	query := []byte(fmt.Sprint(statemachine.QUERY_TYPE_WILDCARD))
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = algorithmDAO.raft.ReadLocal(algorithmDAO.clusterID, query)
		if err != nil {
			algorithmDAO.logger.Errorf("Error (algorithmID=%d): %s", algorithmDAO.clusterID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = algorithmDAO.raft.SyncRead(algorithmDAO.clusterID, query)
		if err != nil {
			algorithmDAO.logger.Errorf("Error (algorithmID=%d): %s", algorithmDAO.clusterID, err)
			return nil, err
		}
	}
	switch v := result.(type) {
	case []*config.Algorithm:
		return v, nil
	default:
		errmsg := fmt.Sprintf("unexpected query type %T", v)
		algorithmDAO.logger.Error(errmsg)
		return nil, errors.New(errmsg)
	}
}

func (algorithmDAO *RaftAlgorithmDAO) Save(algorithm *config.Algorithm) error {

	if algorithm.ID == 0 {
		id := algorithmDAO.raft.GetParams().IdGenerator.NewStringID(algorithm.Name)
		algorithm.ID = id
	}

	algorithmDAO.logger.Debugf("Saving algorithm: %+v", algorithm)

	perm, err := json.Marshal(algorithm)
	if err != nil {
		algorithmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		algorithmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := algorithmDAO.raft.SyncPropose(algorithmDAO.clusterID, proposal); err != nil {
		algorithmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

func (algorithmDAO *RaftAlgorithmDAO) Delete(algorithm *config.Algorithm) error {
	algorithmDAO.logger.Debugf(fmt.Sprintf("Deleting algorithm record: %+v", algorithm))
	perm, err := json.Marshal(&algorithm)
	if err != nil {
		algorithmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, perm).Serialize()
	if err != nil {
		algorithmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := algorithmDAO.raft.SyncPropose(algorithmDAO.clusterID, proposal); err != nil {
		algorithmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}
