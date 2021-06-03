// +build cluster

package provisioner

import (
	"errors"
	"time"

	"github.com/jeremyhahn/cropdroid/cluster"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type RaftFarmProvisioner struct {
	logger          *logging.Logger
	gossip          cluster.GossipCluster
	location        *time.Location
	gormProvisioner *GormFarmProvisioner
	FarmProvisioner
}

func NewRaftFarmProvisioner(logger *logging.Logger, db *gorm.DB, gossip cluster.GossipCluster,
	location *time.Location, farmDAO dao.FarmDAO) FarmProvisioner {
	return &RaftFarmProvisioner{
		logger:   logger,
		gossip:   gossip,
		location: location,
		gormProvisioner: &GormFarmProvisioner{
			logger:   logger,
			db:       db,
			location: location,
			farmDAO:  farmDAO}}
}

func (provisioner *RaftFarmProvisioner) BuildConfig(adminUser common.UserAccount) (config.FarmConfig, error) {
	return provisioner.gormProvisioner.BuildConfig(adminUser)
}

func (provisioner *RaftFarmProvisioner) Provision(userAccount common.UserAccount) (config.FarmConfig, error) {
	farmConfig, err := provisioner.BuildConfig(userAccount)
	if err != nil {
		return nil, err
	}
	if err := provisioner.gossip.Provision(farmConfig); err != nil {
		return nil, err
	}
	if err := provisioner.gormProvisioner.farmDAO.Save(farmConfig.(*config.Farm)); err != nil {
		return nil, err
	}
	return farmConfig, nil
}

func (provisioner *RaftFarmProvisioner) Deprovision(userAccount common.UserAccount) error {
	return errors.New("Not implemented")
}
