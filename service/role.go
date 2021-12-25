package service

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	logging "github.com/op/go-logging"
)

var (
	ErrRoleNotFound = errors.New("role not found")
)

type DefaultRoleService struct {
	logger  *logging.Logger
	roleDAO dao.RoleDAO
	RoleService
}

func NewRoleService(logger *logging.Logger,
	roleDAO dao.RoleDAO) RoleService {

	return &DefaultRoleService{
		logger:  logger,
		roleDAO: roleDAO}
}

// Returns a list of all Role entities in the database
func (service *DefaultRoleService) GetAll() ([]config.RoleConfig, error) {
	return service.roleDAO.GetAll()
}
