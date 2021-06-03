package gorm

import (
	"testing"

	"github.com/jeremyhahn/cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestAlgorithmCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.LogMode(true)
	currentTest.gorm.AutoMigrate(&config.Algorithm{})

	algorithmDAO := NewAlgorithmDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, algorithmDAO)

	algorithm1 := config.NewAlgorithm()
	algorithm1.SetName("Test Algorithm 1")

	algorithm2 := config.NewAlgorithm()
	algorithm2.SetName("Test Algorithm 2")

	err := algorithmDAO.Create(algorithm1)
	assert.Nil(t, err)

	err = algorithmDAO.Create(algorithm2)
	assert.Nil(t, err)

	algorithmConfigs, err := algorithmDAO.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(algorithmConfigs))

	assert.Equal(t, algorithm1.ID, algorithmConfigs[0].ID)
	assert.Equal(t, algorithm1.Name, algorithmConfigs[0].Name)

	assert.Equal(t, algorithm2.ID, algorithmConfigs[1].ID)
	assert.Equal(t, algorithm2.Name, algorithmConfigs[1].Name)

	currentTest.Cleanup()
}
