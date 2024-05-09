//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"
)

func TestRoleCRUD(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	consistencyLevel := common.CONSISTENCY_LOCAL
	testRoleName := "test"

	err := Cluster.CreateRoleCluster()
	assert.Nil(t, err)

	roleDAO := NewRaftRoleDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), RoleClusterID)
	assert.NotNil(t, roleDAO)

	role := &config.Role{
		ID:   idGenerator.NewStringID(testRoleName),
		Name: testRoleName}
	err = roleDAO.Save(role)
	assert.Nil(t, err)

	persistedRole1, err := roleDAO.GetByName(role.GetName(), consistencyLevel)
	assert.Nil(t, err)
	assert.Equal(t, role.GetName(), persistedRole1.GetName())

	err = roleDAO.Delete(role)
	assert.Nil(t, err)

	persistedRole2, err := roleDAO.GetByName(role.GetName(), consistencyLevel)
	assert.Nil(t, persistedRole2)
	assert.Equal(t, err, datastore.ErrNotFound)
}
