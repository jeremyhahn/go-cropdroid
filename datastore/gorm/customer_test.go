package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/stretchr/testify/assert"
)

func TestCustomer_CRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.AddressStruct{})
	currentTest.gorm.AutoMigrate(&config.ShippingAddressStruct{})
	currentTest.gorm.AutoMigrate(&config.CustomerStruct{})

	customerDAO := NewGenericGormDAO[*config.CustomerStruct](currentTest.logger, currentTest.gorm)
	customer1 := &config.CustomerStruct{
		ProcessorID: "123",
		Name:        "admin",
		Email:       "customer1@test.com"}
	err := customerDAO.Save(customer1)
	assert.Nil(t, err)

	customer2 := &config.CustomerStruct{
		ProcessorID: "456",
		Name:        "analyst",
		Email:       "customer2@test.com"}
	err = customerDAO.Save(customer2)
	assert.Nil(t, err)

	customer3 := &config.CustomerStruct{
		ProcessorID: "789",
		Name:        "cultivator",
		Email:       "customer3@test.com"}
	err = customerDAO.Save(customer3)
	assert.Nil(t, err)

	// customerByProcessorID, err := customerDAO.GetByProcessorID("456", common.CONSISTENCY_LOCAL)
	// assert.Nil(t, err)
	// assert.NotNil(t, customerByProcessorID)
	// assert.Equal(t, customerByProcessorID.ID, customer2.ID)

	page1, err := customerDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(page1.Entities))
	assert.Equal(t, "admin", page1.Entities[0].Name)
	assert.Equal(t, "analyst", page1.Entities[1].Name)
	assert.Equal(t, "cultivator", page1.Entities[2].Name)
}

func TestCustomer_CustomerDoesntExist_ReturnsNull(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.AddressStruct{})
	currentTest.gorm.AutoMigrate(&config.ShippingAddressStruct{})
	currentTest.gorm.AutoMigrate(&config.CustomerStruct{})

	customerDAO := NewCustomerDAO(currentTest.logger, currentTest.gorm)

	customerConfig, err := customerDAO.Get(uint64(1), common.CONSISTENCY_LOCAL)
	assert.NotNil(t, err)
	assert.Empty(t, customerConfig.ID)
}

// func TestCustomer_GetByUserAndOrgID_MultiCustomer(t *testing.T) {

// 	currentTest := NewIntegrationTest()
// 	defer currentTest.Cleanup()

// 	currentTest.gorm.AutoMigrate(&config.Permission{})
// 	currentTest.gorm.AutoMigrate(&config.User{})
// 	currentTest.gorm.AutoMigrate(&config.Customer{})

// 	customerDAO := NewCustomerDAO(currentTest.logger, currentTest.gorm)
// 	customerDAO.Save(&config.Customer{
// 		ID:   1,
// 		Name: "admin"})
// 	customerDAO.Save(&config.Customer{
// 		ID:   2,
// 		Name: "cultivator"})
// 	customerDAO.Save(&config.Customer{
// 		ID:   3,
// 		Name: "analyst"})

// 	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
// 	userDAO.Save(&config.User{
// 		ID:       1,
// 		Email:    "root@localhost",
// 		Password: "foo"})

// 	currentTest.gorm.Create(&config.Permission{
// 		UserID:         1,
// 		CustomerID:     1,
// 		OrganizationID: 0})
// 	currentTest.gorm.Create(&config.Permission{
// 		UserID:         1,
// 		CustomerID:     3,
// 		OrganizationID: 0})

// 	customers, err := customerDAO.GetAll(common.CONSISTENCY_LOCAL)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 3, len(customers))
// 	assert.Equal(t, "admin", customers[0].GetName())
// 	assert.Equal(t, "analyst", customers[1].GetName())
// 	assert.Equal(t, "cultivator", customers[2].GetName())
// }

// func TestCustomer_GetAll(t *testing.T) {

// 	currentTest := NewIntegrationTest()
// 	defer currentTest.Cleanup()

// 	currentTest.gorm.AutoMigrate(&config.Customer{})

// 	customerDAO := NewCustomerDAO(currentTest.logger, currentTest.gorm)
// 	customerDAO.Save(&config.Customer{
// 		ID:   1,
// 		Name: "admin"})
// 	customerDAO.Save(&config.Customer{
// 		ID:   2,
// 		Name: "cultivator"})
// 	customerDAO.Save(&config.Customer{
// 		ID:   3,
// 		Name: "analyst"})

// 	customers, err := customerDAO.GetAll(common.CONSISTENCY_LOCAL)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 3, len(customers))
// 	assert.Equal(t, "admin", customers[0].GetName())
// 	assert.Equal(t, "analyst", customers[1].GetName())
// 	assert.Equal(t, "cultivator", customers[2].GetName())
// }
