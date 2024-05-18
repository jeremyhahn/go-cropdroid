//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationCRUD(t *testing.T) {

	consistencyLevel := common.CONSISTENCY_LOCAL
	testRegistrationName := "root@localhost"

	err := Cluster.CreateRegistrationCluster()
	assert.Nil(t, err)

	registrationDAO := NewRaftRegistrationDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), RegistrationClusterID)
	assert.NotNil(t, registrationDAO)

	registration := &config.Registration{
		ID:    Cluster.app.IdGenerator.NewStringID(testRegistrationName),
		Email: testRegistrationName}
	err = registrationDAO.Save(registration)
	assert.Nil(t, err)

	persistedRegistration, err := registrationDAO.Get(registration.ID, consistencyLevel)
	assert.Nil(t, err)
	assert.Equal(t, testRegistrationName, persistedRegistration.GetEmail())

	err = registrationDAO.Delete(persistedRegistration)
	assert.Nil(t, err)

	persistedRegistration2, err := registrationDAO.Get(registration.ID, consistencyLevel)
	assert.Nil(t, persistedRegistration2)
	assert.Equal(t, err.Error(), "not found")
}
