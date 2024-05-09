//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"fmt"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"

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

	algorithmConfigs, err := genericDAO.GetPage(1, 10, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(algorithmConfigs))

	assert.Equal(t, algorithm1.ID, (*(algorithmConfigs[0])).ID)
	assert.Equal(t, algorithm1.Name, (*(algorithmConfigs[0])).Name)

	assert.Equal(t, algorithm2.ID, (*(algorithmConfigs[1])).ID)
	assert.Equal(t, algorithm2.Name, (*(algorithmConfigs[1])).Name)

	algorithm1ID := algorithm1.ID
	algorithm1.Name = "New updated name"
	err = genericDAO.Update(algorithm1)
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

	page1, err := genericDAO.GetPage(1, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, entities[0].ID, (*(page1[0])).ID)

	page2, err := genericDAO.GetPage(2, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2))
	assert.Equal(t, entities[5].ID, (*(page2[0])).ID)

	page3, err := genericDAO.GetPage(3, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3))
	assert.Equal(t, entities[10].ID, (*(page3[0])).ID)

	pageSize = 10

	page1, err = genericDAO.GetPage(1, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, entities[0].ID, (*(page1[0])).ID)

	page2, err = genericDAO.GetPage(2, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2))
	assert.Equal(t, entities[10].ID, (*(page2[0])).ID)

	page3, err = genericDAO.GetPage(3, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3))
	assert.Equal(t, entities[20].ID, (*(page3[0])).ID)

	pageSize = 1

	page1, err = genericDAO.GetPage(100, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, entities[99].ID, (*(page1[0])).ID)
}
