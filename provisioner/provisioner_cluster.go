// +build cluster

package provisioner

import (
	"errors"
	"time"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"

	logging "github.com/op/go-logging"
)

type RaftFarmProvisioner struct {
	logger      *logging.Logger
	gossip      cluster.GossipCluster
	location    *time.Location
	farmDAO     dao.FarmDAO
	userMapper  mapper.UserMapper
	initializer datastore.Initializer
	FarmProvisioner
}

func NewRaftFarmProvisioner(logger *logging.Logger, gossip cluster.GossipCluster,
	location *time.Location, farmDAO dao.FarmDAO, userMapper mapper.UserMapper,
	initializer datastore.Initializer) FarmProvisioner {

	return &RaftFarmProvisioner{
		logger:      logger,
		gossip:      gossip,
		location:    location,
		farmDAO:     farmDAO,
		userMapper:  userMapper,
		initializer: initializer}
}

func (provisioner *RaftFarmProvisioner) Provision(userAccount common.UserAccount, params *ProvisionerParams) (config.FarmConfig, error) {
	userConfig := provisioner.userMapper.MapUserModelToEntity(userAccount)
	farmConfig, err := provisioner.initializer.BuildConfig(userConfig, userAccount.GetRoles()[0])
	if err != nil {
		return nil, err
	}
	if err := provisioner.farmDAO.Save(farmConfig.(*config.Farm)); err != nil {
		return nil, err
	}
	if err := provisioner.gossip.Provision(farmConfig); err != nil {
		return nil, err
	}
	return farmConfig, nil
}

func (provisioner *RaftFarmProvisioner) Deprovision(userAccount common.UserAccount, farmID uint64) error {
	return errors.New("Not implemented")
}
