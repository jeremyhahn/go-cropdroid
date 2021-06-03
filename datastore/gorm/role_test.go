package gorm

import (
	"testing"

	"github.com/jeremyhahn/cropdroid/config"
	"github.com/stretchr/testify/assert"
)

func TestRole_GetByUserAndOrgID_SingleRole(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	roleDAO.Save(&config.Role{
		ID:   1,
		Name: "admin"})
	roleDAO.Save(&config.Role{
		ID:   2,
		Name: "cultivator"})
	roleDAO.Save(&config.Role{
		ID:   3,
		Name: "analyst"})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	userDAO.Create(&config.User{
		ID:       1,
		Email:    "root@localhost",
		Password: "foo"})

	currentTest.gorm.Create(&config.Permission{
		UserID:         1,
		RoleID:         1,
		OrganizationID: 0})

	roles, err := roleDAO.GetByUserAndOrgID(1, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, "admin", roles[0].GetName())

	currentTest.Cleanup()
}

func TestRole_GetByUserAndOrgID_MultiRole(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Permission{})
	currentTest.gorm.AutoMigrate(&config.User{})
	currentTest.gorm.AutoMigrate(&config.Role{})

	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	roleDAO.Save(&config.Role{
		ID:   1,
		Name: "admin"})
	roleDAO.Save(&config.Role{
		ID:   2,
		Name: "cultivator"})
	roleDAO.Save(&config.Role{
		ID:   3,
		Name: "analyst"})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	userDAO.Create(&config.User{
		ID:       1,
		Email:    "root@localhost",
		Password: "foo"})

	currentTest.gorm.Create(&config.Permission{
		UserID:         1,
		RoleID:         1,
		OrganizationID: 0})
	currentTest.gorm.Create(&config.Permission{
		UserID:         1,
		RoleID:         3,
		OrganizationID: 0})

	roles, err := roleDAO.GetByUserAndOrgID(1, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(roles))
	assert.Equal(t, "admin", roles[0].GetName())
	assert.Equal(t, "analyst", roles[1].GetName())

	currentTest.Cleanup()
}
