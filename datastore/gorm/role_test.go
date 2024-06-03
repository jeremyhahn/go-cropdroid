package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/stretchr/testify/assert"
)

func TestRole_GetByUserAndOrgID_SingleRole(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})

	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	roleDAO.Save(&config.RoleStruct{
		ID:   1,
		Name: "admin"})
	roleDAO.Save(&config.RoleStruct{
		ID:   2,
		Name: "cultivator"})
	roleDAO.Save(&config.RoleStruct{
		ID:   3,
		Name: "analyst"})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	userDAO.Save(&config.UserStruct{
		ID:       1,
		Email:    "root@localhost",
		Password: "foo"})

	currentTest.gorm.Create(&config.PermissionStruct{
		UserID:         1,
		RoleID:         1,
		OrganizationID: 0})

	page1, err := roleDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(page1.Entities))
	assert.Equal(t, "admin", page1.Entities[0].GetName())
	assert.Equal(t, "cultivator", page1.Entities[1].GetName())
	assert.Equal(t, "analyst", page1.Entities[2].GetName())
}

func TestRole_GetByUserAndOrgID_MultiRole(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.PermissionStruct{})
	currentTest.gorm.AutoMigrate(&config.UserStruct{})
	currentTest.gorm.AutoMigrate(&config.RoleStruct{})

	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	roleDAO.Save(&config.RoleStruct{
		ID:   1,
		Name: "admin"})
	roleDAO.Save(&config.RoleStruct{
		ID:   2,
		Name: "cultivator"})
	roleDAO.Save(&config.RoleStruct{
		ID:   3,
		Name: "analyst"})

	userDAO := NewUserDAO(currentTest.logger, currentTest.gorm)
	userDAO.Save(&config.UserStruct{
		ID:       1,
		Email:    "root@localhost",
		Password: "foo"})

	currentTest.gorm.Create(&config.PermissionStruct{
		UserID:         1,
		RoleID:         1,
		OrganizationID: 0})
	currentTest.gorm.Create(&config.PermissionStruct{
		UserID:         1,
		RoleID:         3,
		OrganizationID: 0})

	page1, err := roleDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(page1.Entities))
	assert.Equal(t, "admin", page1.Entities[0].GetName())
	assert.Equal(t, "cultivator", page1.Entities[1].GetName())
	assert.Equal(t, "analyst", page1.Entities[2].GetName())
}

func TestRole_GetAll(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.RoleStruct{})

	roleDAO := NewRoleDAO(currentTest.logger, currentTest.gorm)
	roleDAO.Save(&config.RoleStruct{
		ID:   1,
		Name: "admin"})
	roleDAO.Save(&config.RoleStruct{
		ID:   2,
		Name: "cultivator"})
	roleDAO.Save(&config.RoleStruct{
		ID:   3,
		Name: "analyst"})

	page1, err := roleDAO.GetPage(query.NewPageQuery(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(page1.Entities))
	assert.Equal(t, "admin", page1.Entities[0].GetName())
	assert.Equal(t, "cultivator", page1.Entities[1].GetName())
	assert.Equal(t, "analyst", page1.Entities[2].GetName())
}
