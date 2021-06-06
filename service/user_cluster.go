// +build cluster

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

// RegisterCluster signs up a new cluster account
func (service *DefaultUserService) Register(userCredentials *UserCredentials) (common.UserAccount, error) {
	if authService, ok := service.authServices[userCredentials.AuthType]; ok {
		return authService.Register(userCredentials)
	}
	return nil, ErrUnsupportedAuthType
}

// Login authenticates a user account against the AuthService
func (service *DefaultUserService) Login(userCredentials *UserCredentials,
	farmProvisioner provisioner.FarmProvisioner) (common.UserAccount, []config.OrganizationConfig, error) {

	if userCredentials.AuthType == common.AUTH_TYPE_GOOGLE {
		//farmFactory := service.serviceRegistry.GetFarmFactory()
		//farmProvisioner = provisioner.NewGormFarmProvisioner(service.app.Logger, service.app.NewGormDB(),
		//service.app.Location, service.farmDAO, farmFactory.GetFarmProvisionerChan())

		farmProvisioner = provisioner.NewRaftFarmProvisioner(service.app.Logger, service.app.NewGormDB(),
			service.app.GossipCluster, service.app.Location, service.farmDAO)
	}

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
