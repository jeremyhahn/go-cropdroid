// +build cluster

package service

import (
	"context"
	"fmt"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/provisioner"
	"golang.org/x/crypto/bcrypt"
	oauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type GoogleAuthService struct {
	app     *app.App
	userDAO dao.UserDAO
	orgDAO  dao.OrganizationDAO
	farmDAO dao.FarmDAO
	mapper  mapper.UserMapper
	AuthService
}

func NewGoogleAuthService(app *app.App, userDAO dao.UserDAO, orgDAO dao.OrganizationDAO,
	farmDAO dao.FarmDAO, userMapper mapper.UserMapper) AuthService {
	return &GoogleAuthService{
		app:     app,
		userDAO: userDAO,
		orgDAO:  orgDAO,
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

func (service *GoogleAuthService) Login(userCredentials *UserCredentials,
	farmProvisioner provisioner.FarmProvisioner) (common.UserAccount, []config.OrganizationConfig, error) {

	service.app.Logger.Debugf("[GoogleAuthService.Login] Authenticating user: %+v", userCredentials)

	idToken := userCredentials.Email
	context := context.Background()
	oauth2Service, err := oauth2.NewService(context, option.WithoutAuthentication())
	if err != nil {
		service.app.Logger.Error(err)
		return nil, nil, err
	}
	tokenInfo, err := oauth2Service.Tokeninfo().IdToken(idToken).Do()
	if err != nil {
		service.app.Logger.Errorf("[GoogleAuthService.Login] Error: %s", err)
		return nil, nil, ErrInvalidCredentials
	}

	service.app.Logger.Debugf("tokenInfo: %+v", tokenInfo)

	// Create a new trial account if this is a new user
	userEntity, err := service.userDAO.GetByEmail(tokenInfo.Email)
	if err != nil && err.Error() == ErrRecordNotFound.Error() {

		service.app.Logger.Debugf("[UserService.Login] Provisioning new Google account: %s", userCredentials.Email)

		userAccount, err := service.Register(&UserCredentials{
			Email:    tokenInfo.Email,
			Password: idToken})
		if err != nil {
			return nil, nil, err
		}

		farmConfig, err := farmProvisioner.Provision(userAccount)
		// TODO: Wait for account creation confirmation
		if err != nil {
			return nil, nil, err
		}

		/*
				userAccount := &model.User{
					Email:    userCredentials.Email,
					Password: userCredentials.Password,
					Roles:    []common.Role{model.NewRole(common.DEFAULT_ROLE)}}
			farmConfig, err := farmProvisioner.BuildConfig(userAccount)
			if err != nil {
				return nil, nil, err
			}
			farmStateChangeChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)
			farmFactory.BuildService(farmConfig, farmStateChangeChan)
		*/

		newOrg := &config.Organization{ID: 0, Farms: []config.Farm{*farmConfig.(*config.Farm)}}
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

	organizations, err := service.orgDAO.GetByUserID(userEntity.GetID())
	if err != nil {
		service.app.Logger.Errorf("[LocalAuthService.Login] Database error: %s", err)
		return nil, nil, ErrInternalDatabase
	}
	if len(organizations) == 0 {
		farms, err := service.farmDAO.GetByOrgAndUserID(0, userEntity.GetID())
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

func (service *GoogleAuthService) Register(userCredentials *UserCredentials) (common.UserAccount, error) {
	if !service.app.EnableRegistrations {
		return nil, ErrRegistrationDisabled
	}
	email := userCredentials.Email
	token := userCredentials.Password
	_, err := service.userDAO.GetByEmail(email)
	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		service.app.Logger.Errorf("[GoogleAuthService.Register] %s", err.Error())
		return nil, fmt.Errorf("Unexpected error: %s", err.Error())
	}
	encrypted, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	userConfig := &config.User{
		Email:    email,
		Password: string(encrypted)}

	err = service.userDAO.Create(userConfig) // creates userConfig.id
	if err != nil {
		return nil, err
	}
	userAccount := &model.User{
		ID:       userConfig.GetID(),
		Email:    email,
		Password: token,
		Roles:    []common.Role{model.NewRole(common.DEFAULT_ROLE)}}
	return userAccount, err
}
