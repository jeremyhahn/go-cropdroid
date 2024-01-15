package provisioner

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormFarmProvisioner struct {
	logger                *logging.Logger
	db                    *gorm.DB
	location              *time.Location
	farmDAO               dao.FarmDAO
	farmProvisionerChan   chan config.Farm
	farmDeprovisionerChan chan config.Farm
	userMapper            mapper.UserMapper
	initializer           datastore.Initializer
	FarmProvisioner
}

func NewGormFarmProvisioner(logger *logging.Logger, db *gorm.DB, location *time.Location,
	farmDAO dao.FarmDAO, farmProvisionerChan chan config.Farm,
	farmDeprovisionerChan chan config.Farm, userMapper mapper.UserMapper,
	initializer datastore.Initializer) FarmProvisioner {

	return &GormFarmProvisioner{
		logger:                logger,
		db:                    db,
		location:              location,
		farmDAO:               farmDAO,
		farmProvisionerChan:   farmProvisionerChan,
		farmDeprovisionerChan: farmDeprovisionerChan,
		userMapper:            userMapper,
		initializer:           initializer}
}

func (provisioner *GormFarmProvisioner) Provision(userAccount common.UserAccount, params *common.ProvisionerParams) (*config.Farm, error) {
	userConfig := provisioner.userMapper.MapUserModelToConfig(userAccount)
	farmConfig, err := provisioner.initializer.BuildConfig(params.OrganizationID, userConfig, userAccount.GetRoles()[0])
	if err != nil {
		return nil, err
	}
	if provisioner.farmProvisionerChan != nil {
		select {
		case provisioner.farmProvisionerChan <- *farmConfig:
		default:
			provisioner.logger.Debugf("[FarmProvisioner.Provision] Waiting for provisioning request to complete. user=%s", userAccount.GetEmail())
		}
	}
	return farmConfig, provisioner.farmDAO.Save(farmConfig)
}

func (provisioner *GormFarmProvisioner) Deprovision(
	userAccount common.UserAccount, farmID uint64) error {

	farmConfig, err := provisioner.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	provisioner.farmDeprovisionerChan <- *farmConfig
	return nil
}
