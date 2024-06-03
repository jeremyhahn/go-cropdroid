package gorm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func TestAlgorithmSaveAndSerializeToAndFromJson(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.AlgorithmStruct{})

	algorithmDAO := NewGenericGormDAO[*config.AlgorithmStruct](currentTest.logger, currentTest.gorm)
	assert.NotNil(t, algorithmDAO)

	algorithm1 := config.AlgorithmStruct{
		Name: "Test Algorithm 1"}

	algorithm2 := config.AlgorithmStruct{
		Name: "Test Algorithm 2"}

	err := algorithmDAO.Save(&algorithm1)
	assert.Nil(t, err)

	err = algorithmDAO.Save(&algorithm2)
	assert.Nil(t, err)

	// Test Save and GetAll
	pageResult, err := algorithmDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(pageResult.Entities))
	assert.Equal(t, algorithm1.ID, pageResult.Entities[0].ID)
	assert.Equal(t, algorithm1.Name, pageResult.Entities[0].Name)
	assert.Equal(t, algorithm2.ID, pageResult.Entities[1].ID)
	assert.Equal(t, algorithm2.Name, pageResult.Entities[1].Name)

	// Test JSON marshalling
	jsonAlgorithmConfigs, err := json.Marshal(pageResult.Entities)
	assert.Nil(t, err)
	assert.NotEmpty(t, jsonAlgorithmConfigs)

	var unmarshalledAlgorithms []*config.AlgorithmStruct
	err = json.Unmarshal(jsonAlgorithmConfigs, &unmarshalledAlgorithms)
	assert.Nil(t, err)
	assert.Equal(t, len(pageResult.Entities), len(unmarshalledAlgorithms))
	assert.Equal(t, pageResult.Entities[0].ID, unmarshalledAlgorithms[0].ID)
	assert.Equal(t, pageResult.Entities[0].Name, unmarshalledAlgorithms[0].Name)
	assert.Equal(t, pageResult.Entities[1].ID, unmarshalledAlgorithms[1].ID)
	assert.Equal(t, pageResult.Entities[1].Name, unmarshalledAlgorithms[1].Name)
}

func TestAlgorithmUpdateAndDelete(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.AlgorithmStruct{})

	algorithmDAO := NewGenericGormDAO[*config.AlgorithmStruct](currentTest.logger, currentTest.gorm)
	assert.NotNil(t, algorithmDAO)

	algorithm1 := config.AlgorithmStruct{
		Name: "Test Algorithm 1"}

	err := algorithmDAO.Save(&algorithm1)
	assert.Nil(t, err)

	savedAlgo, err := algorithmDAO.Get(algorithm1.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, algorithm1.ID, savedAlgo.ID)
	assert.Equal(t, algorithm1.Name, savedAlgo.Name)

	newName := "updated name"
	savedAlgo.Name = newName
	err = algorithmDAO.Save(savedAlgo)
	assert.Nil(t, err)

	updatedAlgo, err := algorithmDAO.Get(savedAlgo.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, updatedAlgo.ID, savedAlgo.ID)
	assert.Equal(t, updatedAlgo.Name, newName)

	err = algorithmDAO.Delete(updatedAlgo)
	assert.Nil(t, err)

	_, err = algorithmDAO.Get(updatedAlgo.ID, common.CONSISTENCY_LOCAL)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestAlgorithmGetPage(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.AlgorithmStruct{})

	algorithmDAO := NewGenericGormDAO[*config.AlgorithmStruct](currentTest.logger, currentTest.gorm)
	assert.NotNil(t, algorithmDAO)

	numberOfAlgorithmsToCreate := 40
	algorithms := make([]config.AlgorithmStruct, numberOfAlgorithmsToCreate)
	for i := 0; i < numberOfAlgorithmsToCreate; i++ {
		algo := config.AlgorithmStruct{
			ID:   uint64(i),
			Name: fmt.Sprintf("Test Algorithm %d", i)}
		err := algorithmDAO.Save(&algo)
		assert.Nil(t, err)
		algorithms[i] = algo
	}

	pageSize := 5

	page1, err := algorithmDAO.GetPage(query.PageQuery{Page: 1, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, uint64(1), page1.Entities[0].ID)

	page2, err := algorithmDAO.GetPage(query.PageQuery{Page: 2, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2.Entities))
	assert.Equal(t, uint64(6), page2.Entities[0].ID)

	page3, err := algorithmDAO.GetPage(query.PageQuery{Page: 3, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3.Entities))
	assert.Equal(t, uint64(11), page3.Entities[0].ID)

	pageSize = 10

	page1, err = algorithmDAO.GetPage(query.PageQuery{Page: 1, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, uint64(1), page1.Entities[0].ID)

	page2, err = algorithmDAO.GetPage(query.PageQuery{Page: 2, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2.Entities))
	assert.Equal(t, uint64(11), page2.Entities[0].ID)

	page3, err = algorithmDAO.GetPage(query.PageQuery{Page: 3, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3.Entities))
	assert.Equal(t, uint64(21), page3.Entities[0].ID)
}
