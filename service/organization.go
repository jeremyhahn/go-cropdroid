package service

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
)

type DefaultOrganizationService struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	orgDAO      dao.OrganizationDAO
	OrganizationService
}

func NewOrganizationService(logger *logging.Logger, idGenerator util.IdGenerator,
	orgDAO dao.OrganizationDAO) OrganizationService {
	return &DefaultOrganizationService{
		logger:      logger,
		idGenerator: idGenerator,
		orgDAO:      orgDAO}
}

// Creates a new organization
func (service *DefaultOrganizationService) Create(organization *config.Organization) error {
	organization.SetID(service.idGenerator.NewStringID(organization.GetName()))
	return service.orgDAO.Save(organization)
}

// Returns a list of User entities that belong to the organization
func (service *DefaultOrganizationService) GetAll(session Session) ([]*config.Organization, error) {
	if !session.GetUser().HasRole(common.ROLE_ADMIN) {
		return nil, ErrPermissionDenied
	}
	return service.orgDAO.GetAll(common.CONSISTENCY_LOCAL)
}

// Returns a list of User entities that belong to the organization
func (service *DefaultOrganizationService) GetUsers(session Session) ([]*config.User, error) {
	if !session.HasRole(common.ROLE_ADMIN) {
		return nil, ErrPermissionDenied
	}
	return service.orgDAO.GetUsers(session.GetRequestedOrganizationID())
}

// Deletes an existing organization and all associated entites from the database
func (service *DefaultOrganizationService) Delete(session Session) error {
	if !session.GetUser().HasRole(common.ROLE_ADMIN) {
		return ErrPermissionDenied
	}
	return service.orgDAO.Delete(
		&config.Organization{
			ID: session.GetRequestedOrganizationID()})
}
