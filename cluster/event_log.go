//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"
	"sort"

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	logging "github.com/op/go-logging"
)

type RaftEventLogDAO struct {
	logger    *logging.Logger
	raft      RaftNode
	clusterID uint64
	dao.EventLogDAO
}

func NewRaftEventLogDAO(logger *logging.Logger, raftNode RaftNode,
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

func (eventLogDAO *RaftEventLogDAO) GetAll(CONSISTENCY_LEVEL int) ([]entity.EventLog, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = eventLogDAO.raft.ReadLocal(eventLogDAO.clusterID, statemachine.QUERY_TYPE_WILDCARD)
		if err != nil {
			eventLogDAO.logger.Errorf("Error: %s", err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = eventLogDAO.raft.SyncRead(eventLogDAO.clusterID, statemachine.QUERY_TYPE_WILDCARD)
		if err != nil {
			eventLogDAO.logger.Errorf("Error: %s", err)
			return nil, err
		}
	}
	if result != nil {
		items := result.([]entity.EventLog)
		return items, nil
	}
	return nil, datastore.ErrNotFound
}

func (eventLogDAO *RaftEventLogDAO) GetAllDesc(CONSISTENCY_LEVEL int) ([]entity.EventLog, error) {
	records, err := eventLogDAO.GetAll(CONSISTENCY_LEVEL)
	if err != nil {
		return nil, err
	}
	// Reverse the records array
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}
	return records, nil
}

func (eventLogDAO *RaftEventLogDAO) Save(record entity.EventLogEntity) error {
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

func (eventLogDAO *RaftEventLogDAO) Count(CONSISTENCY_LEVEL int) (int64, error) {
	// TODO: Make this efficient
	records, err := eventLogDAO.GetAll(common.CONSISTENCY_LOCAL)
	if err != nil {
		return 0, err
	}
	return int64(len(records)), nil
}

func (eventLogDAO *RaftEventLogDAO) GetPage(CONSISTENCY_LEVEL int, page, size int64) ([]entity.EventLog, error) {
	// TODO: Make this efficient
	// PebbleDB requires iterating the entire data set to get the count
	// Alternative approach is adding an on-write counter to the application
	records, err := eventLogDAO.GetAll(common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	recordLen := int64(len(records))
	var lowBound = int64(0)
	var highBound = recordLen
	if page == 0 {
		page = 1
	}
	if page > 1 {
		lowBound = (page - 1) * size
	}
	if recordLen > size {
		highBound = lowBound + size
	}
	recordSet := records[lowBound:highBound]
	// Sort records in descending order; most recent events first
	sort.SliceStable(recordSet, func(i, j int) bool {
		return recordSet[i].Timestamp.Unix() > recordSet[j].Timestamp.Unix()
	})
	return recordSet, nil
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
