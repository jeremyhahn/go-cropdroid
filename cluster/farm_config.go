//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	logging "github.com/op/go-logging"
)

type RaftFarmConfigDAO struct {
	logger    *logging.Logger
	raft      RaftNode
	serverDAO ServerDAO
	userDAO   dao.UserDAO
	dao.FarmDAO
}

func NewRaftFarmConfigDAO(logger *logging.Logger, raftNode RaftNode,
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

	idGenerator := farmDAO.raft.GetParams().IdGenerator
	if farm.GetID() == 0 {
		farmKey := fmt.Sprintf("%d-%s", farm.GetOrganizationID(), farm.GetName())
		farm.SetID(idGenerator.NewID(farmKey))
	}
	for _, device := range farm.GetDevices() {
		if device.GetID() == 0 {
			deviceKey := fmt.Sprintf("%d-%s", farm.GetID(), device.GetType())
			device.SetID(idGenerator.NewID(deviceKey))

			for _, deviceSetting := range device.GetSettings() {
				if deviceSetting.GetID() == 0 {
					deviceSettingsKey := fmt.Sprintf("%d-%s", device.GetID(), deviceSetting.GetKey())
					deviceSetting.SetID(idGenerator.NewID(deviceSettingsKey))
				}
			}

			for _, metric := range device.GetMetrics() {
				if metric.GetID() == 0 {
					metricKey := fmt.Sprintf("%d-%s", device.GetID(), metric.GetKey())
					metric.SetID(idGenerator.NewID(metricKey))
				}
			}

			for _, channel := range device.GetChannels() {
				if channel.GetID() == 0 {
					channelKey := fmt.Sprintf("%d-%s", device.GetID(), channel.GetName())
					channel.SetID(idGenerator.NewID(channelKey))
				}

				for _, condition := range channel.GetConditions() {
					if condition.GetID() == 0 {
						conditionKey := fmt.Sprintf("%d-%s", device.GetID(), condition.String())
						condition.SetID(idGenerator.NewID(conditionKey))
					}
				}

				for _, schedule := range channel.GetSchedule() {
					if schedule.GetID() == 0 {
						scheduleKey := fmt.Sprintf("%d-%s", device.GetID(), schedule.String())
						schedule.SetID(idGenerator.NewID(scheduleKey))
					}
				}
			}
		}
	}
	for _, user := range farm.GetUsers() {
		if user.GetID() == 0 {
			user.SetID(idGenerator.NewID(user.GetEmail()))
		}
		for _, role := range user.GetRoles() {
			if role.GetID() == 0 {
				role.SetID(idGenerator.NewID(role.GetName()))
			}
		}
	}
	for _, workflow := range farm.GetWorkflows() {
		if workflow.GetID() == 0 {
			workflowKey := fmt.Sprintf("%d-%s", farm.GetID(), workflow.GetName())
			workflow.SetID(idGenerator.NewID(workflowKey))
		}
		for _, workflowStep := range workflow.GetSteps() {
			if workflowStep.GetID() == 0 {
				workflowStepKey := fmt.Sprintf("%d-%s", workflow.GetID(), workflowStep.String())
				workflowStep.SetID(idGenerator.NewID(workflowStepKey))
			}
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
	if err := farmDAO.raft.SyncPropose(farm.GetID(), proposal); err != nil {
		farmDAO.logger.Errorf("[RaftFarmConfigDAO.Save] Error: %s", err)
		return err
	}

	// Update server farm refs
	serverConfig, err := farmDAO.serverDAO.GetConfig(farm.GetConsistencyLevel())
	if err != nil {
		return err
	}
	if serverConfig.HasFarmRef(farm.GetID()) {
		farmDAO.logger.Warningf("[RaftFarmConfigDAO.Save] Server FarmRef already exists! farmID=%d",
			farm.GetID())
	} else {
		serverConfig.AddFarmRef(farm.GetID())
		if err := farmDAO.serverDAO.Save(serverConfig); err != nil {
			return err
		}
	}

	// Update the user farm refs
	for _, user := range farm.GetUsers() {

		farmDAO.logger.Errorf("Updating farm refs: %+v", user)

		// This query expects / requires the user to be saved first
		userConfig, err := farmDAO.userDAO.Get(user.GetID(), farm.GetConsistencyLevel())
		if err != nil {
			return err
		}
		if userConfig.HasFarmRef(farm.GetID()) {
			farmDAO.logger.Warningf("[RaftFarmConfigDAO.Save] User FarmRef already exists! userID=%d, farmID=%d",
				user.GetID(), farm.GetID())
		} else {
			userConfig.AddFarmRef(farm.GetID())
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
	if err := farmDAO.raft.SyncPropose(farm.GetID(), proposal); err != nil {
		farmDAO.logger.Errorf("Error: %s", err)
		return err
	}
	// Delete server config farm refs
	serverConfig, err := farmDAO.serverDAO.GetConfig(common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	serverConfig.RemoveFarmRef(farm.GetID())
	if err := farmDAO.serverDAO.Save(serverConfig); err != nil {
		return err
	}
	return nil
}
