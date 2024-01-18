//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestAlgorithmCRUD(t *testing.T) {

	err := Cluster.CreateAlgorithmCluster()
	assert.Nil(t, err)

	algorithmDAO := NewRaftAlgorithmDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), AlgorithmClusterID)
	assert.NotNil(t, algorithmDAO)

	algorithm1 := config.NewAlgorithm()
	algorithm1.SetName("Test Algorithm 1")

	algorithm2 := config.NewAlgorithm()
	algorithm2.SetName("Test Algorithm 2")

	err = algorithmDAO.Save(algorithm1)
	assert.Nil(t, err)

	err = algorithmDAO.Save(algorithm2)
	assert.Nil(t, err)

	algorithmConfigs, err := algorithmDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(algorithmConfigs))

	assert.Equal(t, algorithm1.ID, algorithmConfigs[1].GetID())
	assert.Equal(t, algorithm1.Name, algorithmConfigs[1].GetName())

	assert.Equal(t, algorithm2.ID, algorithmConfigs[0].GetID())
	assert.Equal(t, algorithm2.Name, algorithmConfigs[0].GetName())
}
