//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"fmt"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"

	"github.com/stretchr/testify/assert"
)

func TestAlgorithmCRUD(t *testing.T) {

	ClusterID = 1

	genericDAO := NewGenericRaftDAO[*config.Algorithm](
		IntegrationTestCluster.app.Logger,
		IntegrationTestCluster.GetRaftNode1(),
		ClusterID)
	assert.NotNil(t, genericDAO)

	// Add a new algorithm raft to each of the local cluster nodes
	genericDAO.StartLocalCluster(IntegrationTestCluster, true)

	algorithm1 := config.NewAlgorithm()
	algorithm1.Name = "Test Algorithm 1"

	algorithm2 := config.NewAlgorithm()
	algorithm2.Name = "Test Algorithm 2"

	err := genericDAO.Save(algorithm1)
	assert.Nil(t, err)

	err = genericDAO.Save(algorithm2)
	assert.Nil(t, err)

	pageResult, err := genericDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(pageResult.Entities))

	assert.Equal(t, algorithm1.ID, pageResult.Entities[0].ID)
	assert.Equal(t, algorithm1.Name, pageResult.Entities[0].Name)

	assert.Equal(t, algorithm2.ID, pageResult.Entities[1].ID)
	assert.Equal(t, algorithm2.Name, pageResult.Entities[1].Name)

	algorithm1ID := algorithm1.ID
	algorithm1.Name = "New updated name"
	err = genericDAO.Save(algorithm1)
	assert.Nil(t, err)

	updatedAlgo, err := genericDAO.Get(algorithm1.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, algorithm1.Name, (*(updatedAlgo)).Name)
	assert.Equal(t, algorithm1ID, algorithm1.ID) // Make sure ID doesnt change when the entity key changes

	err = genericDAO.Delete(algorithm1)
	assert.Nil(t, err)

	deletedAlgo, err := genericDAO.Get(algorithm1.ID, common.CONSISTENCY_LOCAL)
	assert.Equal(t, datastore.ErrNotFound, err)
	assert.Nil(t, deletedAlgo)
}

func TestAlgorithmGetPage(t *testing.T) {

	ClusterID = 2

	genericDAO := NewGenericRaftDAO[*config.Algorithm](IntegrationTestCluster.app.Logger,
		IntegrationTestCluster.GetRaftNode1(), ClusterID)
	assert.NotNil(t, genericDAO)

	// Add a new algorithm raft to each of the local cluster nodes
	genericDAO.StartLocalCluster(IntegrationTestCluster, true)

	//numberOfAlgorithmsToCreate := 5000
	numberOfAlgorithmsToCreate := 100
	entities := make([]*config.Algorithm, numberOfAlgorithmsToCreate)
	for i := 0; i < numberOfAlgorithmsToCreate; i++ {
		name := fmt.Sprintf("Test Algorithm %d", i)
		algo := &config.Algorithm{Name: name}
		err := genericDAO.Save(algo)
		assert.Nil(t, err)
		entities[i] = algo
	}

	pageSize := 5

	page1, err := genericDAO.GetPage(query.PageQuery{Page: 1, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[0].ID, page1.Entities[0].ID)
	assert.Equal(t, true, page1.HasMore)

	page2, err := genericDAO.GetPage(query.PageQuery{Page: 2, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2.Entities))
	assert.Equal(t, entities[5].ID, page2.Entities[0].ID)
	assert.Equal(t, true, page2.HasMore)

	page3, err := genericDAO.GetPage(query.PageQuery{Page: 3, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3.Entities))
	assert.Equal(t, entities[10].ID, page1.Entities[3].ID)
	assert.Equal(t, true, page3.HasMore)

	pageSize = 10

	page1, err = genericDAO.GetPage(query.PageQuery{Page: 1, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[0].ID, page1.Entities[0].ID)
	assert.Equal(t, true, page1.HasMore)

	page2, err = genericDAO.GetPage(query.PageQuery{Page: 2, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2.Entities))
	assert.Equal(t, entities[10].ID, page2.Entities[0].ID)
	assert.Equal(t, true, page2.HasMore)

	page3, err = genericDAO.GetPage(query.PageQuery{Page: 3, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3.Entities))
	assert.Equal(t, entities[20].ID, page3.Entities[0].ID)
	assert.Equal(t, true, page1.HasMore)

	pageSize = 1

	page1, err = genericDAO.GetPage(query.PageQuery{Page: 100, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[99].ID, page1.Entities[0].ID)
	assert.Equal(t, false, page1.HasMore)
}
