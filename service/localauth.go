package service

import (
	"errors"
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials   = errors.New("Invalid username/password")
	ErrRecordNotFound       = errors.New("record not found")
	ErrInternalDatabase     = errors.New("Internal database error")
	ErrRegistrationDisabled = errors.New("User registrations disabled")
)

type LocalAuthService struct {
	app     *app.App
	orgDAO  dao.OrganizationDAO
	userDAO dao.UserDAO
	roleDAO dao.RoleDAO
	farmDAO dao.FarmDAO
	mapper  mapper.UserMapper
	AuthService
}

func NewLocalAuthService(app *app.App, orgDAO dao.OrganizationDAO, userDAO dao.UserDAO,
	roleDAO dao.RoleDAO, farmDAO dao.FarmDAO, userMapper mapper.UserMapper) AuthService {

	return &LocalAuthService{
		app:     app,
		orgDAO:  orgDAO,
		userDAO: userDAO,
		roleDAO: roleDAO,
		farmDAO: farmDAO,
		mapper:  userMapper}
}

func (service *LocalAuthService) Get(email string) (common.UserAccount, error) {
	userEntity, err := service.userDAO.GetByEmail(email)
	if err != nil && err != ErrRecordNotFound {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	return service.mapper.MapUserEntityToModel(userEntity), nil
}

func (service *LocalAuthService) Login(userCredentials *UserCredentials) (common.UserAccount, []config.OrganizationConfig, error) {

	service.app.Logger.Debugf("Authenticating user: %s", userCredentials.Email)

	userEntity, err := service.userDAO.GetByEmail(userCredentials.Email)
	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, err
	}

	roleConfig, err := service.roleDAO.GetByName(common.ROLE_ADMIN)
	userEntity.SetRoles([]config.Role{*roleConfig.(*config.Role)})

	err = bcrypt.CompareHashAndPassword([]byte(userEntity.GetPassword()), []byte(userCredentials.Password))
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}
	userEntity.RedactPassword()
	/*
		organizations := make([]config.OrganizationConfig, 0)
		for _, org := range service.app.Config.Organizations {
			for _, user := range org.Users {
				if user.Email == email {
					organizations = append(organizations, &org)
					break
				}
			}
		}
		if len(organizations) == 0 {
			org := config.Organization{
				ID:    0,
				Farms: make([]config.Farm, 0)}
			for _, farm := range service.app.Config.Farms {
				for _, user := range farm.Users {
					if user.Email == email {
						org.Farms = append(org.Farms, farm)
						break
					}
				}
			}
			organizations = append(organizations, &org)
		}*/

	organizations, err := service.orgDAO.GetByUserID(userEntity.GetID())
	if err != nil {
		service.app.Logger.Errorf("Error looking up organization user: %s", err)
		return nil, nil, err
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
	return service.mapper.MapUserEntityToModel(userEntity), organizations, nil
}

func (service *LocalAuthService) Register(userCredentials *UserCredentials) (common.UserAccount, error) {
	if !service.app.EnableRegistrations {
		return nil, ErrRegistrationDisabled
	}
	_, err := service.userDAO.GetByEmail(userCredentials.Email)
	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		service.app.Logger.Errorf("%s", err.Error())
		return nil, fmt.Errorf("Unexpected error: %s", err.Error())
	}
	encrypted, err := bcrypt.GenerateFromPassword([]byte(userCredentials.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	userConfig := &config.User{
		Email:    userCredentials.Email,
		Password: string(encrypted)}
	err = service.userDAO.Create(userConfig) // creates userConfig.id
	if err != nil {
		return nil, err
	}
	userAccount := &model.User{
		ID:       userConfig.GetID(),
		Email:    userCredentials.Email,
		Password: userCredentials.Password,
		Roles:    []common.Role{model.NewRole(common.DEFAULT_ROLE)}}
	return userAccount, err
}

/*
func (service *LocalAuthService) GetOrganizations(userID uint64) ([]entity.OrganizationEntity, error) {
	orgs, err := service.orgDAO.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}
*/
