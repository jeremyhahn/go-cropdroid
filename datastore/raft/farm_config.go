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

type RaftFarmConfigDAO interface {
	StartClusterNode(farmID uint64, waitForClusterReady bool) error
	StartLocalCluster(localCluster *LocalCluster, farmID uint64, waitForClusterReady bool) error
	dao.FarmDAO
}

type RaftFarmConfig struct {
	logger    *logging.Logger
	raft      cluster.RaftNode
	serverDAO dao.ServerDAO
	userDAO   dao.UserDAO
	RaftFarmConfigDAO
}

// Creates a new Raft DAO for a FarmConfig. This DAO binds a single FarmConfig
// instance to the Raft database.
func NewRaftFarmConfigDAO(logger *logging.Logger, raftNode cluster.RaftNode,
	serverDAO dao.ServerDAO, userDAO dao.UserDAO) RaftFarmConfigDAO {

	logger.Debugf("Creating new *config.Farm Raft DAO")
	return &RaftFarmConfig{
		logger:    logger,
		raft:      raftNode,
		serverDAO: serverDAO,
		userDAO:   userDAO}
}

func (farmDAO *RaftFarmConfig) StartClusterNode(farmID uint64, waitForClusterReady bool) error {
	params := farmDAO.raft.GetParams()
	nodeID := params.GetNodeID()
	farmDAO.logger.Debugf("StartClusterNode *config.Farm raft cluster %d on node %d", farmID, nodeID)
	sm := statemachine.NewGenericOnDiskStateMachine[*config.FarmStruct](farmDAO.logger,
		params.IdGenerator, params.DataDir, farmID, nodeID)
	err := farmDAO.raft.CreateOnDiskCluster(farmID, params.Join, sm.CreateOnDiskStateMachine)
	if err != nil {
		farmDAO.logger.Errorf("StartClusterNode error starting *config.Farm raft cluster: ", err)
		return err
	}
	if waitForClusterReady {
		farmDAO.raft.WaitForClusterReady(farmID)
	}
	return nil
}

func (farmDAO *RaftFarmConfig) StartLocalCluster(localCluster *LocalCluster, farmID uint64, waitForClusterReady bool) error {
	localCluster.app.Logger.Debugf("Creating local %d node *config.Farm raft cluster: %d",
		localCluster.nodeCount, farmID)
	for i := 0; i < localCluster.nodeCount; i++ {
		raftNode := localCluster.GetRaftNode(i)
		err := NewRaftFarmConfigDAO(farmDAO.logger, raftNode, farmDAO.serverDAO,
			farmDAO.userDAO).StartClusterNode(farmID, false)
		if err != nil {
			farmDAO.logger.Errorf("StartLocalCluster error starting *config.Farm raft cluster on node %d: %s", i, err)
			return err
		}
	}
	if waitForClusterReady {
		farmDAO.raft.WaitForClusterReady(farmID)
	}
	return nil
}

func (farmDAO *RaftFarmConfig) Save(farmConfig *config.FarmStruct) error {

	idSetter := farmDAO.raft.GetParams().IdSetter
	idSetter.SetIds(farmConfig)

	for _, device := range farmConfig.GetDevices() {
		if device.GetInterval() == 0 {
			device.SetInterval(farmConfig.GetInterval())
		}
	}

	farmDAO.logger.Debugf("Save *config.Farm Raft entity: %+v", farmConfig)

	farmJson, err := json.Marshal(farmConfig)
	if err != nil {
		farmDAO.logger.Errorf("Save *config.Farm json.Marshal error: %+v. %s", farmConfig, err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, farmJson).Serialize()
	if err != nil {
		farmDAO.logger.Errorf("Save *config.Farm CreateProposal error: %+v. %s", farmConfig, err)
		return err
	}
	if err := farmDAO.raft.SyncPropose(farmConfig.ID, proposal); err != nil {
		farmDAO.logger.Errorf("Save *config.Farm SyncPropose error: %+v. %s", farmConfig, err)
		return err
	}

	// Update server farm refs
	serverConfig, err := farmDAO.serverDAO.Get(farmDAO.raft.GetParams().ClusterID, farmConfig.GetConsistencyLevel())
	if err != nil {
		farmDAO.logger.Errorf("Save *config.Farm failed to retrieve server config: %s", err)
		return err
	}
	if !serverConfig.HasFarmRef(farmConfig.ID) {
		farmDAO.logger.Debugf("Save Adding server FarmRef: %d", farmConfig.ID)
		serverConfig.AddFarmRef(farmConfig.ID)
		if err := farmDAO.serverDAO.Save(serverConfig); err != nil {
			farmDAO.logger.Errorf("Save *config.Farm server refs error: %s", err)
			return err
		}
	}

	// Update the user farm refs
	for _, user := range farmConfig.GetUsers() {
		// This query expects / requires the user to be saved first
		userConfig, err := farmDAO.userDAO.Get(user.ID, farmConfig.GetConsistencyLevel())
		if err != nil {
			return err
		}
		if !userConfig.HasFarmRef(farmConfig.ID) {
			farmDAO.logger.Debugf("Save Adding user FarmRef. userID=%d, farmID=%d",
				user.ID, farmConfig.ID)
			userConfig.AddFarmRef(farmConfig.ID)
			if err := farmDAO.userDAO.Save(userConfig); err != nil {
				farmDAO.logger.Errorf("Save *config.Farm user refs error: %s", err)
				return err
			}
		}
	}
	return nil
}

func (farmDAO *RaftFarmConfig) Delete(farm *config.FarmStruct) error {
	farmDAO.logger.Debugf(fmt.Sprintf("Delete Raft entity *config.Farm: %d", farm.ID))
	perm, err := json.Marshal(farm)
	if err != nil {
		farmDAO.logger.Errorf("delete error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, perm).Serialize()
	if err != nil {
		farmDAO.logger.Errorf("delete CreateProposal error: %s", err)
		return err
	}
	if err := farmDAO.raft.SyncPropose(farm.ID, proposal); err != nil {
		farmDAO.logger.Errorf("delete SyncPropose error: %s", err)
		return err
	}
	// Delete server config farm refs
	serverID := farmDAO.raft.GetParams().RaftOptions.SystemClusterID
	serverConfig, err := farmDAO.serverDAO.Get(serverID, common.CONSISTENCY_LOCAL)
	if err != nil {
		farmDAO.logger.Errorf("delete couldn't delete server FarmRefs. error: %s", err)
		return err
	}
	serverConfig.RemoveFarmRef(farm.ID)
	if err := farmDAO.serverDAO.Save(serverConfig); err != nil {
		farmDAO.logger.Errorf("delete error saving updated server FarmRefs: %s", err)
		return err
	}
	return nil
}

func (farmDAO *RaftFarmConfig) Get(farmID uint64, CONSISTENCY_LEVEL int) (*config.FarmStruct, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = farmDAO.raft.ReadLocal(farmID, farmID)
		if err != nil {
			farmDAO.logger.Errorf("Get ReadLocal error (farmID=%d): %s", farmID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = farmDAO.raft.SyncRead(farmID, farmID)
		if err != nil {
			farmDAO.logger.Errorf("Get SyncRead error (farmID=%d): %s", farmID, err)
			return nil, err
		}
	}
	if result != nil {
		farmConfig := result.(*config.FarmStruct)
		farmConfig.ParseSettings()
		return farmConfig, nil
	}
	return nil, datastore.ErrRecordNotFound
}

func (farmDAO *RaftFarmConfig) GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) ([]*config.FarmStruct, error) {
	farmDAO.logger.Debugf("GetByIds FarmConfig Raft entity: %+v", farmIds)
	farms := make([]*config.FarmStruct, 0)
	for _, farmID := range farmIds {
		farm, err := farmDAO.Get(farmID, CONSISTENCY_LEVEL)
		if err != nil {
			farmDAO.logger.Errorf("GetByIds FarmConfig unable to locate requested farm ID: %d, Error: %s", farmID, err)
			return nil, err
		}
		farms = append(farms, farm)
	}
	return farms, nil
}

func (farmDAO *RaftFarmConfig) GetByUserID(userID uint64, CONSISTENCY_LEVEL int) ([]*config.FarmStruct, error) {
	farmDAO.logger.Debugf("GetByUserID FarmConfig Raft query, userID: %d", userID)

	user, err := farmDAO.userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}

	farmIDs := user.GetFarmRefs()
	farms := make([]*config.FarmStruct, len(farmIDs))

	for i, farmID := range farmIDs {
		var result interface{}
		if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
			result, err = farmDAO.raft.ReadLocal(farmID, farmID)
			if err != nil {
				farmDAO.logger.Errorf("GetByIds FarmConfig ReadLocal error. userID: %d, farmID: %d, error: %s",
					userID, farmID, err)
				return nil, err
			}
		} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
			result, err = farmDAO.raft.SyncRead(farmID, farmID)
			if err != nil {
				farmDAO.logger.Errorf("GetByIds FarmConfig SyncRead error. userID: %d, farmID: %d, error: %s",
					userID, farmID, err)
				return nil, err
			}
		}
		switch v := result.(type) {
		case *config.FarmStruct:
			farms[i] = v
		default:
			farmDAO.logger.Errorf("GetByIds FarmConfig unexpected query type: %T", v)
			return []*config.FarmStruct{}, nil
		}
	}
	return farms, nil
}
