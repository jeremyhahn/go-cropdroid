package raft

import (
	"encoding/json"
	"time"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/index"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/statemachine"
)

type RaftDAO[E any] interface {
	dao.GenericDAO[E]
	StartClusterNode(waitForClusterReady bool) error
	StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error
	SaveWithTimeSeriesIndex(entity E) error
}

type GenericRaftDAO[E any] struct {
	logger          *logging.Logger
	raft            cluster.RaftNode
	clusterID       uint64
	autoIncrementer uint64
	RaftDAO[E]
}

func NewGenericRaftDAO[E any](logger *logging.Logger, raftNode cluster.RaftNode,
	clusterID uint64) RaftDAO[E] {

	logger.Infof("Creatnig new %T Raft DAO", *new(E))
	return &GenericRaftDAO[E]{
		logger:    logger,
		raft:      raftNode,
		clusterID: clusterID}
}

func (dao *GenericRaftDAO[E]) StartClusterNode(waitForClusterReady bool) error {
	params := dao.raft.GetParams()
	clusterID := dao.clusterID
	nodeID := params.GetNodeID()
	join := false
	sm := statemachine.NewGenericOnDiskStateMachine[E](dao.logger,
		params.IdGenerator, params.DataDir, clusterID, nodeID)
	err := dao.raft.CreateOnDiskCluster(clusterID, join, sm.CreateOnDiskStateMachine)
	if err != nil {
		dao.logger.Fatal(err)
		return err
	}
	if waitForClusterReady {
		dao.raft.WaitForClusterReady(clusterID)
	}
	return nil
}

func (dao *GenericRaftDAO[E]) StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error {
	localCluster.app.Logger.Debugf("Creating generic raft cluster %T: %d", *new(E), dao.clusterID)
	for i := 0; i < localCluster.nodeCount; i++ {
		raftNode := localCluster.GetRaftNode(i)
		genericDAO := NewGenericRaftDAO[E](dao.logger, raftNode, dao.clusterID)
		err := genericDAO.StartClusterNode(false)
		if err != nil {
			dao.logger.Fatal(err)
		}
	}
	if waitForClusterReady {
		dao.raft.WaitForClusterReady(dao.clusterID)
	}
	return nil
}

func (dao *GenericRaftDAO[E]) SaveWithTimeSeriesIndex(entity E) error {

	dao.logger.Infof("Save Raft entity: %+v", entity)

	tsEntity := any(entity).(config.TimeSeriesIndexeder)

	if tsEntity != nil {
		// The entity ID is expected to be set by the caller
		timeSeriesIndex := index.NewTimeSeriesIndex()
		tsEntity.SetTimestamp(timeSeriesIndex.Timestamp)
	}

	perm, err := json.Marshal(tsEntity)
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}

	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}

	if err := dao.raft.SyncPropose(dao.clusterID, proposal); err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}

	return nil
}

func (dao *GenericRaftDAO[E]) Get(id uint64, CONSISTENCY_LEVEL int) (E, error) {
	dao.logger.Infof("Get Raft entity with id: %d", id)
	var result interface{}
	var err error
	var emptyResponsse = *new(E)
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = dao.raft.ReadLocal(dao.clusterID, id)
		if err != nil {
			dao.logger.Errorf("Error with entity id: %d. %s", id, err)
			return emptyResponsse, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = dao.raft.SyncRead(dao.clusterID, id)
		if err != nil {
			dao.logger.Errorf("Error with entity id: %d. %s", id, err)
			return emptyResponsse, err
		}
	}
	if result != nil {
		return result.(E), nil
	}
	return emptyResponsse, datastore.ErrNotFound
}

func (dao *GenericRaftDAO[E]) GetPage(page, pageSize, CONSISTENCY_LEVEL int) ([]E, error) {
	dao.logger.Debugf("Fetching supported algorithms")
	var result interface{}
	var err error
	wildcardQuery := &query.Page{
		Page:     page,
		PageSize: pageSize}
	jsonWildcardQuery, err := json.Marshal(wildcardQuery)
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = dao.raft.ReadLocal(dao.clusterID, jsonWildcardQuery)
		if err != nil {
			dao.logger.Errorf("Error (algorithmID=%d): %s", dao.clusterID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = dao.raft.SyncRead(dao.clusterID, jsonWildcardQuery)
		if err != nil {
			dao.logger.Errorf("Error (algorithmID=%d): %s", dao.clusterID, err)
			return nil, err
		}
	}
	switch v := result.(type) {
	case []E:
		return v, nil
	default:
		dao.logger.Errorf("%s:  query type: %T", statemachine.ErrUnsupportedQuery, v)
		algorithms := make([]E, 0)
		return algorithms, statemachine.ErrUnsupportedQuery
	}
}

func (dao *GenericRaftDAO[E]) Save(entity E) error {

	dao.logger.Infof("Save Raft entity: %+v", entity)

	kvEntity := any(entity).(config.KeyValueEntity)

	if kvEntity.Identifier() == 0 {
		kvEntity.SetID(uint64(time.Now().UnixMicro()))
	}

	perm, err := json.Marshal(entity)
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}

	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}

	if err := dao.raft.SyncPropose(dao.clusterID, proposal); err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}

	return nil
}

func (dao *GenericRaftDAO[E]) Delete(entity E) error {
	dao.logger.Infof("Delete Raft entity: %+v", entity)
	marshaledEntity, err := json.Marshal(entity)
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, marshaledEntity).Serialize()
	if err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}
	if err := dao.raft.SyncPropose(dao.clusterID, proposal); err != nil {
		dao.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

// Only here to satisfy interface compatibility with ORM DAO's
func (dao *GenericRaftDAO[E]) Update(entity E) error {
	return dao.Save(entity)
}
