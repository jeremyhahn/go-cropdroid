//go:build cluster && pebble
// +build cluster,pebble

package cluster

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

type RaftAlgorithmDAO struct {
	logger    *logging.Logger
	raft      cluster.RaftNode
	clusterID uint64
	dao.AlgorithmDAO
}

func NewRaftAlgorithmDAO(logger *logging.Logger, raftNode cluster.RaftNode,
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
	sm := statemachine.NewGenericOnDiskStateMachine[*config.Algorithm](algorithmDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, nodeID)
	err := algorithmDAO.raft.CreateOnDiskCluster(clusterID, join, sm.CreateOnDiskStateMachine)
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

func (algorithmDAO *RaftAlgorithmDAO) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.Algorithm], error) {
	algorithmDAO.logger.Infof("GetPage Raft entity *config.Algorithm page query: %+v", pageQuery)
	var result interface{}
	var err error
	emptyPageResult := dao.PageResult[*config.Algorithm]{
		Page:     pageQuery.Page,
		PageSize: pageQuery.PageSize}
	jsonPageQuery, err := json.Marshal(pageQuery)
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = algorithmDAO.raft.ReadLocal(algorithmDAO.clusterID, jsonPageQuery)
		if err != nil {
			algorithmDAO.logger.Errorf("Error (algorithmID=%d): %s", algorithmDAO.clusterID, err)
			return emptyPageResult, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = algorithmDAO.raft.SyncRead(algorithmDAO.clusterID, jsonPageQuery)
		if err != nil {
			algorithmDAO.logger.Errorf("Error (algorithmID=%d): %s", algorithmDAO.clusterID, err)
			return emptyPageResult, err
		}
	}
	switch v := result.(type) {
	case dao.PageResult[*config.Algorithm]:
		return v, nil
	default:
		algorithmDAO.logger.Errorf("%s:  query type: %T", statemachine.ErrUnsupportedQuery, v)
		return emptyPageResult, statemachine.ErrUnsupportedQuery
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
