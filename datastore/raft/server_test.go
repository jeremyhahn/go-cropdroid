//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/test/data"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"
)

func TestServerCRUD(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	consistencyLevel := common.CONSISTENCY_LOCAL

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	serverDAO := IntegrationTestCluster.serverDAO
	assert.NotNil(t, serverDAO)

	orgDAO := NewRaftOrganizationDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		OrganizationClusterID,
		serverDAO)
	assert.NotNil(t, orgDAO)
	orgDAO.StartLocalCluster(IntegrationTestCluster, true)

	testOrg1 := data.CreateTestOrganization1(idGenerator)
	orgDAO.Save(testOrg1)

	testOrg2 := data.CreateTestOrganization2(idGenerator)
	orgDAO.Save(testOrg1)

	farm2 := testOrg1.GetFarms()[1]
	farm4 := testOrg2.GetFarms()[1]

	server := &config.Server{
		ID: ClusterID,
		OrganizationRefs: []uint64{
			testOrg1.ID,
			testOrg2.ID},
		FarmRefs: []uint64{
			farm2.ID,
			farm4.ID}}

	err := serverDAO.Save(server)
	assert.Nil(t, err)

	persistedServer, err := serverDAO.Get(server.ID, consistencyLevel)
	assert.Nil(t, err)
	assert.Equal(t, server.GetOrganizationRefs(), persistedServer.GetOrganizationRefs())
	assert.Equal(t, server.GetFarmRefs(), persistedServer.GetFarmRefs())

	err = serverDAO.Delete(persistedServer)
	assert.Nil(t, err)

	persistedServer2, err := serverDAO.Get(server.ID, consistencyLevel)
	assert.Nil(t, persistedServer2)
	assert.Equal(t, err, datastore.ErrRecordNotFound)
}
