package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/util"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type GoogleAuthService struct {
	app           *app.App
	idGenerator   util.IdGenerator
	permissionDAO dao.PermissionDAO
	userDAO       dao.UserDAO
	roleDAO       dao.RoleDAO
	farmDAO       dao.FarmDAO
	mapper        mapper.UserMapper
	AuthServicer
}

func NewGoogleAuthService(
	app *app.App,
	permissionDAO dao.PermissionDAO,
	userDAO dao.UserDAO,
	roleDAO dao.RoleDAO,
	farmDAO dao.FarmDAO,
	userMapper mapper.UserMapper) AuthServicer {

	return &GoogleAuthService{
		app:           app,
		idGenerator:   util.NewIdGenerator(app.DataStoreEngine),
		permissionDAO: permissionDAO,
		userDAO:       userDAO,
		roleDAO:       roleDAO,
		farmDAO:       farmDAO,
		mapper:        userMapper}
}

// func (service *GoogleAuthService) Get(userID uint64) (model.User, error) {
// 	userEntity, err := service.userDAO.Get(userID)
// 	if err != nil && err.Error() != ErrRecordNotFound.Error() {
// 		return nil, ErrInvalidCredentials
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return service.mapper.MapUserEntityToModel(userEntity), nil
// }

func (service *GoogleAuthService) ResetPassword(userCredentials *UserCredentials) error {
	return ErrResetPasswordUnsupported
}

func (service *GoogleAuthService) Login(userCredentials *UserCredentials) (model.User,
	[]config.Organization, []config.Farm, error) {

	service.app.Logger.Debugf("Authenticating user: %+v", userCredentials)

	idToken := userCredentials.Email
	context := context.Background()
	oauth2Service, err := oauth2.NewService(context, option.WithoutAuthentication())
	if err != nil {
		service.app.Logger.Error(err)
		return nil, nil, nil, err
	}
	tokenInfo, err := oauth2Service.Tokeninfo().IdToken(idToken).Do()
	if err != nil {
		service.app.Logger.Errorf("Error: %s", err)
		return nil, nil, nil, ErrInvalidCredentials
	}
	service.app.Logger.Debugf("tokenInfo: %+v", tokenInfo)

	userID := service.idGenerator.NewStringID(tokenInfo.Email)
	userEntity, err := service.userDAO.Get(userID, common.CONSISTENCY_LOCAL)

	// Create a new trial account if this is a new user
	if err != nil && err.Error() == ErrRecordNotFound.Error() {

		service.app.Logger.Debugf("Provisioning new Google account: %s", userCredentials.Email)

		userAccount, err := service.Register(&UserCredentials{
			Email:    tokenInfo.Email,
			Password: idToken}, "")
		if err != nil {
			return nil, nil, nil, err
		}

		roleConfig, err := service.roleDAO.GetByName(common.ROLE_ADMIN, common.CONSISTENCY_LOCAL)
		if err != nil {
			return nil, nil, nil, err
		}
		userAccount.SetRoles([]model.Role{
			&model.RoleStruct{
				ID:   roleConfig.ID,
				Name: roleConfig.GetName()}})

		// provisionerParams := &provisioner.ProvisionerParams{}

		// if service.app.Mode == common.MODE_STANDALONE {
		// 	provisionerParams.StateStore = state.MEMORY_STORE
		// 	provisionerParams.ServerStore = config.GORM_STORE
		// 	provisionerParams.DataStore = datastore.GORM_STORE
		// } else {
		// 	provisionerParams.StateStore = state.RAFT_STORE
		// 	provisionerParams.ServerStore = config.RAFT_MEMORY_STORE
		// 	provisionerParams.DataStore = datastore.GORM_STORE
		// }

		// farmConfig, err := service.farmProvisioner.Provision(userAccount, provisionerParams)
		// // TODO: Wait for account creation confirmation
		// if err != nil {
		// 	return nil, nil, err
		// }

		// /*
		// 		userAccount := &model.User{
		// 			Email:    userCredentials.Email,
		// 			Password: userCredentials.Password,
		// 			Roles:    []common.Role{model.NewRole(common.DEFAULT_ROLE)}}
		// 	farmConfig, err := farmProvisioner.BuildConfig(userAccount)
		// 	if err != nil {
		// 		return nil, nil, err
		// 	}
		// 	farmStateChangeChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)
		// 	farmFactory.BuildService(farmConfig, farmStateChangeChan)
		// */

		//newOrg := &config.Organization{ID: 0, Farms: []config.Farm{*farmConfig.(*config.Farm)}}
		newOrg := &config.OrganizationStruct{ID: 0}
		return userAccount, []config.Organization{newOrg}, nil, nil
	}

	/*
		organizations := make([]config.OrganizationConfig, 0)
		for _, org := range service.app.Organizations {
			for _, user := range org.Users {
				if user.Email == tokenInfo.Email {
					organizations = append(organizations, &org)
					break
				}
			}
		}
		if len(organizations) == 0 {
			org := config.Organization{ID: 0, Farms: make([]config.Farm, 0)}
			for _, farm := range service.app.Farms {
				for _, user := range farm.Users {
					if user.Email == tokenInfo.Email {
						org.Farms = append(org.Farms, farm)
						break
					}
				}
			}
			organizations = append(organizations, &org)
		}*/

	organizations, err := service.permissionDAO.GetOrganizations(userEntity.ID, common.CONSISTENCY_LOCAL)
	if err != nil {
		service.app.Logger.Errorf("Database error: %s", err)
		return nil, nil, nil, ErrInternalDatabase
	}

	farms, err := service.farmDAO.GetByUserID(userEntity.ID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, nil, nil, err
	}

	userEntity.SetPassword(idToken)

	// Convert from Structs to interface types
	orgs := make([]config.Organization, len(organizations))
	for i, org := range organizations {
		orgs[i] = org
	}
	_farms := make([]config.Farm, len(farms))
	for i, farm := range farms {
		_farms[i] = farm
	}
	return service.mapper.MapUserConfigToModel(userEntity), orgs, _farms, nil
}

func (service *GoogleAuthService) Register(userCredentials *UserCredentials,
	baseURI string) (model.User, error) {

	if !service.app.EnableRegistrations {
		return nil, ErrRegistrationDisabled
	}
	email := userCredentials.Email
	token := userCredentials.Password
	userID := service.idGenerator.NewStringID(email)
	_, err := service.userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		service.app.Logger.Errorf("%s", err.Error())
		return nil, fmt.Errorf("Unexpected error: %s", err.Error())
	}

	passwordHasher := util.CreatePasswordHasher(service.app.PasswordHasherParams)
	encrypted, err := passwordHasher.Encrypt(token)
	if err != nil {
		return nil, err
	}

	// var roleConfig config.RoleConfig
	// if userCredentials.OrganizationID > 0 {
	// 	roleConfig, err = service.roleDAO.GetByName(common.ROLE_ANALYST)
	// } else {
	// 	roleConfig, err = service.roleDAO.GetByName(common.ROLE_ADMIN)
	// }
	//roleConfig, err := service.roleDAO.GetByName(common.ROLE_ADMIN)

	defaultRole, err := service.roleDAO.GetByName(service.app.DefaultRole, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}

	userConfig := &config.UserStruct{
		ID:       service.idGenerator.NewStringID(email),
		Email:    email,
		Password: string(encrypted),
		Roles:    []*config.RoleStruct{defaultRole}}

	err = service.userDAO.Save(userConfig) // creates userConfig.id
	if err != nil {
		return nil, err
	}

	userAccount := &model.UserStruct{
		ID:       userConfig.ID,
		Email:    email,
		Password: token}

	return userAccount, err
}

func (service *GoogleAuthService) Activate(registrationID uint64) (model.User, error) {
	// Google already verified the email address, no need
	// to perform registration/activation process
	err := errors.New("GoogleAuthService.Activate not implemented")
	service.app.Logger.Error(err)
	return nil, err
}
