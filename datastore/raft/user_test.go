//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/stretchr/testify/assert"
)

func TestUserCRUD(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	consistencyLevel := common.CONSISTENCY_LOCAL
	testUserName := "root@localhost"

	userDAO := NewGenericRaftDAO[*config.User](
		IntegrationTestCluster.app.Logger,
		raftNode1,
		UserClusterID)
	assert.NotNil(t, userDAO)
	userDAO.StartLocalCluster(IntegrationTestCluster, true)

	user := &config.User{
		ID:    raftNode1.GetParams().IdGenerator.NewStringID(testUserName),
		Email: testUserName}
	err := userDAO.Save(user)
	assert.Nil(t, err)

	persistedUser, err := userDAO.Get(user.ID, consistencyLevel)
	assert.Nil(t, err)
	assert.Equal(t, testUserName, persistedUser.GetEmail())

	err = userDAO.Delete(persistedUser)
	assert.Nil(t, err)

	persistedUser2, err := userDAO.Get(user.ID, consistencyLevel)
	assert.Nil(t, persistedUser2)
	assert.Equal(t, err, datastore.ErrNotFound)
}
