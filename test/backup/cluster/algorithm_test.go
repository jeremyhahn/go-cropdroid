//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"

	"github.com/stretchr/testify/assert"
)

func TestAlgorithmCRUD(t *testing.T) {

	err := Cluster.CreateAlgorithmCluster()
	assert.Nil(t, err)

	algorithmDAO := NewRaftAlgorithmDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), AlgorithmClusterID)
	assert.NotNil(t, algorithmDAO)

	algorithm1 := config.NewAlgorithm()
	algorithm1.Name = "Test Algorithm 1"

	algorithm2 := config.NewAlgorithm()
	algorithm2.Name = "Test Algorithm 2"

	err = algorithmDAO.Save(algorithm1)
	assert.Nil(t, err)

	err = algorithmDAO.Save(algorithm2)
	assert.Nil(t, err)

	page1, err := algorithmDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(page1.Entities))

	assert.Equal(t, algorithm1.ID, page1.Entities[1].ID)
	assert.Equal(t, algorithm1.Name, page1.Entities[1].Name)

	assert.Equal(t, algorithm2.ID, page1.Entities[0].ID)
	assert.Equal(t, algorithm2.Name, page1.Entities[0].Name)
}
