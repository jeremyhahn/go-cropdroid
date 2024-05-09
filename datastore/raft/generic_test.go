package raft

import (
	"fmt"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"

	"github.com/stretchr/testify/assert"
)

// TestEntity is a standard time series entity using the ID field
type TestEntity struct {
	ID                    uint64 `yaml:"id" json:"id"`
	Name                  string `yaml:"name" json:"name"`
	config.KeyValueEntity `yaml:"-" json:"-"`
}

func NewTestEntity() *TestEntity {
	return new(TestEntity)
}

func (te *TestEntity) SetID(id uint64) {
	te.ID = id
}

func (te *TestEntity) Identifier() uint64 {
	return te.ID
}

// TestEntityWithTimeSeriesIndex is a cusotm entity whose ID field is set
// to an arbitary ID, and uses the Created field to store the timestamp value
// of when the entity was created. The Raft DAO and state machines will see
// that this entity implements the TimeSeries
type TestEntityWithTimeSeriesIndex struct {
	ID                         uint64 `yaml:"id" json:"id"`
	Name                       string `yaml:"name" json:"name"`
	Created                    uint64 `yaml:"created" json:"created"`
	config.TimeSeriesIndexeder `yaml:"-" json:"-"`
}

func NewTestEntityWithTimeSeriesIndex() *TestEntityWithTimeSeriesIndex {
	return new(TestEntityWithTimeSeriesIndex)
}

func (te *TestEntityWithTimeSeriesIndex) SetID(id uint64) {
	te.ID = id
}

func (te *TestEntityWithTimeSeriesIndex) Identifier() uint64 {
	return te.ID
}

func (te *TestEntityWithTimeSeriesIndex) SetTimestamp(timestamp uint64) {
	te.Created = timestamp
}

func (te *TestEntityWithTimeSeriesIndex) Timestamp() uint64 {
	return te.Created
}

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

	algorithmConfigs, err := genericDAO.GetPage(1, 10, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(algorithmConfigs))

	assert.Equal(t, testEntity1.ID, algorithmConfigs[0].ID)
	assert.Equal(t, testEntity1.Name, algorithmConfigs[0].Name)

	assert.Equal(t, testEntity2.ID, algorithmConfigs[1].ID)
	assert.Equal(t, testEntity2.Name, algorithmConfigs[1].Name)

	testEntity1ID := testEntity1.ID
	testEntity1.Name = "New updated name"
	err = genericDAO.Update(testEntity1)
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

	// Add a new algorithm raft to each of the local cluster nodes
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

	page1, err := genericDAO.GetPage(1, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, entities[0].ID, page1[0].ID)

	page2, err := genericDAO.GetPage(2, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2))
	assert.Equal(t, entities[5].ID, page2[0].ID)

	page3, err := genericDAO.GetPage(3, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3))
	assert.Equal(t, entities[10].ID, page3[0].ID)

	pageSize = 10

	page1, err = genericDAO.GetPage(1, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, entities[0].ID, page1[0].ID)

	page2, err = genericDAO.GetPage(2, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2))
	assert.Equal(t, entities[10].ID, page2[0].ID)

	page3, err = genericDAO.GetPage(3, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3))
	assert.Equal(t, entities[20].ID, page3[0].ID)

	pageSize = 1

	page1, err = genericDAO.GetPage(100, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, entities[99].ID, page1[0].ID)
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

	algorithmConfigs, err := genericDAO.GetPage(1, 10, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(algorithmConfigs))

	assert.Equal(t, testEntity1.ID, algorithmConfigs[0].ID)
	assert.Equal(t, testEntity1.Name, algorithmConfigs[0].Name)

	assert.Equal(t, testEntity2.ID, algorithmConfigs[1].ID)
	assert.Equal(t, testEntity2.Name, algorithmConfigs[1].Name)

	testEntity1ID := testEntity1.ID
	testEntity1.Name = "New updated name"
	err = genericDAO.Update(testEntity1)
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

	page1, err := genericDAO.GetPage(1, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, idGenerator.NewID([]byte(entities[0].Name)), page1[0].ID)

	page2, err := genericDAO.GetPage(2, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2))
	assert.Equal(t, idGenerator.NewID([]byte(entities[5].Name)), page2[0].ID)

	page3, err := genericDAO.GetPage(3, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3))
	assert.Equal(t, idGenerator.NewID([]byte(entities[10].Name)), page3[0].ID)

	pageSize = 10

	page1, err = genericDAO.GetPage(1, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, idGenerator.NewID([]byte(entities[0].Name)), page1[0].ID)

	page2, err = genericDAO.GetPage(2, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2))
	assert.Equal(t, idGenerator.NewID([]byte(entities[10].Name)), page2[0].ID)

	page3, err = genericDAO.GetPage(3, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3))
	assert.Equal(t, idGenerator.NewID([]byte(entities[20].Name)), page3[0].ID)

	pageSize = 1

	page1, err = genericDAO.GetPage(100, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, idGenerator.NewID([]byte(entities[99].Name)), page1[0].ID)
}
