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
	"github.com/jeremyhahn/go-cropdroid/test/backup/cluster/statemachine"
	logging "github.com/op/go-logging"
)

type RaftFarmConfigDAO struct {
	logger    *logging.Logger
	raft      cluster.RaftNode
	serverDAO ServerDAO
	userDAO   dao.UserDAO
	dao.FarmDAO
}

func NewRaftFarmConfigDAO(logger *logging.Logger, raftNode cluster.RaftNode,
	serverDAO ServerDAO, userDAO dao.UserDAO) dao.FarmDAO {
	return &RaftFarmConfigDAO{
		logger:    logger,
		raft:      raftNode,
		serverDAO: serverDAO,
		userDAO:   userDAO}
}

func (farmDAO *RaftFarmConfigDAO) StartCluster(clusterID uint64) {
	params := farmDAO.raft.GetParams()
	sm := statemachine.NewUserConfigMachine(farmDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, params.NodeID)
	err := farmDAO.raft.CreateOnDiskCluster(clusterID, params.Join, sm.CreateUserConfigMachine)
	if err != nil {
		farmDAO.logger.Fatal(err)
	}
	farmDAO.raft.WaitForClusterReady(clusterID)
}

func (farmDAO *RaftFarmConfigDAO) GetByUserID(userID uint64,
	CONSISTENCY_LEVEL int) ([]*config.Farm, error) {

	farmDAO.logger.Debugf("Fetching farms for user: %d", userID)

	user, err := farmDAO.userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}

	farmIDs := user.GetFarmRefs()
	farms := make([]*config.Farm, len(farmIDs))

	for i, farmID := range farmIDs {
		var result interface{}
		if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
			result, err = farmDAO.raft.ReadLocal(farmID, farmID)
			if err != nil {
				farmDAO.logger.Errorf("Error (farmID=%d): %s", farmID, err)
				return nil, err
			}
		} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
			result, err = farmDAO.raft.SyncRead(farmID, farmID)
			if err != nil {
				farmDAO.logger.Errorf("Error (farmID=%d): %s", farmID, err)
				return nil, err
			}
		}
		switch v := result.(type) {
		case *config.Farm:
			farms[i] = v
		default:
			farmDAO.logger.Errorf("unexpected query type %T", v)
			return []*config.Farm{}, nil
		}
	}
	return farms, nil
}

func (farmDAO *RaftFarmConfigDAO) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.Farm], error) {
	pageResult := dao.PageResult[*config.Farm]{
		Page:     pageQuery.Page,
		PageSize: pageQuery.PageSize}
	entities, err := farmDAO.GetAll(CONSISTENCY_LEVEL)
	if err != nil {
		return pageResult, err
	}
	pageResult.Entities = entities
	return pageResult, nil
}

func (farmDAO *RaftFarmConfigDAO) GetAll(CONSISTENCY_LEVEL int) ([]*config.Farm, error) {
	farmDAO.logger.Debugf("Fetching all farms")
	var result interface{}
	var err error
	serverConfig, err := farmDAO.serverDAO.GetConfig(CONSISTENCY_LEVEL)
	if err != nil {
		return nil, err
	}
	farmIDs := serverConfig.GetFarmRefs()
	farms := make([]*config.Farm, len(farmIDs))
	for i, farmID := range farmIDs {
		if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
			result, err = farmDAO.raft.ReadLocal(farmID, farmID)
			if err != nil {
				farmDAO.logger.Errorf("Error (farmID=%d): %s", farmID, err)
				return nil, err
			}
		} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
			result, err = farmDAO.raft.SyncRead(farmID, farmID)
			if err != nil {
				farmDAO.logger.Errorf("Error (farmID=%d): %s", farmID, err)
				return nil, err
			}
		}
		switch v := result.(type) {
		case *config.Farm:
			v.ParseSettings()
			farms[i] = v
		default:
			farmDAO.logger.Errorf("unexpected query type %T", v)
			return []*config.Farm{}, nil
		}
	}
	return farms, nil
}

func (farmDAO *RaftFarmConfigDAO) GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) ([]*config.Farm, error) {
	farms := make([]*config.Farm, 0)
	for _, farmID := range farmIds {
		farm, err := farmDAO.Get(farmID, CONSISTENCY_LEVEL)
		if err != nil {
			return nil, err
		}
		farms = append(farms, farm)
	}
	return farms, nil
}

func (farmDAO *RaftFarmConfigDAO) Get(farmID uint64, CONSISTENCY_LEVEL int) (*config.Farm, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = farmDAO.raft.ReadLocal(farmID, farmID)
		if err != nil {
			farmDAO.logger.Errorf("Error (farmID=%d): %s", farmID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = farmDAO.raft.SyncRead(farmID, farmID)
		if err != nil {
			farmDAO.logger.Errorf("Error (farmID=%d): %s", farmID, err)
			return nil, err
		}
	}
	if result != nil {
		farmConfig := result.(*config.Farm)
		farmConfig.ParseSettings()
		return farmConfig, nil
	}
	return nil, datastore.ErrNotFound
}

func (farmDAO *RaftFarmConfigDAO) Save(farm *config.Farm) error {

	idSetter := farmDAO.raft.GetParams().IdSetter
	idSetter.SetIds(farm)

	for _, device := range farm.GetDevices() {
		if device.GetInterval() == 0 {
			device.SetInterval(farm.GetInterval())
		}
	}

	farmDAO.logger.Debugf("Saving farm: %+v", farm)

	farmJson, err := json.Marshal(farm)
	if err != nil {
		farmDAO.logger.Errorf("[RaftFarmConfigDAO.Save] Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, farmJson).Serialize()
	if err != nil {
		farmDAO.logger.Errorf("[RaftFarmConfigDAO.Save] Error: %s", err)
		return err
	}
	if err := farmDAO.raft.SyncPropose(farm.ID, proposal); err != nil {
		farmDAO.logger.Errorf("[RaftFarmConfigDAO.Save] Error: %s", err)
		return err
	}

	// Update server farm refs
	serverConfig, err := farmDAO.serverDAO.GetConfig(farm.GetConsistencyLevel())
	if err != nil {
		return err
	}
	if !serverConfig.HasFarmRef(farm.ID) {
		farmDAO.logger.Debugf("[RaftFarmConfigDAO.Save] Adding server FarmRef: %d", farm.ID)
		serverConfig.AddFarmRef(farm.ID)
		if err := farmDAO.serverDAO.Save(serverConfig); err != nil {
			return err
		}
	}

	// Update the user farm refs
	for _, user := range farm.GetUsers() {
		// This query expects / requires the user to be saved first
		userConfig, err := farmDAO.userDAO.Get(user.ID, farm.GetConsistencyLevel())
		if err != nil {
			return err
		}
		if !userConfig.HasFarmRef(farm.ID) {
			farmDAO.logger.Debugf("[RaftFarmConfigDAO.Save] Adding user FarmRef. userID=%d, farmID=%d",
				user.ID, farm.ID)
			userConfig.AddFarmRef(farm.ID)
			if err := farmDAO.userDAO.Save(userConfig); err != nil {
				return err
			}
		}
	}
	return nil
}

func (farmDAO *RaftFarmConfigDAO) Delete(farm *config.Farm) error {
	farmDAO.logger.Debugf(fmt.Sprintf("Deleting farm record: %+v", farm))
	perm, err := json.Marshal(farm)
	if err != nil {
		farmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, perm).Serialize()
	if err != nil {
		farmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := farmDAO.raft.SyncPropose(farm.ID, proposal); err != nil {
		farmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	// Delete server config farm refs
	serverConfig, err := farmDAO.serverDAO.GetConfig(common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	serverConfig.RemoveFarmRef(farm.ID)
	if err := farmDAO.serverDAO.Save(serverConfig); err != nil {
		return err
	}
	return nil
}
