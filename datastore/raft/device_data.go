//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"encoding/json"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/statemachine"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"

	logging "github.com/op/go-logging"
)

type RaftDeviceDataDAO interface {
	ClusterID() uint64
	CreateClusterNode(deviceID uint64) (uint64, error)
	datastore.DeviceDataStore
	RaftCluster
}

type RaftDeviceData struct {
	logger      *logging.Logger
	raft        cluster.RaftNode
	idGenerator util.IdGenerator
	GenericRaftDAO[*state.DeviceState]
}

// Creates a new DeviceDataStore Raft DAO using a persistent OnDiskStateMachine to store device
// telemetry data. This DAO implements both the DeviceDataStore which works with multiple devices,
// or it can be bound to a single device to perform standard DAO operations.
func NewRaftDeviceDataDAO(logger *logging.Logger, raftNode cluster.RaftNode, deviceID uint64) RaftDeviceDataDAO {

	idGenerator := raftNode.GetParams().IdGenerator
	deviceDataClusterID := idGenerator.CreateDeviceDataClusterID(deviceID)

	logger.Debugf("Creating new *state.DeviceState Raft DAO for deviceID: %d, deviceDataClusterID: %d",
		deviceID, deviceDataClusterID)

	return &RaftDeviceData{
		logger:      logger,
		raft:        raftNode,
		idGenerator: idGenerator,
		GenericRaftDAO: GenericRaftDAO[*state.DeviceState]{
			logger:    logger,
			raft:      raftNode,
			clusterID: deviceDataClusterID,
		}}
}

// Returns the unique device data cluster ID that belongs to the specified deviceID
func (deviceDataDAO *RaftDeviceData) ClusterID() uint64 {
	return deviceDataDAO.GenericRaftDAO.clusterID
}

/* Cluster operational methods */

func (deviceDataDAO *RaftDeviceData) CreateClusterNode(deviceID uint64) (uint64, error) {
	deviceDataDAO.logger.Debugf("Creating device data raft cluster for deviceID %d on node %d",
		deviceID, deviceDataDAO.raft.GetConfig().NodeID)
	deviceDataClusterID := deviceDataDAO.idGenerator.CreateDeviceDataClusterID(deviceID)
	params := deviceDataDAO.raft.GetParams()
	sm := statemachine.NewGenericOnDiskStateMachine[*state.DeviceState](deviceDataDAO.logger, deviceDataDAO.idGenerator,
		params.DataDir, deviceDataClusterID, params.NodeID)
	if err := deviceDataDAO.raft.CreateOnDiskCluster(deviceDataClusterID, params.Join, sm.CreateOnDiskStateMachine); err != nil {
		return 0, err
	}
	return deviceDataClusterID, nil
}

/* GenericRaftDAO methods */

// Starts a device data Raft cluster on the current node
func (deviceDataDAO *RaftDeviceData) StartClusterNode(waitForClusterReady bool) error {
	return deviceDataDAO.GenericRaftDAO.StartClusterNode(waitForClusterReady)
}

// Starts a local multi-node device data Raft cluster
func (deviceDataDAO *RaftDeviceData) StartLocalCluster(localCluster *LocalCluster, waitForClusterReady bool) error {
	return deviceDataDAO.GenericRaftDAO.StartLocalCluster(localCluster, waitForClusterReady)
}

func (deviceDataDAO *RaftDeviceData) WaitForClusterReady() {
	deviceDataDAO.GenericRaftDAO.WaitForClusterReady()
}

func (deviceDataDAO *RaftDeviceData) Update(deviceStateMap *state.DeviceState) error {
	return deviceDataDAO.GenericRaftDAO.Update(deviceStateMap)
}

func (deviceDataDAO *RaftDeviceData) Delete(deviceStateMap *state.DeviceState) error {
	return deviceDataDAO.GenericRaftDAO.Delete(deviceStateMap)
}

func (deviceDataDAO *RaftDeviceData) Get(id uint64, CONSISTENCY_LEVEL int) (*state.DeviceState, error) {
	return deviceDataDAO.GenericRaftDAO.Get(id, CONSISTENCY_LEVEL)
}

func (deviceDataDAO *RaftDeviceData) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*state.DeviceState], error) {
	return deviceDataDAO.GenericRaftDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
}

func (deviceDataDAO *RaftDeviceData) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*state.DeviceState], CONSISTENCY_LEVEL int) error {

	return deviceDataDAO.GenericRaftDAO.ForEachPage(pageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
}

/* DeviceDataStore methods */

// Puts a new device state entry into the Raft database
func (deviceDataDAO *RaftDeviceData) Save(deviceID uint64, deviceStateMap state.DeviceStateMap) error {
	deviceDataClusterID := deviceDataDAO.idGenerator.CreateDeviceDataClusterID(deviceID)
	deviceDataDAO.logger.Debugf("Save Raft entity *state.DeviceState for deviceID: %d, deviceDataClusterID: %d",
		deviceID, deviceDataClusterID)
	data, err := json.Marshal(deviceStateMap)
	if err != nil {
		deviceDataDAO.logger.Errorf("Save json.Marshal error (deviceID=%d, deviceDataClusterID=%d): %s",
			deviceID, deviceDataClusterID, err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, data).Serialize()
	if err != nil {
		deviceDataDAO.logger.Errorf("Save CreateProposal error (deviceID=%d, deviceDataClusterID=%d): %s",
			deviceID, deviceDataClusterID, err)
		return err
	}
	if err := deviceDataDAO.raft.SyncPropose(deviceDataClusterID, proposal); err != nil {
		deviceDataDAO.logger.Errorf("Save SyncPropose error (deviceID=%d, deviceDataClusterID=%d): %s",
			deviceID, deviceDataClusterID, err)
		return err
	}
	return nil
}

// Returns all records for the given metric within the last 30 days
func (deviceDataDAO *RaftDeviceData) GetLast30Days(deviceID uint64, metric string) ([]float64, error) {
	minutesInDay := 1440
	pageQuery := &query.PageQuery{
		Page:      1,
		PageSize:  minutesInDay * 30, // 43,200 minutes in a month
		SortOrder: query.SORT_DESCENDING}
	jsonPageQuery, err := json.Marshal(pageQuery)
	deviceDataClusterID := deviceDataDAO.idGenerator.CreateDeviceDataClusterID(deviceID)
	result, err := deviceDataDAO.raft.SyncRead(deviceDataClusterID, jsonPageQuery)
	if err != nil {
		deviceDataDAO.logger.Errorf("GetLast30Days SyncRead error (deviceID=%d, deviceDataClusterID=%d): %s",
			deviceID, deviceDataClusterID, err)
		return nil, err
	}
	var resultSet []float64
	if result != nil {
		pageResult := result.(dao.PageResult[*state.DeviceState])
		resultSet = make([]float64, len(pageResult.Entities))
		if len(pageResult.Entities) > 0 {
			for i, record := range pageResult.Entities {
				val, exists := record.GetMetrics()[metric]
				if !exists {
					return resultSet, datastore.ErrMetricKeyNotFound
				}
				resultSet[i] = val
			}
		}
		return resultSet, nil
	}
	return resultSet, nil
}
