//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"encoding/json"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	logging "github.com/op/go-logging"
)

type ServerDAO interface {
	GetConfig(CONSISTENCY_LEVEL int) (*config.Server, error)
	Get(id uint64, CONSISTENCY_LEVEL int) (*config.Server, error)
	Save(server *config.Server) error
	Delete(server *config.Server) error
}

type RaftServerDAO struct {
	logger    *logging.Logger
	raft      RaftNode
	clusterID uint64
	ServerDAO
}

func NewRaftServerDAO(logger *logging.Logger, raftNode RaftNode,
	clusterID uint64) ServerDAO {
	return &RaftServerDAO{
		logger:    logger,
		raft:      raftNode,
		clusterID: clusterID}
}

func (serverDAO *RaftServerDAO) StartCluster() {
	params := serverDAO.raft.GetParams()
	clusterID := params.RaftOptions.SystemClusterID
	sm := statemachine.NewServerConfigMachine(serverDAO.logger,
		params.IdGenerator, params.DataDir, clusterID, params.NodeID)
	err := serverDAO.raft.CreateOnDiskCluster(clusterID, params.Join, sm.CreateServerConfigMachine)
	if err != nil {
		serverDAO.logger.Fatal(err)
	}
	serverDAO.raft.WaitForClusterReady(clusterID)
}

func (serverDAO *RaftServerDAO) GetConfig(CONSISTENCY_LEVEL int) (*config.Server, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = serverDAO.raft.ReadLocal(serverDAO.clusterID, serverDAO.clusterID)
		if err != nil {
			serverDAO.logger.Errorf("Error (serverID=%d): %s", serverDAO.clusterID, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = serverDAO.raft.SyncRead(serverDAO.clusterID, serverDAO.clusterID)
		if err != nil {
			serverDAO.logger.Errorf("Error (serverID=%d): %s", serverDAO.clusterID, err)
			return nil, err
		}
	}
	if result != nil {
		return result.(*config.Server), nil
	}
	return nil, datastore.ErrNotFound
}

func (serverDAO *RaftServerDAO) Get(id uint64, CONSISTENCY_LEVEL int) (*config.Server, error) {
	var result interface{}
	var err error
	if CONSISTENCY_LEVEL == common.CONSISTENCY_LOCAL {
		result, err = serverDAO.raft.ReadLocal(serverDAO.clusterID, id)
		if err != nil {
			serverDAO.logger.Errorf("Error (serverClusterID=%d, serverID=%d): %s",
				serverDAO.clusterID, id, err)
			return nil, err
		}
	} else if CONSISTENCY_LEVEL == common.CONSISTENCY_QUORUM {
		result, err = serverDAO.raft.SyncRead(serverDAO.clusterID, id)
		if err != nil {
			serverDAO.logger.Errorf("Error (serverClusterID=%d, serverID=%d): %s",
				serverDAO.clusterID, id, err)
			return nil, err
		}
	}
	if result != nil {
		return result.(*config.Server), nil
	}
	return nil, datastore.ErrNotFound
}

func (serverDAO *RaftServerDAO) Save(server *config.Server) error {
	if server.GetID() == 0 {
		server.SetID(serverDAO.clusterID)
	}
	serverDAO.logger.Debugf("Saving server: %+v", server)
	perm, err := json.Marshal(server)
	if err != nil {
		serverDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	if err != nil {
		serverDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := serverDAO.raft.SyncPropose(serverDAO.clusterID, proposal); err != nil {
		serverDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

func (serverDAO *RaftServerDAO) Delete(server *config.Server) error {
	serverDAO.logger.Debugf(fmt.Sprintf("Deleting server record: %+v", server))
	perm, err := json.Marshal(server)
	if err != nil {
		serverDAO.logger.Errorf("Error: %s", err)
		return err
	}
	proposal, err := statemachine.CreateProposal(
		statemachine.QUERY_TYPE_DELETE, perm).Serialize()
	if err != nil {
		serverDAO.logger.Errorf("Error: %s", err)
		return err
	}
	if err := serverDAO.raft.SyncPropose(serverDAO.clusterID, proposal); err != nil {
		serverDAO.logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}
