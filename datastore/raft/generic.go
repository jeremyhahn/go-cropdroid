package raft

import (
	"encoding/json"
	"time"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/index"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/statemachine"
)

type RaftCluster interface {
	StartClusterNode(waitForClusterReady bool) error
	StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error
	WaitForClusterReady()
}

type RaftDAO[E any] interface {
	dao.GenericDAO[E]
	RaftCluster
	SaveWithTimeSeriesIndex(entity E) error
}

type GenericRaftDAO[E any] struct {
	logger          *logging.Logger
	raft            cluster.RaftNode
	clusterID       uint64
	autoIncrementer uint64
	RaftDAO[E]
}

// Creates a new Raft DAO for the entity specified in the Generics parameter
func NewGenericRaftDAO[E any](logger *logging.Logger, raftNode cluster.RaftNode,
	clusterID uint64) RaftDAO[E] {

	logger.Debugf("Creating new %T Raft DAO", *new(E))
	return &GenericRaftDAO[E]{
		logger:    logger,
		raft:      raftNode,
		clusterID: clusterID}
}

// Starts a Raft cluster on the current node
func (genericDAO *GenericRaftDAO[E]) StartClusterNode(waitForClusterReady bool) error {
	params := genericDAO.raft.GetParams()
	clusterID := genericDAO.clusterID
	nodeID := params.GetNodeID()
	genericDAO.logger.Debugf("Starting raft cluster %d for %T on node %d", clusterID, *new(E), nodeID)
	sm := statemachine.NewGenericOnDiskStateMachine[E](genericDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, nodeID)
	err := genericDAO.raft.CreateOnDiskCluster(clusterID, params.Join, sm.CreateOnDiskStateMachine)
	if err != nil {
		genericDAO.logger.Errorf("StartClusterNode error starting OnDiskCluster %T on node %d", *new(E), nodeID)
		return err
	}
	if waitForClusterReady {
		genericDAO.raft.WaitForClusterReady(clusterID)
	}
	return nil
}

// Starts a local multi-node Raft cluster
func (genericDAO *GenericRaftDAO[E]) StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error {
	localCluster.app.Logger.Debugf("Creating local %d node raft cluster %T: %d",
		localCluster.nodeCount, *new(E), genericDAO.clusterID)
	for i := 0; i < localCluster.nodeCount; i++ {
		raftNode := localCluster.GetRaftNode(i)
		nodeGenericDAO := NewGenericRaftDAO[E](genericDAO.logger, raftNode, genericDAO.clusterID)
		err := nodeGenericDAO.StartClusterNode(false)
		if err != nil {
			genericDAO.logger.Errorf("StartLocalCluster error starting cluster %T on node %d", *new(E), i)
			return err
		}
	}
	if waitForClusterReady {
		genericDAO.raft.WaitForClusterReady(genericDAO.clusterID)
	}
	return nil
}

// Waits for the Raft cluster to establish the quorum and become ready for requests
func (genericDAO *GenericRaftDAO[E]) WaitForClusterReady() {
	genericDAO.logger.Debugf("Waiting for raft cluster %d for %T to become ready", genericDAO.clusterID, *new(E))
	genericDAO.raft.WaitForClusterReady(genericDAO.clusterID)
}

// Retrieves the entity with the specified ID from the database
func (genericDAO *GenericRaftDAO[E]) Get(id uint64, CONSISTENCY_LEVEL int) (E, error) {
	genericDAO.logger.Infof("Get Raft entity %T: %d", *new(E), id)
	var result interface{}
	var err error
	var emptyResponsse = *new(E)
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = genericDAO.raft.ReadLocal(genericDAO.clusterID, id)
		if err != nil {
			genericDAO.logger.Errorf("Get ReadLocal error: %T with ID: %d. %s", *new(E), id, err)
			return emptyResponsse, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = genericDAO.raft.SyncRead(genericDAO.clusterID, id)
		if err != nil {
			genericDAO.logger.Errorf("Get SyncRead error: %T with ID: %d. %s", *new(E), id, err)
			return emptyResponsse, err
		}
	}
	if result != nil {
		return result.(E), nil
	}
	return emptyResponsse, datastore.ErrRecordNotFound
}

// Retrieves a paginated set of records from the database
func (genericDAO *GenericRaftDAO[E]) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[E], error) {
	genericDAO.logger.Infof("GetPage Raft entity %T: pageQuery: %+v", *new(E), pageQuery)
	var result interface{}
	var err error
	emptyPageResult := dao.PageResult[E]{
		Page:     pageQuery.Page,
		PageSize: pageQuery.PageSize}
	jsonPageQuery, err := json.Marshal(pageQuery)
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = genericDAO.raft.ReadLocal(genericDAO.clusterID, jsonPageQuery)
		if err != nil {
			genericDAO.logger.Errorf("GetPage ReadLocal error: %T with clusterID: %d. %s", *new(E), genericDAO.clusterID, err)
			return emptyPageResult, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = genericDAO.raft.SyncRead(genericDAO.clusterID, jsonPageQuery)
		if err != nil {
			genericDAO.logger.Errorf("GetPage SyncRead error: %T with clusterID: %d. %s", *new(E), genericDAO.clusterID, err)
			return emptyPageResult, err
		}
	}
	switch v := result.(type) {
	case dao.PageResult[E]:
		return v, nil
	default:
		genericDAO.logger.Errorf("GetPage error: %s:  query type: %T", statemachine.ErrUnsupportedQuery, v)
		return emptyPageResult, statemachine.ErrUnsupportedQuery
	}
}

// Performs a recursive paginated query, passing the entites retuned in each
// PageResult to the specified PagerProcFunc for processing until there aren't
// anymore pages available to process. This method allows processing all entites
// in the table in a controlled manner by buffering only PageQuery.PageSize results
// at a time into the PagerProcFunc.
func (genericDAO *GenericRaftDAO[E]) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[E], CONSISTENCY_LEVEL int) error {

	genericDAO.logger.Debugf("ForEachPage Raft entity %T. query: %+v", *new(E), pageQuery)

	pageResult, err := genericDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
	if err != nil {
		genericDAO.logger.Errorf("ForEachPage error: %T with clusterID: %d. %s", *new(E), genericDAO.clusterID, err)
		return nil
	}
	if err = pagerProcFunc(pageResult.Entities); err != nil {
		genericDAO.logger.Errorf("ForEachPage PagerProcFunc error: %T with clusterID: %d. %s", *new(E), genericDAO.clusterID, err)
		return err
	}
	if pageResult.HasMore {
		nextPageQuery := query.PageQuery{
			Page:      pageQuery.Page + 1,
			PageSize:  pageQuery.PageSize,
			SortOrder: pageQuery.SortOrder}
		return genericDAO.ForEachPage(nextPageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
	}
	return nil
}

// Saves a new entity to the database
func (genericDAO *GenericRaftDAO[E]) Save(entity E) error {

	kvEntity := any(entity).(config.KeyValueEntity)

	genericDAO.logger.Debugf("Save Raft entity %T: %d", *new(E), kvEntity.Identifier())

	if kvEntity.Identifier() == 0 {
		kvEntity.SetID(uint64(time.Now().UnixMicro()))
	}

	perm, err := json.Marshal(entity)
	if err != nil {
		genericDAO.logger.Errorf("Save json.Marshal error: %T with entity: %+v. %s", *new(E), entity, err)
		return err
	}

	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		genericDAO.logger.Errorf("Save CreateProposal error: %T with entity: %+v. %s", *new(E), entity, err)
		return err
	}

	if err := genericDAO.raft.SyncPropose(genericDAO.clusterID, proposal); err != nil {
		genericDAO.logger.Errorf("Save SyncPropose error: %T with entity: %+v. %s", *new(E), entity, err)
		return err
	}

	return nil
}

// Saves an entity with a time series index to the database.
func (genericDAO *GenericRaftDAO[E]) SaveWithTimeSeriesIndex(entity E) error {

	tsEntity := any(entity).(config.TimeSeriesIndexeder)
	genericDAO.logger.Debugf("Save Raft entity %T with timeseries index: %d", *new(E), tsEntity.Timestamp())

	if tsEntity != nil {
		// The entity ID is expected to be set by the caller
		timeSeriesIndex := index.NewTimeSeriesIndex()
		tsEntity.SetTimestamp(timeSeriesIndex.Timestamp)
	}

	perm, err := json.Marshal(tsEntity)
	if err != nil {
		genericDAO.logger.Errorf("SaveWithTimeSeriesIndex json.Marshal error: %T with ID: %d. %s",
			*new(E), tsEntity.Identifier(), err)
		return err
	}

	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		genericDAO.logger.Errorf("SaveWithTimeSeriesIndex CreateProposal error: %T with ID: %d. %s",
			*new(E), tsEntity.Identifier(), err)
		return err
	}

	if err := genericDAO.raft.SyncPropose(genericDAO.clusterID, proposal); err != nil {
		genericDAO.logger.Errorf("SaveWithTimeSeriesIndex SyncPropose error: %T with ID: %d. %s",
			*new(E), tsEntity.Identifier(), err)
		return err
	}

	return nil
}

// Only here to satisfy interface compatibility with ORM genericDAO's
func (genericDAO *GenericRaftDAO[E]) Update(entity E) error {
	kvEntity := any(entity).(config.KeyValueEntity)
	genericDAO.logger.Infof("Update Raft entity %T: %d", *new(E), kvEntity.Identifier())
	return genericDAO.Save(entity)
}

// Deletes an entity from the database
func (genericDAO *GenericRaftDAO[E]) Delete(entity E) error {
	kvEntity := any(entity).(config.KeyValueEntity)
	genericDAO.logger.Debugf("Delete Raft entity %T: %d", *new(E), kvEntity.Identifier())
	marshaledEntity, err := json.Marshal(entity)
	if err != nil {
		genericDAO.logger.Errorf("Delete json.Marshal error: %T with entity: %+v. %s", *new(E), entity, err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, marshaledEntity).Serialize()
	if err != nil {
		genericDAO.logger.Errorf("Delete CreateProposal error: %T with entity: %+v. %s", *new(E), entity, err)
		return err
	}
	if err := genericDAO.raft.SyncPropose(genericDAO.clusterID, proposal); err != nil {
		genericDAO.logger.Errorf("Delete SyncPropose error: %T with entity: %+v. %s", *new(E), entity, err)
		return err
	}
	return nil
}

// Returns the total number of entities in the database. This function iterates all entities in the
// table and returns the sum. As this is a logarithmic operation, it should be considered an expensive
// operation, and other approaches should be considered instead, such as using the ForEachPage method
// that provides a controlled, buffered read to process all entites in the table without exhausting
// system resources, and both retrieves the entities and processes them within a single logarithmic
// operation.
func (genericDAO *GenericRaftDAO[E]) Count(CONSISTENCY_LEVEL int) (int64, error) {
	genericDAO.logger.Infof("Count Raft entity %T: ", *new(E))
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = genericDAO.raft.ReadLocal(genericDAO.clusterID, query.QUERY_TYPE_COUNT)
		if err != nil {
			genericDAO.logger.Errorf("Count ReadLocal error: %T. %s", *new(E), err)
			return 0, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = genericDAO.raft.SyncRead(genericDAO.clusterID, query.QUERY_TYPE_COUNT)
		if err != nil {
			genericDAO.logger.Errorf("Count SyncRead error: %T. %s", *new(E), err)
			return 0, err
		}
	}
	switch v := result.(type) {
	case int64:
		return v, nil
	default:
		genericDAO.logger.Errorf("Count Raft error %s:  query type: %T", statemachine.ErrUnsupportedQuery, v)
		return 0, statemachine.ErrUnsupportedQuery
	}
}
