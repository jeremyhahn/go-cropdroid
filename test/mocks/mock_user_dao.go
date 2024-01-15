package mocks

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/mock"
)

type MockUserDAO struct {
	dao.UserDAO
	mock.Mock
}

func NewMockUserDAO() dao.UserDAO {
	return &MockUserDAO{}
}

func (dao *MockUserDAO) GetById(userId int) (*config.User, error) {
	args := dao.Called(userId)
	fmt.Printf("[MockUserDAO] GetById() userId = %d", userId)
	return args.Get(0).(*config.User), nil
}

func (dao *MockUserDAO) GetByEmail(email string) (*config.User, error) {
	args := dao.Called(email)
	fmt.Printf("[MockUserDAO] GetByEmail() email = %s", email)
	return args.Get(0).(*config.User), nil
}

func (dao *MockUserDAO) Create(user *config.User) error {
	args := dao.Called(user)
	fmt.Printf("[MockUserDAO] Create user = %+v", user)
	return args.Error(0)
}

func (dao *MockUserDAO) Save(user *config.User) error {
	args := dao.Called(user)
	fmt.Printf("[MockUserDAO] Save user=%+v", user)
	return args.Error(0)
}

func (dao *MockUserDAO) Find() ([]*config.User, error) {
	args := dao.Called()
	fmt.Printf("[MockUserDAO] Find")
	return args.Get(0).([]*config.User), args.Error(0)
}
