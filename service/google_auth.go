package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"golang.org/x/crypto/bcrypt"
	oauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type GoogleAuthService struct {
	app     *app.App
	orgDAO  dao.OrganizationDAO
	userDAO dao.UserDAO
	roleDAO dao.RoleDAO
	farmDAO dao.FarmDAO
	mapper  mapper.UserMapper
	AuthService
}

func NewGoogleAuthService(app *app.App, orgDAO dao.OrganizationDAO,
	userDAO dao.UserDAO, roleDAO dao.RoleDAO, farmDAO dao.FarmDAO,
	userMapper mapper.UserMapper) AuthService {

	return &GoogleAuthService{
		app:     app,
		orgDAO:  orgDAO,
		userDAO: userDAO,
		roleDAO: roleDAO,
		farmDAO: farmDAO,
		mapper:  userMapper}
}

func (service *GoogleAuthService) Get(email string) (common.UserAccount, error) {
	userEntity, err := service.userDAO.GetByEmail(email)
	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	return service.mapper.MapUserEntityToModel(userEntity), nil
}

func (service *GoogleAuthService) Login(userCredentials *UserCredentials) (common.UserAccount, []config.OrganizationConfig, error) {

	service.app.Logger.Debugf("Authenticating user: %+v", userCredentials)

	idToken := userCredentials.Email
	context := context.Background()
	oauth2Service, err := oauth2.NewService(context, option.WithoutAuthentication())
	if err != nil {
		service.app.Logger.Error(err)
		return nil, nil, err
	}
	tokenInfo, err := oauth2Service.Tokeninfo().IdToken(idToken).Do()
	if err != nil {
		service.app.Logger.Errorf("Error: %s", err)
		return nil, nil, ErrInvalidCredentials
	}

	service.app.Logger.Debugf("tokenInfo: %+v", tokenInfo)

	userEntity, err := service.userDAO.GetByEmail(tokenInfo.Email)

	// Create a new trial account if this is a new user
	if err != nil && err.Error() == ErrRecordNotFound.Error() {

		service.app.Logger.Debugf("Provisioning new Google account: %s", userCredentials.Email)

		userAccount, err := service.Register(&UserCredentials{
			Email:    tokenInfo.Email,
			Password: idToken}, "")
		if err != nil {
			return nil, nil, err
		}

		roleConfig, err := service.roleDAO.GetByName(common.ROLE_ADMIN)
		if err != nil {
			return nil, nil, err
		}
		userAccount.SetRoles([]common.Role{
			&model.Role{
				ID:   roleConfig.GetID(),
				Name: roleConfig.GetName()}})

		// provisionerParams := &provisioner.ProvisionerParams{}

		// if service.app.Mode == common.MODE_STANDALONE {
		// 	provisionerParams.StateStore = state.MEMORY_STORE
		// 	provisionerParams.ConfigStore = config.GORM_STORE
		// 	provisionerParams.DataStore = datastore.GORM_STORE
		// } else {
		// 	provisionerParams.StateStore = state.RAFT_STORE
		// 	provisionerParams.ConfigStore = config.RAFT_MEMORY_STORE
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
		newOrg := &config.Organization{ID: 0}
		return userAccount, []config.OrganizationConfig{newOrg}, nil
	}

	/*
		organizations := make([]config.OrganizationConfig, 0)
		for _, org := range service.app.Config.Organizations {
			for _, user := range org.Users {
				if user.Email == tokenInfo.Email {
					organizations = append(organizations, &org)
					break
				}
			}
		}
		if len(organizations) == 0 {
			org := config.Organization{ID: 0, Farms: make([]config.Farm, 0)}
			for _, farm := range service.app.Config.Farms {
				for _, user := range farm.Users {
					if user.Email == tokenInfo.Email {
						org.Farms = append(org.Farms, farm)
						break
					}
				}
			}
			organizations = append(organizations, &org)
		}*/

	organizations, err := service.orgDAO.GetByUserID(userEntity.GetID(), true)
	if err != nil {
		service.app.Logger.Errorf("Database error: %s", err)
		return nil, nil, ErrInternalDatabase
	}
	if len(organizations) == 0 {
		farms, err := service.farmDAO.GetByUserID(userEntity.GetID())
		if err != nil {
			return nil, nil, err
		}
		org := &config.Organization{
			ID:    0,
			Farms: farms}
		organizations = append(organizations, org)
	}

	userEntity.SetPassword(idToken)
	return service.mapper.MapUserEntityToModel(userEntity), organizations, nil
}

func (service *GoogleAuthService) Register(userCredentials *UserCredentials,
	baseURI string) (common.UserAccount, error) {

	if !service.app.Config.EnableRegistrations {
		return nil, ErrRegistrationDisabled
	}
	email := userCredentials.Email
	token := userCredentials.Password
	_, err := service.userDAO.GetByEmail(email)
	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		service.app.Logger.Errorf("%s", err.Error())
		return nil, fmt.Errorf("Unexpected error: %s", err.Error())
	}
	encrypted, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
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
	userConfig := &config.User{
		Email:    email,
		Password: string(encrypted)}
	//	Roles:    []config.Role{*roleConfig.(*config.Role)}}

	err = service.userDAO.Create(userConfig) // creates userConfig.id
	if err != nil {
		return nil, err
	}

	userAccount := &model.User{
		ID:       userConfig.GetID(),
		Email:    email,
		Password: token}

	return userAccount, err
}

func (service *GoogleAuthService) Activate(registrationID uint64) (common.UserAccount, error) {
	err := errors.New("GoogleAuthService.Activate not implemented")
	service.app.Logger.Error(err)
	return nil, err
}
