//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"fmt"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationCRUD(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	testRegistrationName := "root@localhost"
	testRegistrationName2 := "root2@localhost"

	registrationDAO := NewGenericRaftDAO[*config.RegistrationStruct](
		IntegrationTestCluster.app.Logger,
		raftNode1,
		RegistrationClusterID)
	assert.NotNil(t, registrationDAO)
	registrationDAO.StartLocalCluster(IntegrationTestCluster, true)

	registration := &config.RegistrationStruct{
		Email: testRegistrationName}
	err := registrationDAO.Save(registration)
	assert.Nil(t, err)

	registration2 := &config.RegistrationStruct{
		Email: testRegistrationName2}
	err = registrationDAO.Save(registration2)
	assert.Nil(t, err)

	persistedRegistration, err := registrationDAO.Get(registration.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, registration.ID, persistedRegistration.ID)
	assert.Equal(t, testRegistrationName, persistedRegistration.GetEmail())

	page1, err := registrationDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(page1.Entities))
	assert.Equal(t, registration.ID, page1.Entities[0].Identifier())
	assert.Equal(t, registration.Email, page1.Entities[0].GetEmail())
	assert.Equal(t, registration2.ID, page1.Entities[1].Identifier())
	assert.Equal(t, registration2.Email, page1.Entities[1].GetEmail())

	registrationID := registration.ID
	registration.Email = "root@example.com"
	err = registrationDAO.Save(registration)
	assert.Nil(t, err)

	updatedAlgo, err := registrationDAO.Get(registration.ID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, registration.Email, updatedAlgo.GetEmail())
	assert.Equal(t, registrationID, registration.ID) // Make sure ID doesnt change when the entity key changes

	err = registrationDAO.Delete(registration)
	assert.Nil(t, err)

	deletedAlgo, err := registrationDAO.Get(registration.ID, common.CONSISTENCY_LOCAL)
	assert.Equal(t, datastore.ErrRecordNotFound, err)
	assert.Nil(t, deletedAlgo)
}

func TestRegistrationGetPage(t *testing.T) {

	ClusterID = 2

	registrationDAO := NewGenericRaftDAO[*config.RegistrationStruct](IntegrationTestCluster.app.Logger,
		IntegrationTestCluster.GetRaftNode1(), ClusterID)
	assert.NotNil(t, registrationDAO)
	registrationDAO.StartLocalCluster(IntegrationTestCluster, true)

	//numberOfRegistrationsToCreate := 5000
	numberOfRegistrationsToCreate := 100
	entities := make([]*config.RegistrationStruct, numberOfRegistrationsToCreate)
	for i := 0; i < numberOfRegistrationsToCreate; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		algo := &config.RegistrationStruct{Email: email}
		err := registrationDAO.Save(algo)
		assert.Nil(t, err)
		entities[i] = algo
	}

	pageSize := 5

	page1, err := registrationDAO.GetPage(query.PageQuery{Page: 1, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[0].ID, page1.Entities[0].ID)

	page2, err := registrationDAO.GetPage(query.PageQuery{Page: 2, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2.Entities))
	assert.Equal(t, entities[5].ID, page2.Entities[0].ID)

	page3, err := registrationDAO.GetPage(query.PageQuery{Page: 3, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3.Entities))
	assert.Equal(t, entities[10].ID, page3.Entities[0].ID)

	pageSize = 10

	page1, err = registrationDAO.GetPage(query.PageQuery{Page: 1, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[0].ID, page1.Entities[0].ID)

	page2, err = registrationDAO.GetPage(query.PageQuery{Page: 2, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page2.Entities))
	assert.Equal(t, entities[10].ID, page2.Entities[0].ID)

	page3, err = registrationDAO.GetPage(query.PageQuery{Page: 3, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page3.Entities))
	assert.Equal(t, entities[20].ID, page3.Entities[0].ID)

	pageSize = 1

	page1, err = registrationDAO.GetPage(query.PageQuery{Page: 100, PageSize: pageSize}, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, pageSize, len(page1.Entities))
	assert.Equal(t, entities[99].ID, page1.Entities[0].ID)
}
