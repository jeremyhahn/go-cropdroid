package provisioner

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormFarmProvisioner struct {
	logger                *logging.Logger
	db                    *gorm.DB
	location              *time.Location
	farmDAO               dao.FarmDAO
	permissionDAO         dao.PermissionDAO
	farmProvisionerChan   chan config.Farm
	farmDeprovisionerChan chan config.Farm
	userMapper            mapper.UserMapper
	initializer           dao.Initializer
	FarmProvisioner
}

func NewGormFarmProvisioner(logger *logging.Logger, db *gorm.DB,
	location *time.Location, farmDAO dao.FarmDAO,
	permissionDAO dao.PermissionDAO, farmProvisionerChan chan config.Farm,
	farmDeprovisionerChan chan config.Farm, userMapper mapper.UserMapper,
	initializer dao.Initializer) FarmProvisioner {

	return &GormFarmProvisioner{
		logger:                logger,
		db:                    db,
		location:              location,
		farmDAO:               farmDAO,
		permissionDAO:         permissionDAO,
		farmProvisionerChan:   farmProvisionerChan,
		farmDeprovisionerChan: farmDeprovisionerChan,
		userMapper:            userMapper,
		initializer:           initializer}
}

func (provisioner *GormFarmProvisioner) Provision(userAccount common.UserAccount, params *common.ProvisionerParams) (*config.Farm, error) {

	userMappser := mapper.NewUserMapper()
	user := userMappser.MapUserModelToConfig(userAccount)

	// Build farm config
	farmConfig, permissions, err := provisioner.initializer.BuildConfig(params, user)
	if err != nil {
		provisioner.logger.Error(err)
	}

	if provisioner.farmProvisionerChan != nil {
		select {
		case provisioner.farmProvisionerChan <- *farmConfig:
		default:
			provisioner.logger.Debugf("[FarmProvisioner.Provision] Waiting for provisioning request to complete. user=%s", userAccount.GetEmail())
		}
	}

	// Save the farm config
	if err := provisioner.farmDAO.Save(farmConfig); err != nil {
		return nil, err
	}

	// Save permission entries
	for _, permission := range permissions {
		if err = provisioner.permissionDAO.Save(&permission); err != nil {
			return nil, err
		}
	}

	return farmConfig, nil
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
