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
func (service *DefaultRoleService) GetPage(pageQuery query.PageQuery) (dao.PageResult[*config.Role], error) {
	return service.roleDAO.GetPage(pageQuery, common.CONSISTENCY_LOCAL)
}
