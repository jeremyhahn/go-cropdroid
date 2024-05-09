package gorm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func TestGenericSaveAndSerializeToAndFromJson(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Algorithm{})

	algorithmDAO := NewGenericGormDAO[*config.Algorithm](currentTest.logger, currentTest.gorm)
	assert.NotNil(t, algorithmDAO)

	algorithm1 := config.Algorithm{
		Name: "Test Algorithm 1"}

	algorithm2 := config.Algorithm{
		Name: "Test Algorithm 2"}

	err := algorithmDAO.Save(&algorithm1)
	assert.Nil(t, err)

	err = algorithmDAO.Save(&algorithm2)
	assert.Nil(t, err)

	// Test Save and GetAll
	algorithmConfigs, err := algorithmDAO.GetPage(1, 10, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(algorithmConfigs))
	assert.Equal(t, algorithm1.ID, algorithmConfigs[0].ID)
	assert.Equal(t, algorithm1.Name, algorithmConfigs[0].Name)
	assert.Equal(t, algorithm2.ID, algorithmConfigs[1].ID)
	assert.Equal(t, algorithm2.Name, algorithmConfigs[1].Name)

	// Test JSON marshalling
	jsonAlgorithmConfigs, err := json.Marshal(algorithmConfigs)
	assert.Nil(t, err)
	assert.NotEmpty(t, jsonAlgorithmConfigs)

	var unmarshalledAlgorithms []*config.Algorithm
	err = json.Unmarshal(jsonAlgorithmConfigs, &unmarshalledAlgorithms)
	assert.Nil(t, err)
	assert.Equal(t, len(algorithmConfigs), len(unmarshalledAlgorithms))
	assert.Equal(t, algorithmConfigs[0].ID, unmarshalledAlgorithms[0].ID)
	assert.Equal(t, algorithmConfigs[0].Name, unmarshalledAlgorithms[0].Name)
	assert.Equal(t, algorithmConfigs[1].ID, unmarshalledAlgorithms[1].ID)
	assert.Equal(t, algorithmConfigs[1].Name, unmarshalledAlgorithms[1].Name)
}

func TestGenericUpdateAndDelete(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Algorithm{})

	algorithmDAO := NewGenericGormDAO[*config.Algorithm](currentTest.logger, currentTest.gorm)
	assert.NotNil(t, algorithmDAO)

	algorithm1 := config.Algorithm{
		Name: "Test Algorithm 1"}

	err := algorithmDAO.Save(&algorithm1)
	assert.Nil(t, err)

	savedAlgo, err := algorithmDAO.Get(algorithm1.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, algorithm1.ID, savedAlgo.ID)
	assert.Equal(t, algorithm1.Name, savedAlgo.Name)

	newName := "updated name"
	savedAlgo.Name = newName
	err = algorithmDAO.Update(savedAlgo)
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

func TestGenericGetPage(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Algorithm{})

	algorithmDAO := NewGenericGormDAO[*config.Algorithm](currentTest.logger, currentTest.gorm)
	assert.NotNil(t, algorithmDAO)

	numberOfAlgorithmsToCreate := 40
	algorithms := make([]config.Algorithm, numberOfAlgorithmsToCreate)
	for i := 0; i < numberOfAlgorithmsToCreate; i++ {
		algo := config.Algorithm{
			ID:   uint64(i),
			Name: fmt.Sprintf("Test Algorithm %d", i)}
		err := algorithmDAO.Save(&algo)
		assert.Nil(t, err)
		algorithms[i] = algo

	}

	pageSize := 5

	page1, err := algorithmDAO.GetPage(1, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, uint64(1), page1[0].ID)

	page2, err := algorithmDAO.GetPage(2, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2))
	assert.Equal(t, uint64(6), page2[0].ID)

	page3, err := algorithmDAO.GetPage(3, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3))
	assert.Equal(t, uint64(11), page3[0].ID)

	pageSize = 10

	page1, err = algorithmDAO.GetPage(1, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1))
	assert.Equal(t, uint64(1), page1[0].ID)

	page2, err = algorithmDAO.GetPage(2, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2))
	assert.Equal(t, uint64(11), page2[0].ID)

	page3, err = algorithmDAO.GetPage(3, pageSize, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3))
	assert.Equal(t, uint64(21), page3[0].ID)
}
