package provisioner

import (
	"errors"
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
	logger              *logging.Logger
	db                  *gorm.DB
	location            *time.Location
	farmDAO             dao.FarmDAO
	farmProvisionerChan chan config.FarmConfig
	userMapper          mapper.UserMapper
	initializer         datastore.Initializer
	FarmProvisioner
}

func NewGormFarmProvisioner(logger *logging.Logger, db *gorm.DB, location *time.Location,
	farmDAO dao.FarmDAO, farmProvisionerChan chan config.FarmConfig,
	userMapper mapper.UserMapper, initializer datastore.Initializer) FarmProvisioner {

	return &GormFarmProvisioner{
		logger:              logger,
		db:                  db,
		location:            location,
		farmDAO:             farmDAO,
		farmProvisionerChan: farmProvisionerChan,
		userMapper:          userMapper,
		initializer:         initializer}
}

func (provisioner *GormFarmProvisioner) Provision(userAccount common.UserAccount) (config.FarmConfig, error) {
	userConfig := provisioner.userMapper.MapUserModelToEntity(userAccount)
	farmConfig, err := provisioner.initializer.BuildConfig(userConfig)
	if err != nil {
		return nil, err
	}
	if provisioner.farmProvisionerChan != nil {
		select {
		case provisioner.farmProvisionerChan <- farmConfig:
		default:
			provisioner.logger.Error("[FarmProvisioner.Provision] Unable to send farm provisioner request! user=%s", userAccount.GetEmail())
		}
	}
	return farmConfig, provisioner.farmDAO.Save(farmConfig.(*config.Farm))
}

func (provisioner *GormFarmProvisioner) Deprovision(userAccount common.UserAccount) error {
	return errors.New("Not implemented")
}
