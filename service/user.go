package service

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

var (
	ErrUnsupportedAuthType = errors.New("unsupported auth type")
	ErrUserNotFound        = errors.New("user not found")
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
func (service *DefaultUserService) Register(userCredentials *UserCredentials,
	baseURI string) (common.UserAccount, error) {

	if authService, ok := service.authServices[userCredentials.AuthType]; ok {
		return authService.Register(userCredentials, baseURI)
	}
	return nil, ErrUnsupportedAuthType
}

// Activates a pending registration
func (service *DefaultUserService) Activate(registrationID uint64) (common.UserAccount, error) {
	if authService, ok := service.authServices[common.AUTH_TYPE_LOCAL]; ok {
		return authService.Activate(registrationID)
	}
	return nil, ErrUnsupportedAuthType
}

// Login authenticates a user account against the AuthService
func (service *DefaultUserService) Login(userCredentials *UserCredentials) (common.UserAccount, []config.OrganizationConfig, error) {
	if authService, ok := service.authServices[userCredentials.AuthType]; ok {
		user, orgs, err := authService.Login(userCredentials)
		if err != nil {
			return nil, nil, err
		}
		return user, orgs, nil
	}
	return nil, nil, ErrUnsupportedAuthType
}

// Reloads the users organizations, farms and permissions
func (service *DefaultUserService) Refresh(userID uint64) (common.UserAccount, []config.OrganizationConfig, error) {

	service.app.Logger.Debugf("Refreshing user: %d", userID)

	var user *config.User

	organizations, err := service.orgDAO.GetByUserID(userID, true)
	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, err
	}

LOOP:
	for _, org := range organizations {
		for _, u := range org.GetUsers() {
			if user.GetID() == userID {
				user = &u
				user.RedactPassword()
				break LOOP
			}
			// user.SetRoles([]config.Role{{ID: 1, Name: common.DEFAULT_ROLE}})
		}
	}
	if user == nil {
		return nil, nil, ErrUserNotFound
	}

	if len(organizations) == 0 {
		farms, err := service.farmDAO.GetByUserID(userID)
		if err != nil {
			return nil, nil, err
		}
		org := &config.Organization{
			ID:    0,
			Farms: farms}
		organizations = append(organizations, org)
	}

	return service.userMapper.MapUserEntityToModel(user), organizations, nil
}

// CreateUser creates a new user account
func (service *DefaultUserService) CreateUser(user common.UserAccount) {
	service.userDAO.Create(&config.User{
		Email: user.GetEmail()})
}

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
