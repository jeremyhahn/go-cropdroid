package provisioner

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"

	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
)

func TestProvisioner(t *testing.T) {

	farmDAO, provisioner, params := createDefaultProvisioner()

	role := model.NewRole()
	role.SetName("test")

	user := model.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("dev")
	user.SetRoles([]common.Role{role})

	farm, err := provisioner.Provision(user, params)
	assert.Nil(t, err)

	assert.NotNil(t, farm)
	assert.Equal(t, 4, len(farm.GetDevices()))

	farms, err := farmDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(farms))

	CurrentTest.Cleanup()
}

func TestProvisionerMultipleFarms(t *testing.T) {

	farmDAO, provisioner, params := createDefaultProvisioner()

	role := model.NewRole()
	role.SetName("test")

	user := model.NewUser()
	user.SetEmail("root@localhost")
	user.SetPassword("dev")
	user.SetRoles([]common.Role{role})

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

	farms, err := farmDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(farms))

	CurrentTest.Cleanup()
}

func createDefaultProvisioner() (dao.FarmDAO, FarmProvisioner, *common.ProvisionerParams) {
	it := NewIntegrationTest()
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	userMapper := mapper.NewUserMapper()
	farmDAO := gormstore.NewFarmDAO(it.logger, it.gorm, it.idGenerator)
	initializer := gormstore.NewGormInitializer(it.logger, it.db, idGenerator,
		it.location, common.CONFIG_MODE_VIRTUAL)
	params := &common.ProvisionerParams{
		ConfigStoreType: config.MEMORY_STORE,
		StateStoreType:  state.MEMORY_STORE,
		DataStoreType:   datastore.GORM_STORE,
	}
	return farmDAO, NewGormFarmProvisioner(it.logger, it.gorm, it.location,
		farmDAO, nil, nil, userMapper, initializer), params
}
