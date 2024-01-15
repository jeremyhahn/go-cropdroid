//go:build cluster
// +build cluster

package provisioner

import (
	"errors"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type RaftFarmProvisioner struct {
	app                 *app.App
	gossip              cluster.GossipNode
	location            *time.Location
	farmDAO             dao.FarmDAO
	userDAO             dao.UserDAO
	userMapper          mapper.UserMapper
	initializer         dao.Initializer
	farmProvisionerChan chan config.Farm
	FarmProvisioner
}

func NewRaftFarmProvisioner(app *app.App, gossip cluster.GossipNode,
	location *time.Location, farmDAO dao.FarmDAO, userDAO dao.UserDAO,
	userMapper mapper.UserMapper, initializer dao.Initializer) FarmProvisioner {

	return &RaftFarmProvisioner{
		app:         app,
		gossip:      gossip,
		location:    location,
		farmDAO:     farmDAO,
		userDAO:     userDAO,
		userMapper:  userMapper,
		initializer: initializer}
}

func (provisioner *RaftFarmProvisioner) Provision(
	userAccount common.UserAccount, params *common.ProvisionerParams) (*config.Farm, error) {

	userConfig := provisioner.userMapper.MapUserModelToConfig(userAccount)

	// Send provisioning request to all nodes in the cluster
	if err := provisioner.gossip.Provision(params); err != nil {
		return nil, err
	}

	// Construct a new config.Farm and set the default user
	farmConfig, err := provisioner.initializer.BuildConfig(params)
	if err != nil {
		return nil, err
	}
	farmConfig.SetUsers([]*config.User{userConfig})

	// Add org/farm refs to the user
	if params.OrganizationID > 0 {
		userConfig.AddOrganizationRef(params.OrganizationID)
	}
	userConfig.AddFarmRef(farmConfig.GetID())
	if err := provisioner.userDAO.Save(userConfig); err != nil {
		return nil, err
	}

	// Save the newly provisioned farm to the database
	if err := provisioner.farmDAO.Save(farmConfig); err != nil {
		return nil, err
	}

	return farmConfig, nil
}

func (provisioner *RaftFarmProvisioner) Deprovision(userAccount common.UserAccount, farmID uint64) error {
	return errors.New("Not implemented")
}
