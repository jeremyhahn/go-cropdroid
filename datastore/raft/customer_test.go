//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"
)

func TestCustomerCRUD(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	consistencyLevel := common.CONSISTENCY_LOCAL
	testCustomerName := "test"
	testCustomerEmail := "root@localhost"
	raftNode1 := IntegrationTestCluster.GetRaftNode1()

	customerDAO := NewRaftCustomerDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		CustomerClusterID)
	assert.NotNil(t, customerDAO)
	customerDAO.StartLocalCluster(IntegrationTestCluster, true)

	customer := config.Customer{
		ID:    idGenerator.NewCustomerID(testCustomerEmail),
		Name:  testCustomerName,
		Email: testCustomerEmail}
	err := customerDAO.Save(&customer)
	assert.Nil(t, err)

	persistedCustomer1, err := customerDAO.GetByEmail(customer.Email, consistencyLevel)
	assert.Nil(t, err)
	assert.Equal(t, customer.Email, persistedCustomer1.Email)

	err = customerDAO.Delete(&customer)
	assert.Nil(t, err)

	persistedCustomer2, err := customerDAO.GetByEmail(customer.Email, consistencyLevel)
	assert.Equal(t, datastore.ErrNotFound, err)
	assert.Nil(t, persistedCustomer2)
}
