package mocks

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/mock"
)

type MockOrganizationDAO struct {
	dao.OrganizationDAO
	mock.Mock
}

func NewMockOrganizationDAO() *MockOrganizationDAO {
	return &MockOrganizationDAO{}
}

func (dao *MockOrganizationDAO) Create(organization config.Organization) error {
	args := dao.Called(organization)
	fmt.Println("Creating organization record")
	return args.Error(0)
}

func (dao *MockOrganizationDAO) Save(organization config.Organization) error {
	args := dao.Called(organization)
	fmt.Println("Saving organization record")
	return args.Error(0)
}

func (dao *MockOrganizationDAO) Update(organization config.Organization) error {
	args := dao.Called(organization)
	fmt.Println("Updating organization record")
	return args.Error(0)
}

func (dao *MockOrganizationDAO) Get(name string) (config.Organization, error) {
	args := dao.Called(name)
	fmt.Printf("Getting organization record '%s'\n", name)
	return args.Get(0).(config.Organization), args.Error(1)
}

func (dao *MockOrganizationDAO) First() (config.Organization, error) {
	args := dao.Called()
	fmt.Println("Updating organization record")
	return args.Get(0).(config.Organization), args.Error(1)
}

func (dao *MockOrganizationDAO) GetByID(id string) (config.Organization, error) {
	args := dao.Called(id)
	fmt.Printf("Getting organization id '%s'\n", id)
	return args.Get(0).(config.Organization), nil
}

func (dao *MockOrganizationDAO) GetByUserID(userID int) ([]config.Organization, error) {
	args := dao.Called(userID)
	fmt.Printf("Getting organization id for user %d\n", userID)
	return args.Get(0).([]config.Organization), nil
}
