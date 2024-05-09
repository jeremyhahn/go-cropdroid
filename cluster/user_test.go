//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/stretchr/testify/assert"
)

func TestUserCRUD(t *testing.T) {

	consistencyLevel := common.CONSISTENCY_LOCAL
	testUserName := "root@localhost"

	err := Cluster.CreateUserCluster()
	assert.Nil(t, err)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)
	assert.NotNil(t, userDAO)

	user := &config.User{
		ID:    Cluster.app.IdGenerator.NewStringID(testUserName),
		Email: testUserName}
	err = userDAO.Save(user)
	assert.Nil(t, err)

	persistedUser, err := userDAO.Get(user.GetID(), consistencyLevel)
	assert.Nil(t, err)
	assert.Equal(t, testUserName, persistedUser.GetEmail())

	err = userDAO.Delete(persistedUser)
	assert.Nil(t, err)

	persistedUser2, err := userDAO.Get(user.GetID(), consistencyLevel)
	assert.Nil(t, persistedUser2)
	assert.Equal(t, err, datastore.ErrNotFound)
}
