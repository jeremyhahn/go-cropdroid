package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Registration{})

	consistencyLevel := common.CONSISTENCY_LOCAL
	testRegistrationName := "root@localhost"

	registrationDAO := NewRegistrationDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, registrationDAO)

	registration := &config.Registration{
		ID:    currentTest.idGenerator.NewID(testRegistrationName),
		Email: testRegistrationName}
	err := registrationDAO.Save(registration)
	assert.Nil(t, err)

	persistedRegistration, err := registrationDAO.Get(registration.GetID(), consistencyLevel)
	assert.Nil(t, err)
	assert.Equal(t, testRegistrationName, persistedRegistration.GetEmail())

	err = registrationDAO.Delete(persistedRegistration)
	assert.Nil(t, err)

	persistedRegistration2, err := registrationDAO.Get(registration.GetID(), consistencyLevel)
	assert.Nil(t, persistedRegistration2)
	assert.Equal(t, err.Error(), "record not found")
}
