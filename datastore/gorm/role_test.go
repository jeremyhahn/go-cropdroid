package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/stretchr/testify/assert"
)

func TestRole_GetByUserAndOrgID_SingleRole(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

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
	userDAO.Save(&config.User{
		ID:       1,
		Email:    "root@localhost",
		Password: "foo"})

	currentTest.gorm.Create(&config.Permission{
		UserID:         1,
		RoleID:         1,
		OrganizationID: 0})

	roles, err := roleDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(roles))
	assert.Equal(t, "admin", roles[0].GetName())
	assert.Equal(t, "analyst", roles[1].GetName())
	assert.Equal(t, "cultivator", roles[2].GetName())
}

func TestRole_GetByUserAndOrgID_MultiRole(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

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
	userDAO.Save(&config.User{
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

	roles, err := roleDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(roles))
	assert.Equal(t, "admin", roles[0].GetName())
	assert.Equal(t, "analyst", roles[1].GetName())
	assert.Equal(t, "cultivator", roles[2].GetName())
}

func TestRole_GetAll(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

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

	roles, err := roleDAO.GetAll(common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(roles))
	assert.Equal(t, "admin", roles[0].GetName())
	assert.Equal(t, "analyst", roles[1].GetName())
	assert.Equal(t, "cultivator", roles[2].GetName())
}
