package provisioner

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/stretchr/testify/assert"

	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
)

func TestProvisioner(t *testing.T) {

	farmDAO, provisioner, params := createDefaultProvisioner()

	user := &model.User{
		Email:    "root@localhost",
		Password: "dev"}

	farm, err := provisioner.Provision(user, params)
	assert.Nil(t, err)

	assert.NotNil(t, farm)
	assert.Equal(t, 4, len(farm.GetDevices()))

	farms, err := farmDAO.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(farms))

	CurrentTest.Cleanup()
}

func TestProvisionerMultipleFarms(t *testing.T) {

	farmDAO, provisioner, params := createDefaultProvisioner()

	user := &model.User{
		Email:    "root@localhost",
		Password: "dev"}

	farm1, err := provisioner.Provision(user, params)
	assert.Nil(t, err)

	farm2, err := provisioner.Provision(user, params)
	assert.Nil(t, err)

	farm3, err := provisioner.Provision(user, params)
	assert.Nil(t, err)

	assert.NotNil(t, farm1)
	assert.Equal(t, 4, len(farm1.GetDevices()))
	assert.Equal(t, 4, len(farm2.GetDevices()))
	assert.Equal(t, 4, len(farm3.GetDevices()))

	farms, err := farmDAO.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(farms))

	CurrentTest.Cleanup()
}

func createDefaultProvisioner() (dao.FarmDAO, FarmProvisioner, *ProvisionerParams) {
	it := NewIntegrationTest()
	userMapper := mapper.NewUserMapper()
	farmDAO := gormstore.NewFarmDAO(it.logger, it.gorm)
	initializer := gormstore.NewGormInitializer(it.logger, it.db, it.location)
	params := &ProvisionerParams{
		ConfigStore: config.MEMORY_STORE,
		StateStore:  state.MEMORY_STORE,
		DataStore:   datastore.GORM_STORE,
	}
	return farmDAO, NewGormFarmProvisioner(it.logger, it.gorm, it.location,
		farmDAO, nil, nil, userMapper, initializer), params
}
