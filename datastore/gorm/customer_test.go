package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/stretchr/testify/assert"
)

func TestCustomer_CRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Address{})
	currentTest.gorm.AutoMigrate(&config.ShippingAddress{})
	currentTest.gorm.AutoMigrate(&config.Customer{})

	customerDAO := NewCustomerDAO(currentTest.logger, currentTest.gorm)
	customer1 := &config.Customer{
		ID:          1,
		ProcessorID: "123",
		Name:        "admin",
		Email:       "customer1@test.com"}
	err := customerDAO.Save(customer1)
	assert.Nil(t, err)

	customer2 := &config.Customer{
		ID:          2,
		ProcessorID: "456",
		Name:        "cultivator",
		Email:       "customer2@test.com"}
	err = customerDAO.Save(customer2)
	assert.Nil(t, err)

	customer3 := &config.Customer{
		ID:          3,
		ProcessorID: "789",
		Name:        "analyst",
		Email:       "customer3@test.com"}
	err = customerDAO.Save(customer3)
	assert.Nil(t, err)

	customerByProcessorID, err := customerDAO.GetByProcessorID("456", common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.NotNil(t, customerByProcessorID)
	assert.Equal(t, customerByProcessorID.ID, customer2.ID)

	allCustomers, err := customerDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allCustomers))
	assert.Equal(t, "admin", allCustomers[0].Name)
	assert.Equal(t, "analyst", allCustomers[1].Name)
	assert.Equal(t, "cultivator", allCustomers[2].Name)
}

func TestCustomer_CustomerDoesntExist_ReturnsNull(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Address{})
	currentTest.gorm.AutoMigrate(&config.ShippingAddress{})
	currentTest.gorm.AutoMigrate(&config.Customer{})

	customerDAO := NewCustomerDAO(currentTest.logger, currentTest.gorm)

	customerConfig, err := customerDAO.Get(uint64(1), common.CONSISTENCY_LOCAL)
	assert.NotNil(t, err)
	assert.Nil(t, customerConfig)
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
