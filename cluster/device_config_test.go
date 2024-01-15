//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCRUD(t *testing.T) {

	err := Cluster.CreateDeviceConfigCluster(DeviceConfigClusterID)
	assert.Nil(t, err)

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)
	assert.NotNil(t, userDAO)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)
	assert.NotNil(t, farmDAO)

	deviceDAO := NewRaftDeviceConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmDAO)
	assert.NotNil(t, deviceDAO)

	org := createRaftTestOrganization(t, Cluster,
		ClusterID, serverDAO, userDAO, farmDAO)

	farm := org.GetFarms()[0]

	dstest.TestDeviceCRUD(t, deviceDAO, farm)
}
