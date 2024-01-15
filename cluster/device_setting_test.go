//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestDeviceSettingCRUD(t *testing.T) {

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)

	deviceDAO := NewRaftDeviceConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmDAO)

	deviceSettingDAO := NewRaftDeviceSettingDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), deviceDAO)

	org := createRaftTestOrganization(t, Cluster,
		ClusterID, serverDAO, userDAO, farmDAO)

	dstest.TestDeviceSettingCRUD(t, deviceDAO, deviceSettingDAO, org)
}
