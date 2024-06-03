package service

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
)

type OrganizationService interface {
	Create(organization config.Organization) error
	Page(session Session, pageQuery query.PageQuery) (dao.PageResult[*config.OrganizationStruct], error)
	GetUsers(session Session) ([]model.User, error)
	Delete(session Session) error
}

type Organization struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	orgDAO      dao.OrganizationDAO
	userMapper  mapper.UserMapper
	OrganizationService
}

func NewOrganizationService(
	logger *logging.Logger,
	idGenerator util.IdGenerator,
	orgDAO dao.OrganizationDAO,
	userMapper mapper.UserMapper) OrganizationService {

	return &Organization{
		logger:      logger,
		idGenerator: idGenerator,
		orgDAO:      orgDAO,
		userMapper:  userMapper}
}

// Creates a new organization
func (service *Organization) Create(organization config.Organization) error {
	organization.SetID(service.idGenerator.NewStringID(organization.GetName()))
	return service.orgDAO.Save(organization.(*config.OrganizationStruct))
}

// Returns a list of User entities that belong to the organization
func (service *Organization) Page(session Session,
	pageQuery query.PageQuery) (dao.PageResult[*config.OrganizationStruct], error) {

	if !session.GetUser().HasRole(common.ROLE_ADMIN) {
		return dao.PageResult[*config.OrganizationStruct]{}, ErrPermissionDenied
	}
	return service.orgDAO.GetPage(pageQuery, common.CONSISTENCY_LOCAL)
}

// Returns a list of User entities that belong to the organization
func (service *Organization) GetUsers(session Session) ([]model.User, error) {
	if !session.HasRole(common.ROLE_ADMIN) {
		return nil, ErrPermissionDenied
	}
	userStructs, err := service.orgDAO.GetUsers(session.GetRequestedOrganizationID())
	if err != nil {
		service.logger.Error(err)
		return nil, err
	}
	userModels := make([]model.User, len(userStructs))
	for i, user := range userStructs {
		userModels[i] = service.userMapper.MapUserConfigToModel(user)
	}
	return userModels, nil
}

// Deletes an existing organization and all associated entites from the database
func (service *Organization) Delete(session Session) error {
	if !session.GetUser().HasRole(common.ROLE_ADMIN) {
		return ErrPermissionDenied
	}
	return service.orgDAO.Delete(
		&config.OrganizationStruct{
			ID: session.GetRequestedOrganizationID()})
}
