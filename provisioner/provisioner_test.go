package provisioner

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"

	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
)

func TestProvisioner(t *testing.T) {

	initParams := &common.ProvisionerParams{
		UserID:           0,
		RoleID:           0,
		OrganizationID:   0,
		FarmName:         common.DEFAULT_CROP_NAME,
		ConfigStoreType:  config.MEMORY_STORE,
		StateStoreType:   state.MEMORY_STORE,
		DataStoreType:    datastore.GORM_STORE,
		ConsistencyLevel: common.CONSISTENCY_LOCAL}

	farmDAO, userDAO, provisioner, initializer := createDefaultProvisioner()

	_, err := initializer.Initialize(false, initParams)
	assert.Nil(t, err)

	roleName := "test"
	roleID := CurrentTest.idGenerator.NewRoleID(roleName)
	role := config.NewRole()
	role.SetID(roleID)
	role.SetName(roleName)

	email := "root@localhost"
	userID := CurrentTest.idGenerator.NewUserID(email)
	user := config.NewUser()
	user.SetID(userID)
	user.SetEmail(email)
	user.SetPassword("dev")
	user.SetRoles([]*config.Role{role})

	err = userDAO.Save(user)
	assert.Nil(t, err)

	persistedUser, err := userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)

	userMapper := mapper.NewUserMapper()
	userModel := userMapper.MapUserConfigToModel(persistedUser)
	assert.Equal(t, email, userModel.GetEmail())

	provParams := &common.ProvisionerParams{
		UserID:           userID,
		RoleID:           roleID,
		OrganizationID:   0,
		FarmName:         common.DEFAULT_CROP_NAME,
		ConfigStoreType:  config.MEMORY_STORE,
		StateStoreType:   state.MEMORY_STORE,
		DataStoreType:    datastore.GORM_STORE,
		ConsistencyLevel: common.CONSISTENCY_LOCAL}

	farm, err := provisioner.Provision(userModel, provParams)
	assert.Nil(t, err)

	assert.NotNil(t, farm)
	assert.Equal(t, 4, len(farm.GetDevices()))

	page1, err := farmDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(page1.Entities))

	CurrentTest.Cleanup()
}

func TestProvisionerMultipleFarms(t *testing.T) {

	initParams := &common.ProvisionerParams{
		UserID:           0,
		RoleID:           0,
		OrganizationID:   0,
		FarmName:         common.DEFAULT_CROP_NAME,
		ConfigStoreType:  config.MEMORY_STORE,
		StateStoreType:   state.MEMORY_STORE,
		DataStoreType:    datastore.GORM_STORE,
		ConsistencyLevel: common.CONSISTENCY_LOCAL}

	farmDAO, userDAO, provisioner, initializer := createDefaultProvisioner()

	_, err := initializer.Initialize(false, initParams)
	assert.Nil(t, err)

	roleName := "test"
	roleID := CurrentTest.idGenerator.NewRoleID(roleName)
	role := config.NewRole()
	role.SetID(roleID)
	role.SetName(roleName)

	email := "root@localhost"
	userID := CurrentTest.idGenerator.NewUserID(email)
	user := config.NewUser()
	user.SetID(userID)
	user.SetEmail(email)
	user.SetPassword("dev")
	user.SetRoles([]*config.Role{role})

	err = userDAO.Save(user)
	assert.Nil(t, err)

	persistedUser, err := userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)

	userMapper := mapper.NewUserMapper()
	userModel := userMapper.MapUserConfigToModel(persistedUser)
	assert.Equal(t, email, userModel.GetEmail())

	farm1, err := provisioner.Provision(userModel, &common.ProvisionerParams{
		UserID:           userID,
		RoleID:           roleID,
		OrganizationID:   0,
		FarmName:         "farm1",
		ConfigStoreType:  config.MEMORY_STORE,
		StateStoreType:   state.MEMORY_STORE,
		DataStoreType:    datastore.GORM_STORE,
		ConsistencyLevel: common.CONSISTENCY_LOCAL})
	assert.Nil(t, err)

	farm2, err := provisioner.Provision(userModel, &common.ProvisionerParams{
		UserID:           userID,
		RoleID:           roleID,
		OrganizationID:   0,
		FarmName:         "farm2",
		ConfigStoreType:  config.MEMORY_STORE,
		StateStoreType:   state.MEMORY_STORE,
		DataStoreType:    datastore.GORM_STORE,
		ConsistencyLevel: common.CONSISTENCY_LOCAL})
	assert.Nil(t, err)

	farm3, err := provisioner.Provision(userModel, &common.ProvisionerParams{
		UserID:           userID,
		RoleID:           roleID,
		OrganizationID:   0,
		FarmName:         "farm3",
		ConfigStoreType:  config.MEMORY_STORE,
		StateStoreType:   state.MEMORY_STORE,
		DataStoreType:    datastore.GORM_STORE,
		ConsistencyLevel: common.CONSISTENCY_LOCAL})
	assert.Nil(t, err)

	assert.NotNil(t, farm1)
	assert.Equal(t, 4, len(farm1.GetDevices()))
	assert.Equal(t, 4, len(farm2.GetDevices()))
	assert.Equal(t, 4, len(farm3.GetDevices()))

	page1, err := farmDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(page1.Entities))

	CurrentTest.Cleanup()
}

func createDefaultProvisioner() (dao.FarmDAO, dao.UserDAO, FarmProvisioner,
	dao.Initializer) {

	it := NewIntegrationTest()
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	userMapper := mapper.NewUserMapper()
	farmDAO := gormstore.NewFarmDAO(it.logger, it.gorm, it.idGenerator)
	userDAO := gormstore.NewUserDAO(it.logger, it.gorm)
	permissionDAO := gormstore.NewPermissionDAO(it.logger, it.gorm)

	datastoreRegistry := gorm.NewGormRegistry(it.logger, it.db)

	configInitializer := dao.NewConfigInitializer(it.logger,
		idGenerator, it.location, datastoreRegistry, it.passwordHasher,
		"virtual")

	return farmDAO, userDAO, NewGormFarmProvisioner(it.logger, it.gorm, it.location,
		farmDAO, permissionDAO, nil, nil, userMapper, configInitializer), configInitializer
}
