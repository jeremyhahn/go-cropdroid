package raft

import (
	"fmt"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"

	"github.com/stretchr/testify/assert"
)

func TestGenericRaftCRUD(t *testing.T) {

	ClusterID = 1

	genericDAO := NewGenericRaftDAO[*TestEntity](IntegrationTestCluster.app.Logger,
		IntegrationTestCluster.GetRaftNode1(), ClusterID)
	assert.NotNil(t, genericDAO)

	// Add a new algorithm raft to each of the local cluster nodes
	genericDAO.StartLocalCluster(IntegrationTestCluster, true)

	testEntity1 := NewTestEntity()
	testEntity1.Name = "Test Entity 1"

	testEntity2 := NewTestEntity()
	testEntity2.Name = "Test Entity 2"

	err := genericDAO.Save(testEntity1)
	assert.Nil(t, err)

	err = genericDAO.Save(testEntity2)
	assert.Nil(t, err)

	page1, err := genericDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(page1.Entities))

	assert.Equal(t, testEntity1.ID, page1.Entities[0].ID)
	assert.Equal(t, testEntity1.Name, page1.Entities[0].Name)

	assert.Equal(t, testEntity2.ID, page1.Entities[1].ID)
	assert.Equal(t, testEntity2.Name, page1.Entities[1].Name)

	testEntity1ID := testEntity1.ID
	testEntity1.Name = "New updated name"
	err = genericDAO.Save(testEntity1)
	assert.Nil(t, err)

	updatedAlgo, err := genericDAO.Get(testEntity1.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, testEntity1.Name, updatedAlgo.Name)
	assert.Equal(t, testEntity1ID, testEntity1.ID) // Make sure ID doesnt change when the entity key changes

	err = genericDAO.Delete(updatedAlgo)
	assert.Nil(t, err)

	deletedAlgo, err := genericDAO.Get(testEntity1.ID, common.CONSISTENCY_LOCAL)
	assert.Equal(t, datastore.ErrNotFound, err)
	assert.Nil(t, deletedAlgo)
}

func TestGenericRaftGetPage(t *testing.T) {

	ClusterID = 2

	genericDAO := NewGenericRaftDAO[*TestEntity](IntegrationTestCluster.app.Logger,
		IntegrationTestCluster.GetRaftNode1(), ClusterID)
	assert.NotNil(t, genericDAO)
	genericDAO.StartLocalCluster(IntegrationTestCluster, true)

	//numberOfAlgorithmsToCreate := 5000
	numberOfAlgorithmsToCreate := 100
	entities := make([]TestEntity, numberOfAlgorithmsToCreate)
	for i := 0; i < numberOfAlgorithmsToCreate; i++ {
		name := fmt.Sprintf("Test Algorithm %d", i)
		te := TestEntity{Name: name}
		err := genericDAO.Save(&te)
		assert.Nil(t, err)
		entities[i] = te
	}

	pageSize := 5

	page1, err := genericDAO.GetPage(query.PageQuery{Page: 1, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[0].ID, page1.Entities[0].ID)

	page2, err := genericDAO.GetPage(query.PageQuery{Page: 2, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2.Entities))
	assert.Equal(t, entities[5].ID, page2.Entities[0].ID)

	page3, err := genericDAO.GetPage(query.PageQuery{Page: 3, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3.Entities))
	assert.Equal(t, entities[10].ID, page3.Entities[0].ID)

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
	assert.Equal(t, true, page3.HasMore)

	page1, err = genericDAO.GetPage(query.PageQuery{Page: 100, PageSize: 1}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[99].ID, page1.Entities[0].ID)
	assert.Equal(t, false, page1.HasMore)
}

func TestGenericRaftCRUDWithTimeSeriesIndex(t *testing.T) {

	ClusterID = 3

	genericDAO := NewGenericRaftDAO[*TestEntityWithTimeSeriesIndex](IntegrationTestCluster.app.Logger,
		IntegrationTestCluster.GetRaftNode1(), ClusterID)
	assert.NotNil(t, genericDAO)

	// Add a new algorithm raft to each of the local cluster nodes
	genericDAO.StartLocalCluster(IntegrationTestCluster, true)

	testEntity1 := NewTestEntityWithTimeSeriesIndex()
	testEntity1.ID = 1
	testEntity1.Name = "Test Entity 1"

	testEntity2 := NewTestEntityWithTimeSeriesIndex()
	testEntity2.ID = 2
	testEntity2.Name = "Test Entity 2"

	err := genericDAO.SaveWithTimeSeriesIndex(testEntity1)
	assert.Nil(t, err)

	err = genericDAO.SaveWithTimeSeriesIndex(testEntity2)
	assert.Nil(t, err)

	page1, err := genericDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(page1.Entities))

	assert.Equal(t, testEntity1.ID, page1.Entities[0].ID)
	assert.Equal(t, testEntity1.Name, page1.Entities[0].Name)

	assert.Equal(t, testEntity2.ID, page1.Entities[1].ID)
	assert.Equal(t, testEntity2.Name, page1.Entities[1].Name)

	testEntity1ID := testEntity1.ID
	testEntity1.Name = "New updated name"
	err = genericDAO.Save(testEntity1)
	assert.Nil(t, err)

	updatedAlgo, err := genericDAO.Get(testEntity1.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, testEntity1.Name, updatedAlgo.Name)
	assert.Equal(t, testEntity1ID, testEntity1.ID) // Make sure ID doesnt change when the entity key changes

	err = genericDAO.Delete(updatedAlgo)
	assert.Nil(t, err)

	deletedAlgo, err := genericDAO.Get(testEntity1.ID, common.CONSISTENCY_LOCAL)
	assert.Equal(t, datastore.ErrNotFound, err)
	assert.Nil(t, deletedAlgo)
}

func TestGenericRaftGetPageWithTimeSeriesIndex(t *testing.T) {

	ClusterID = 4
	idGenerator := IntegrationTestCluster.app.IdGenerator

	genericDAO := NewGenericRaftDAO[*TestEntityWithTimeSeriesIndex](IntegrationTestCluster.app.Logger,
		IntegrationTestCluster.GetRaftNode1(), ClusterID)
	assert.NotNil(t, genericDAO)

	// Add a new algorithm raft to each of the local cluster nodes
	genericDAO.StartLocalCluster(IntegrationTestCluster, true)

	//numberOfAlgorithmsToCreate := 5000
	numberOfAlgorithmsToCreate := 100
	entities := make([]*TestEntityWithTimeSeriesIndex, numberOfAlgorithmsToCreate)
	for i := 0; i < numberOfAlgorithmsToCreate; i++ {
		name := fmt.Sprintf("Test Algorithm %d", i)
		te := &TestEntityWithTimeSeriesIndex{
			ID:   idGenerator.NewID([]byte(name)),
			Name: name}
		err := genericDAO.SaveWithTimeSeriesIndex(te)
		assert.Nil(t, err)
		entities[i] = te
	}

	pageSize := 5

	page1, err := genericDAO.GetPage(query.PageQuery{Page: 1, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[0].ID, page1.Entities[0].ID)

	page2, err := genericDAO.GetPage(query.PageQuery{Page: 2, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2.Entities))
	assert.Equal(t, entities[5].ID, page2.Entities[0].ID)

	page3, err := genericDAO.GetPage(query.PageQuery{Page: 3, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3.Entities))
	assert.Equal(t, entities[10].ID, page3.Entities[0].ID)

	pageSize = 10

	page1, err = genericDAO.GetPage(query.PageQuery{Page: 1, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[0].ID, page1.Entities[0].ID)

	page2, err = genericDAO.GetPage(query.PageQuery{Page: 2, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2.Entities))
	assert.Equal(t, entities[10].ID, page2.Entities[0].ID)

	page3, err = genericDAO.GetPage(query.PageQuery{Page: 3, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3.Entities))
	assert.Equal(t, entities[20].ID, page3.Entities[0].ID)

	page1, err = genericDAO.GetPage(query.PageQuery{Page: 100, PageSize: 1}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[99].ID, page1.Entities[0].ID)
}
