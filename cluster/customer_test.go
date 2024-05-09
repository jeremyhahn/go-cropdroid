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

func TestCustomerCRUD(t *testing.T) {

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	consistencyLevel := common.CONSISTENCY_LOCAL
	testCustomerName := "test"
	testCustomerEmail := "root@localhost"

	err := Cluster.CreateCustomerCluster()
	assert.Nil(t, err)

	customerDAO := NewRaftCustomerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), CustomerClusterID)
	assert.NotNil(t, customerDAO)

	customer := config.Customer{
		ID:    idGenerator.NewCustomerID(testCustomerEmail),
		Name:  testCustomerName,
		Email: testCustomerEmail}
	err = customerDAO.Save(&customer)
	assert.Nil(t, err)

	persistedCustomer1, err := customerDAO.GetByEmail(customer.Email, consistencyLevel)
	assert.Nil(t, err)
	assert.Equal(t, customer.Email, persistedCustomer1.Email)

	err = customerDAO.Delete(&customer)
	assert.Nil(t, err)

	persistedCustomer2, err := customerDAO.GetByEmail(customer.Email, consistencyLevel)
	assert.Empty(t, persistedCustomer2.ID)
	assert.Equal(t, err, datastore.ErrNotFound)
}
