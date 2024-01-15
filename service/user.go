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
	permissionDAO   dao.PermissionDAO
	farmDAO         dao.FarmDAO
	userMapper      mapper.UserMapper
	authServices    map[int]AuthService
	serviceRegistry ServiceRegistry
	UserService
	AuthService
}

func NewUserService(app *app.App, userDAO dao.UserDAO, orgDAO dao.OrganizationDAO,
	roleDAO dao.RoleDAO, permissionDAO dao.PermissionDAO, farmDAO dao.FarmDAO,
	userMapper mapper.UserMapper, authServices map[int]AuthService,
	serviceRegistry ServiceRegistry) UserService {

	return &DefaultUserService{
		app:             app,
		userDAO:         userDAO,
		orgDAO:          orgDAO,
		roleDAO:         roleDAO,
		permissionDAO:   permissionDAO,
		farmDAO:         farmDAO,
		userMapper:      userMapper,
		authServices:    authServices,
		serviceRegistry: serviceRegistry}
}

// Looks up the user account by user ID
func (service *DefaultUserService) Get(userID uint64) (common.UserAccount, error) {
	userEntity, err := service.userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	return service.userMapper.MapUserConfigToModel(userEntity), nil
}

// Sets a new user password
func (service *DefaultUserService) ResetPassword(userCredentials *UserCredentials) error {
	if authService, ok := service.authServices[userCredentials.AuthType]; ok {
		return authService.ResetPassword(userCredentials)
	}
	return ErrUnsupportedAuthType
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
func (service *DefaultUserService) Login(userCredentials *UserCredentials) (common.UserAccount,
	[]*config.Organization, []*config.Farm, error) {

	if authService, ok := service.authServices[userCredentials.AuthType]; ok {
		user, orgs, farms, err := authService.Login(userCredentials)
		if err != nil {
			return nil, nil, nil, err
		}
		return user, orgs, farms, nil
	}
	return nil, nil, nil, ErrUnsupportedAuthType
}

// Reloads the users organizations, farms and permissions
func (service *DefaultUserService) Refresh(userID uint64) (common.UserAccount,
	[]*config.Organization, []*config.Farm, error) {

	service.app.Logger.Debugf("Refreshing user: %d", userID)

	var user *config.User

	organizations, err := service.permissionDAO.GetOrganizations(userID, common.CONSISTENCY_LOCAL)
	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		return nil, nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, nil, err
	}

ORG_LOOP:
	for _, org := range organizations {
		for _, u := range org.GetUsers() {
			if user.GetID() == userID {
				user = u
				user.RedactPassword()
				break ORG_LOOP
			}
			// user.SetRoles([]config.Role{{ID: 1, Name: common.DEFAULT_ROLE}})
		}
	}

	farms, err := service.farmDAO.GetByUserID(userID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, nil, nil, err
	}

	if user == nil {
		user, err = service.userDAO.Get(userID, common.CONSISTENCY_LOCAL)
		if err != nil {
			return nil, nil, nil, err
		}
		userAccount := service.userMapper.MapUserConfigToModel(user)
		return userAccount, organizations, farms, nil
	}

	if user.GetID() == 0 && len(organizations) == 0 {
	FARM_LOOP:
		for _, farm := range farms {
			for _, u := range farm.GetUsers() {
				if u.GetID() == userID {
					user = u
					user.RedactPassword()
					break FARM_LOOP
				}
			}
		}
	}

	if user.GetID() == 0 {
		return nil, nil, nil, ErrUserNotFound
	}

	return service.userMapper.MapUserConfigToModel(user), organizations, farms, nil
}

// CreateUser creates a new user account
func (service *DefaultUserService) CreateUser(user common.UserAccount) error {
	return service.userDAO.Save(&config.User{
		Email: user.GetEmail()})
}

// UpdateUser an existing user account
func (service *DefaultUserService) UpdateUser(user common.UserAccount) error {
	userConfig := service.userMapper.MapUserModelToConfig(user)
	return service.userDAO.Save(userConfig)
}

// Deletes an existing user account
func (service *DefaultUserService) Delete(session Session, userID uint64) error {
	if err := service.DeletePermission(session, userID); err != nil {
		return err
	}
	return service.userDAO.Delete(&config.User{ID: userID})
}

// Sets the users "permission", ie., the role that grants access
// to an organization and/or farm.
func (service *DefaultUserService) SetPermission(session Session, permission *config.Permission) error {
	if !session.GetUser().HasRole(common.ROLE_ADMIN) {
		return ErrPermissionDenied
	}
	if permission.GetUserID() == common.DEFAULT_USER_ID_64 || permission.GetUserID() == common.DEFAULT_USER_ID_32 {
		return ErrChangeAdminRole
	}
	return service.permissionDAO.Update(permission)
}

// Delete a user permission from the requested farm
func (service *DefaultUserService) DeletePermission(session Session, userID uint64) error {
	if !session.GetUser().HasRole(common.ROLE_ADMIN) {
		return ErrPermissionDenied
	}
	if userID == common.DEFAULT_USER_ID_64 || userID == common.DEFAULT_USER_ID_32 {
		return ErrDeleteAdminAccount
	}
	return service.permissionDAO.Delete(&config.Permission{
		OrganizationID: session.GetRequestedOrganizationID(),
		FarmID:         session.GetRequestedFarmID(),
		UserID:         userID})
}
