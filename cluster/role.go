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

type RaftRoleDAO struct {
	logger    *logging.Logger
	raft      RaftNode
	clusterID uint64
	dao.RoleDAO
}

func NewRaftRoleDAO(logger *logging.Logger, raftNode RaftNode,
	clusterID uint64) dao.RoleDAO {

	return &RaftRoleDAO{
		logger:    logger,
		raft:      raftNode,
		clusterID: clusterID}
}

func (roleDAO *RaftRoleDAO) StartCluster() {
	params := roleDAO.raft.GetParams()
	clusterID := params.RaftOptions.RoleClusterID
	sm := statemachine.NewRoleConfigMachine(roleDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, params.NodeID)
	err := roleDAO.raft.CreateOnDiskCluster(clusterID, params.Join, sm.CreateRoleConfigMachine)
	if err != nil {
		roleDAO.logger.Fatal(err)
	}
	roleDAO.raft.WaitForClusterReady(clusterID)
}

func (roleDAO *RaftRoleDAO) Get(roleID uint64, CONSISTENCY_LEVEL int) (*config.Role, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = roleDAO.raft.ReadLocal(roleDAO.clusterID, roleID)
		if err != nil {
			roleDAO.logger.Errorf("Error (roleClusterID=%d, roleID=%d): %s",
				roleDAO.clusterID, roleID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = roleDAO.raft.SyncRead(roleDAO.clusterID, roleID)
		if err != nil {
			roleDAO.logger.Errorf("Error (roleClusterID=%d, roleID=%d): %s",
				roleDAO.clusterID, roleID, err)
			return nil, err
		}
	}
	if result != nil {
		return result.(*config.Role), nil
	}
	return nil, datastore.ErrNotFound
}

func (roleDAO *RaftRoleDAO) GetByName(roleName string, CONSISTENCY_LEVEL int) (*config.Role, error) {
	roleID := roleDAO.raft.GetParams().IdGenerator.NewID(roleName)
	return roleDAO.Get(roleID, CONSISTENCY_LEVEL)
}

func (roleDAO *RaftRoleDAO) Save(role *config.Role) error {

	if role.GetID() == 0 {
		idSetter := roleDAO.raft.GetParams().IdSetter
		idSetter.SetRoleIds([]*config.Role{role})
	}

	roleDAO.logger.Debugf("Saving role: %+v", role)

	perm, err := json.Marshal(role)
	if err != nil {
		roleDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		roleDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := roleDAO.raft.SyncPropose(roleDAO.clusterID, proposal); err != nil {
		roleDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

func (roleDAO *RaftRoleDAO) Delete(role *config.Role) error {
	roleDAO.logger.Debugf(fmt.Sprintf("Deleting role record: %+v", role))
	perm, err := json.Marshal(role)
	if err != nil {
		roleDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, perm).Serialize()
	if err != nil {
		roleDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := roleDAO.raft.SyncPropose(roleDAO.clusterID, proposal); err != nil {
		roleDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}
