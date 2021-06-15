// +build !cluster

package service

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
)

var (
	ErrUnsupportedAuthType = errors.New("Unsupported auth type")
)

type DefaultUserService struct {
	app             *app.App
	userDAO         dao.UserDAO
	orgDAO          dao.OrganizationDAO
	roleDAO         dao.RoleDAO
	farmDAO         dao.FarmDAO
	userMapper      mapper.UserMapper
	authServices    map[int]AuthService
	serviceRegistry ServiceRegistry
	UserService
	AuthService
}

func NewUserService(app *app.App, userDAO dao.UserDAO, orgDAO dao.OrganizationDAO,
	roleDAO dao.RoleDAO, farmDAO dao.FarmDAO, userMapper mapper.UserMapper,
	authServices map[int]AuthService, serviceRegistry ServiceRegistry) UserService {

	return &DefaultUserService{
		app:             app,
		userDAO:         userDAO,
		orgDAO:          orgDAO,
		roleDAO:         roleDAO,
		farmDAO:         farmDAO,
		userMapper:      userMapper,
		authServices:    authServices,
		serviceRegistry: serviceRegistry}
}

func (service *DefaultUserService) Get(email string) (common.UserAccount, error) {
	userEntity, err := service.userDAO.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	return service.userMapper.MapUserEntityToModel(userEntity), nil
}

// Register signs up a new account
func (service *DefaultUserService) Register(userCredentials *UserCredentials) (common.UserAccount, error) {
	if authService, ok := service.authServices[userCredentials.AuthType]; ok {
		return authService.Register(userCredentials)
	}
	return nil, ErrUnsupportedAuthType
}

// Login authenticates a user account against the AuthService
func (service *DefaultUserService) Login(userCredentials *UserCredentials,
	farmProvisioner provisioner.FarmProvisioner) (common.UserAccount, []config.OrganizationConfig, error) {

	if authService, ok := service.authServices[userCredentials.AuthType]; ok {
		user, orgs, err := authService.Login(userCredentials, farmProvisioner)
		if err != nil {
			return nil, nil, err
		}
		return user, orgs, nil
	}
	return nil, nil, ErrUnsupportedAuthType
}

// CreateUser creates a new user account
func (service *DefaultUserService) CreateUser(user common.UserAccount) {
	service.userDAO.Create(&config.User{
		Email: user.GetEmail()})
}

/*
func (service *DefaultUserService) fillRoles(session Session) error {
	user := session.GetUser()
	roleEntities, err := service.roleDAO.GetByUserAndOrgID(user.GetID(), session.GetFarmService().GetConfig().GetOrgID())
	if err != nil {
		return err
	}
	for _, role := range roleEntities {
		user.AddRole(&model.Role{
			ID:   role.GetID(),
			Name: role.GetName()})
	}
	return nil
}*/

// GetUserByEmail retrieves a user from the database by their email address
func (service *DefaultUserService) GetUserByEmail(email string) (common.UserAccount, error) {
	userEntity, err := service.userDAO.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	return service.userMapper.MapUserEntityToModel(userEntity), nil
}

// GetRoles returns a list of Role entities which the specified user belongs
func (service *DefaultUserService) GetRoles(userID, orgID int) ([]config.Role, error) {
	return service.roleDAO.GetByUserAndOrgID(userID, orgID)
}

/*
// GetCurrentUser gets the user account from the database
// using the UserAccount stored in AppContext
func (service *DefaultUserService) GetCurrentUser() (common.UserAccount, error) {

	email := service.session.GetUser().GetEmail()

	userEntity, err := service.userDAO.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	model := service.userMapper.MapUserEntityToModel(userEntity)
	service.fillRoles(service.session)

	//orgs, err := service.GetOrganizations(userID)
	//if err != nil {
	//	return nil, err
	//}
	//model := service.userMapper.MapUserEntityToModel(userEntity)
	//model.SetDeviceID(common.SERVER_CONTROLLER_ID)
	//model.SetOrganizationID(orgEntity.GetID())

	return model, nil
}
*/

/*
// GetUserByID retrieves a user from the database by their userID
func (service *DefaultUserService) GetUserByID(userID int) (common.UserAccount, error) {
	userEntity, err := service.userDAO.GetById(userID)
	if err != nil {
		return nil, err
	}
	roleEntity, err := service.roleDAO.GetByUserAndOrgID(userID, 0)
	if err != nil {
		return nil, err
	}
	model := service.userMapper.MapUserEntityToModel(userEntity)
	model.SetRole(roleEntity)
	return model, nil
}*/

// GetOrganizations returns a list of Organization entities which the specified user belongs
/*
func (service *DefaultUserService) GetOrganizations(userID int) ([]entity.Organization, error) {
	return service.orgDAO.GetByUserID(userID)
}*/
