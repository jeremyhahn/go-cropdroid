//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
	"github.com/stretchr/testify/assert"
)

func TestFarmAssociations(t *testing.T) {

	org, _, farm1DAO, _ := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)
	farm1 := org.GetFarms()[0]

	dstest.TestFarmAssociations(t, IntegrationTestCluster.app.IdGenerator, farm1DAO, farm1)
}

func TestFarmGetPage(t *testing.T) {

	org, _, farm1DAO, farm2DAO := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)
	farm1 := org.GetFarms()[0]

	page1, err := farm1DAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(page1.Entities))
	assert.Equal(t, farm1.Name, page1.Entities[0].GetName())

	farm2 := org.GetFarms()[1]
	page1, err = farm2DAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(page1.Entities))
	assert.Equal(t, farm2.Name, page1.Entities[0].GetName())
}

func TestFarmGet(t *testing.T) {

	org, _, farm1DAO, farm2DAO := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)
	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	result, err := farm1DAO.Get(farm1.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, farm1.Name, result.GetName())

	result, err = farm2DAO.Get(farm2.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, farm2.Name, result.GetName())
}

func TestFarmGetByIds(t *testing.T) {

	org, _, farm1DAO, farm2DAO := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)
	farm1 := org.GetFarms()[0]
	farm2 := org.GetFarms()[1]

	err := farm1DAO.Save(farm1)
	assert.Nil(t, err)

	err = farm2DAO.Save(farm2)
	assert.Nil(t, err)

	farmIDs := []uint64{farm1.ID, farm2.ID}
	farms, err := farm1DAO.GetByIds(farmIDs, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(farms))
}
