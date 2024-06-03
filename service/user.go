package service

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
)

var (
	ErrUnsupportedAuthType = errors.New("unsupported auth type")
	ErrUserNotFound        = errors.New("user not found")
)

type UserServicer interface {
	CreateUser(user model.User) error
	UpdateUser(user model.User) error // replaced with SetPermnission?
	Delete(session Session, userID uint64) error
	DeletePermission(session Session, userID uint64) error
	//Get(email string) (model.User, error)
	Get(userID uint64) (model.User, error)
	SetPermission(session Session, permission config.Permission) error
	// probably needs to be moved to auth service; not implemented in google_auth yet
	Refresh(userID uint64) (model.User, []config.Organization, []config.Farm, error)
	AuthServicer
}

type User struct {
	app             *app.App
	userDAO         dao.UserDAO
	orgDAO          dao.OrganizationDAO
	roleDAO         dao.RoleDAO
	permissionDAO   dao.PermissionDAO
	farmDAO         dao.FarmDAO
	userMapper      mapper.UserMapper
	authServices    map[int]AuthServicer
	serviceRegistry ServiceRegistry
	UserServicer
	AuthServicer
}

func NewUserService(
	app *app.App,
	userDAO dao.UserDAO,
	orgDAO dao.OrganizationDAO,
	roleDAO dao.RoleDAO,
	permissionDAO dao.PermissionDAO,
	farmDAO dao.FarmDAO,
	userMapper mapper.UserMapper,
	authServices map[int]AuthServicer,
	serviceRegistry ServiceRegistry) UserServicer {

	return &User{
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
func (service *User) Get(userID uint64) (model.User, error) {
	userEntity, err := service.userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	return service.userMapper.MapUserConfigToModel(userEntity), nil
}

// Sets a new user password
func (service *User) ResetPassword(userCredentials *UserCredentials) error {
	if authService, ok := service.authServices[userCredentials.AuthType]; ok {
		return authService.ResetPassword(userCredentials)
	}
	return ErrUnsupportedAuthType
}

// Register signs up a new account
func (service *User) Register(userCredentials *UserCredentials,
	baseURI string) (model.User, error) {

	if authService, ok := service.authServices[userCredentials.AuthType]; ok {
		return authService.Register(userCredentials, baseURI)
	}
	return nil, ErrUnsupportedAuthType
}

// Activates a pending registration
func (service *User) Activate(registrationID uint64) (model.User, error) {
	if authService, ok := service.authServices[common.AUTH_TYPE_LOCAL]; ok {
		return authService.Activate(registrationID)
	}
	return nil, ErrUnsupportedAuthType
}

// Login authenticates a user account against the AuthService
func (service *User) Login(userCredentials *UserCredentials) (model.User,
	[]config.Organization, []config.Farm, error) {

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
func (service *User) Refresh(userID uint64) (model.User,
	[]config.Organization, []config.Farm, error) {

	service.app.Logger.Debugf("Refreshing user: %d", userID)

	var user *config.UserStruct

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
			if user.Identifier() == userID {
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
		userModel := service.userMapper.MapUserConfigToModel(user)
		orgs := make([]config.Organization, len(organizations))
		for i, org := range organizations {
			orgs[i] = org
		}
		_farms := make([]config.Farm, len(farms))
		for i, farm := range farms {
			_farms[i] = farm
		}
		return userModel, orgs, _farms, nil
	}

	if user.ID == 0 && len(organizations) == 0 {
	FARM_LOOP:
		for _, farm := range farms {
			for _, u := range farm.GetUsers() {
				if u.ID == userID {
					user = u
					user.RedactPassword()
					break FARM_LOOP
				}
			}
		}
	}

	if user.ID == 0 {
		return nil, nil, nil, ErrUserNotFound
	}

	// Convert structs to interfaces
	orgs := make([]config.Organization, len(organizations))
	for i, org := range organizations {
		orgs[i] = org
	}
	_farms := make([]config.Farm, len(farms))
	for i, farm := range farms {
		_farms[i] = farm
	}
	return service.userMapper.MapUserConfigToModel(user), orgs, _farms, nil
}

// CreateUser creates a new user account
func (service *User) CreateUser(user model.User) error {
	return service.userDAO.Save(&config.UserStruct{Email: user.GetEmail()})
}

// UpdateUser an existing user account
func (service *User) UpdateUser(user model.User) error {
	userConfig := service.userMapper.MapUserModelToConfig(user)
	return service.userDAO.Save(userConfig)
}

// Deletes an existing user account
func (service *User) Delete(session Session, userID uint64) error {
	if err := service.DeletePermission(session, userID); err != nil {
		return err
	}
	return service.userDAO.Delete(&config.UserStruct{ID: userID})
}

// Sets the users "permission", ie., the role that grants access
// to an organization and/or farm.
func (service *User) SetPermission(session Session, permission config.Permission) error {
	if !session.GetUser().HasRole(common.ROLE_ADMIN) {
		return ErrPermissionDenied
	}
	if permission.GetUserID() == common.DEFAULT_USER_ID_64 || permission.GetUserID() == common.DEFAULT_USER_ID_32 {
		return ErrChangeAdminRole
	}
	return service.permissionDAO.Update(permission.(*config.PermissionStruct))
}

// Delete a user permission from the requested farm
func (service *User) DeletePermission(session Session, userID uint64) error {
	if !session.GetUser().HasRole(common.ROLE_ADMIN) {
		return ErrPermissionDenied
	}
	if userID == common.DEFAULT_USER_ID_64 || userID == common.DEFAULT_USER_ID_32 {
		return ErrDeleteAdminAccount
	}
	return service.permissionDAO.Delete(&config.PermissionStruct{
		OrganizationID: session.GetRequestedOrganizationID(),
		FarmID:         session.GetRequestedFarmID(),
		UserID:         userID})
}
