package service

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	logging "github.com/op/go-logging"
)

var (
	ErrRoleNotFound = errors.New("role not found")
)

type RoleServicer interface {
	GetPage(pageQuery query.PageQuery) (dao.PageResult[*config.RoleStruct], error)
	GetByName(name string, CONSISTENCY_LEVEL int) (config.Role, error)
}

type RoleService struct {
	logger  *logging.Logger
	roleDAO dao.RoleDAO
	RoleServicer
}

func NewRoleService(
	logger *logging.Logger,
	roleDAO dao.RoleDAO) RoleServicer {

	return &RoleService{
		logger:  logger,
		roleDAO: roleDAO}
}

// Returns a list of all Role entities in the database
func (service *RoleService) GetPage(pageQuery query.PageQuery) (dao.PageResult[*config.RoleStruct], error) {
	return service.roleDAO.GetPage(pageQuery, common.CONSISTENCY_LOCAL)
}

// Returns the role with the given name
func (service *RoleService) GetByName(name string, CONSISTENCY_LEVEL int) (config.Role, error) {
	return service.roleDAO.GetByName(name, CONSISTENCY_LEVEL)
}
