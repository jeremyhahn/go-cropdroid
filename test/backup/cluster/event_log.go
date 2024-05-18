//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/entity"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/test/backup/cluster/statemachine"
	logging "github.com/op/go-logging"
)

type RaftEventLogDAO struct {
	logger    *logging.Logger
	raft      cluster.RaftNode
	clusterID uint64
	dao.EventLogDAO
}

func NewRaftEventLogDAO(logger *logging.Logger, raftNode cluster.RaftNode,
	farmID uint64) dao.EventLogDAO {

	eventLogClusterID := raftNode.GetParams().
		IdGenerator.CreateEventLogClusterID(farmID)

	return &RaftEventLogDAO{
		logger:    logger,
		raft:      raftNode,
		clusterID: eventLogClusterID}
}

func (eventLogDAO *RaftEventLogDAO) StartCluster() {
	params := eventLogDAO.raft.GetParams()
	sm := statemachine.NewEventLogOnDiskStateMachine(params.Logger, params.IdGenerator,
		params.DataDir, eventLogDAO.clusterID, params.NodeID)
	if err := eventLogDAO.raft.CreateOnDiskCluster(eventLogDAO.clusterID, params.Join, sm.CreateEventLogOnDiskStateMachine); err != nil {
		eventLogDAO.logger.Fatal(err)
	}
	//eventLogDAO.raft.WaitForClusterReady(eventLogDAO.clusterID)
}

func (eventLogDAO *RaftEventLogDAO) Save(record *entity.EventLog) error {
	data, err := json.Marshal(record)
	if err != nil {
		eventLogDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, data).Serialize()
	if err != nil {
		eventLogDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := eventLogDAO.raft.SyncPropose(eventLogDAO.clusterID, proposal); err != nil {
		eventLogDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

// func (eventLogDAO *RaftEventLogDAO) Count(CONSISTENCY_LEVEL int) (int64, error) {
// 	// TODO: Make this efficient
// 	records, err := eventLogDAO.GetAll(common.CONSISTENCY_LOCAL)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return int64(len(records)), nil
// }

// func (eventLogDAO *RaftEventLogDAO) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*entity.EventLog], error) {

// 	pageResult := dao.PageResult[*entity.EventLog]{
// 		Page:     pageQuery.Page,
// 		PageSize: pageQuery.PageSize}
// 	records, err := eventLogDAO.GetAll(common.CONSISTENCY_LOCAL)
// 	if err != nil {
// 		return pageResult, err
// 	}
// 	recordLen := len(records)
// 	var lowBound = 0
// 	var highBound = recordLen
// 	page := pageQuery.Page
// 	pageSize := pageQuery.PageSize
// 	if page == 0 {
// 		page = 1
// 	}
// 	if page > 1 {
// 		lowBound = (page - 1) * pageSize
// 	}
// 	if recordLen > pageSize {
// 		highBound = lowBound + pageSize
// 	}
// 	recordSet := records[lowBound:highBound]
// 	// Sort records in descending order; most recent events first
// 	sort.SliceStable(recordSet, func(i, j int) bool {
// 		return recordSet[i].Timestamp.Unix() > recordSet[j].Timestamp.Unix()
// 	})
// 	pageResult.Entities = recordSet
// 	return pageResult, nil
// }

func (eventLogDAO *RaftEventLogDAO) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*entity.EventLog], error) {
	eventLogDAO.logger.Infof("GetPage Raft entity *entity.EventLog: page query: %+v", pageQuery)
	var result interface{}
	var err error
	emptyPageResult := dao.PageResult[*entity.EventLog]{
		Page:     pageQuery.Page,
		PageSize: pageQuery.PageSize}
	jsonPageQuery, err := json.Marshal(pageQuery)
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = eventLogDAO.raft.ReadLocal(eventLogDAO.clusterID, jsonPageQuery)
		if err != nil {
			eventLogDAO.logger.Errorf("Error (eventLogID=%d): %s", eventLogDAO.clusterID, err)
			return emptyPageResult, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = eventLogDAO.raft.SyncRead(eventLogDAO.clusterID, jsonPageQuery)
		if err != nil {
			eventLogDAO.logger.Errorf("Error (eventLogID=%d): %s", eventLogDAO.clusterID, err)
			return emptyPageResult, err
		}
	}
	switch v := result.(type) {
	case dao.PageResult[*entity.EventLog]:
		return v, nil
	default:
		eventLogDAO.logger.Errorf("%s:  query type: %T", statemachine.ErrUnsupportedQuery, v)
		return emptyPageResult, statemachine.ErrUnsupportedQuery
	}
}

// func (eventLogDAO *RaftEventLogDAO) Delete(registration *entity.EventLog) error {
// 	eventLogDAO.logger.Debugf(fmt.Sprintf("Deleting event log record: %+v", registration))
// 	perm, err := json.Marshal(registration)
// 	if err != nil {
// 		eventLogDAO.logger.Errorf("Error: %s", err)
// 		return err
// 	}
// 	proposal, err := statemachine.CreateProposal(
// 		statemachine.QUERY_TYPE_DELETE, perm).Serialize()
// 	if err != nil {
// 		eventLogDAO.logger.Errorf("Error: %s", err)
// 		return err
// 	}
// 	if err := eventLogDAO.raft.SyncPropose(eventLogDAO.clusterID, proposal); err != nil {
// 		eventLogDAO.logger.Errorf("Error: %s", err)
// 		return err
// 	}
// 	return nil
// }
