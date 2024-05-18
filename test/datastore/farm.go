package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"
)

func TestFarmAssociations(t *testing.T, idGenerator util.IdGenerator,
	farmDAO dao.FarmDAO, farmConfig *config.Farm) {

	// currentTest.gorm.Create(&config.Permission{
	// 	OrganizationID: 0,
	// 	FarmID:         farm.ID,
	// 	UserID:         user.ID,
	// 	RoleID:         role.ID})

	err := farmDAO.Save(farmConfig)
	assert.Nil(t, err)

	persisted, err := farmDAO.Get(farmConfig.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)

	assert.Greater(t, farmConfig.ID, uint64(0))

	assert.Equal(t, 1, len(persisted.GetUsers()))
	assert.Equal(t, "root@localhost", persisted.GetUsers()[0].GetEmail())

	assert.Equal(t, 1, len(persisted.GetUsers()[0].GetRoles()))
	assert.Equal(t, "test", persisted.GetUsers()[0].GetRoles()[0].GetName())
}

func TestFarmGetByIds(t *testing.T, farmDAO dao.FarmDAO,
	farm1, farm2 *config.Farm) {

	err := farmDAO.Save(farm1)
	assert.Nil(t, err)

	err = farmDAO.Save(farm2)
	assert.Nil(t, err)

	farms, err := farmDAO.GetByIds([]uint64{
		farm1.ID,
		farm2.ID}, DEFAULT_CONSISTENCY_LEVEL)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(farms))
}

func TestFarmGetPage(t *testing.T, farmDAO dao.FarmDAO,
	farm1, farm2 *config.Farm) {

	page1, err := farmDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(page1.Entities))

	assert.Equal(t, FARM1_NAME, page1.Entities[0].GetName())
	assert.Equal(t, FARM2_NAME, page1.Entities[1].GetName())
}

func TestFarmGet(t *testing.T, farmDAO dao.FarmDAO,
	farm1, farm2 *config.Farm) {

	persitedFarm1, err := farmDAO.Get(farm1.ID, DEFAULT_CONSISTENCY_LEVEL)
	assert.Nil(t, err)
	assert.Equal(t, FARM1_NAME, persitedFarm1.GetName())
	assert.Equal(t, "test", persitedFarm1.GetMode())

	persitedFarm2, err := farmDAO.Get(farm2.ID, DEFAULT_CONSISTENCY_LEVEL)
	assert.Nil(t, err)
	assert.Equal(t, FARM2_NAME, persitedFarm2.GetName())
	assert.Equal(t, "test2", persitedFarm2.GetMode())
}
