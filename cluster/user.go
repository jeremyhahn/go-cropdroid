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

type RaftUserDAO struct {
	logger    *logging.Logger
	raft      RaftNode
	clusterID uint64
	dao.UserDAO
}

func NewRaftUserDAO(logger *logging.Logger, raftNode RaftNode,
	clusterID uint64) dao.UserDAO {

	return &RaftUserDAO{
		logger:    logger,
		raft:      raftNode,
		clusterID: clusterID}
}

func (userDAO *RaftUserDAO) StartCluster() {
	params := userDAO.raft.GetParams()
	clusterID := params.RaftOptions.UserClusterID
	sm := statemachine.NewUserConfigMachine(userDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, params.NodeID)
	err := userDAO.raft.CreateOnDiskCluster(clusterID, params.Join, sm.CreateUserConfigMachine)
	if err != nil {
		userDAO.logger.Fatal(err)
	}
	userDAO.raft.WaitForClusterReady(clusterID)
}

func (userDAO *RaftUserDAO) Get(userID uint64, CONSISTENCY_LEVEL int) (*config.User, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = userDAO.raft.ReadLocal(userDAO.clusterID, userID)
		if err != nil {
			userDAO.logger.Errorf("Error (userID=%d): %s", userID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = userDAO.raft.SyncRead(userDAO.clusterID, userID)
		if err != nil {
			userDAO.logger.Errorf("Error (userID=%d): %s", userID, err)
			return nil, err
		}
	}
	if result != nil {
		return result.(*config.User), nil
	}
	return nil, datastore.ErrNotFound
}

func (userDAO *RaftUserDAO) Save(user *config.User) error {

	if user.GetID() == 0 {
		idSetter := userDAO.raft.GetParams().IdSetter
		idSetter.SetUserIds([]*config.User{user})
	}

	userDAO.logger.Debugf("Saving user: %+v", user)

	perm, err := json.Marshal(user)
	if err != nil {
		userDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		userDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := userDAO.raft.SyncPropose(userDAO.clusterID, proposal); err != nil {
		userDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

func (userDAO *RaftUserDAO) Delete(user *config.User) error {
	userDAO.logger.Debugf(fmt.Sprintf("Deleting user record: %+v", user))
	perm, err := json.Marshal(user)
	if err != nil {
		userDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, perm).Serialize()
	if err != nil {
		userDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := userDAO.raft.SyncPropose(userDAO.clusterID, proposal); err != nil {
		userDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}
