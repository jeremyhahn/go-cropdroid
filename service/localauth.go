package service

import (
	"errors"
	"fmt"
	"html/template"
	"regexp"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/util"
)

var (
	ErrInvalidCredentials         = errors.New("Invalid username/password")
	ErrRecordNotFound             = errors.New("record not found")
	ErrInternalDatabase           = errors.New("Internal database error")
	ErrRegistrationDisabled       = errors.New("User registrations disabled")
	ErrOrgAlreadyExists           = errors.New("organization already exists")
	ErrUserAlreadyExists          = errors.New("user already exists")
	ErrOrgRegistrationUnsupported = errors.New("organization registration unsupported")
	ErrInvalidEmailAddress        = errors.New("invalid email address")
)

type LocalAuthService struct {
	app           *app.App
	idGenerator   util.IdGenerator
	permissionDAO dao.PermissionDAO
	regDAO        dao.RegistrationDAO
	orgDAO        dao.OrganizationDAO
	farmDAO       dao.FarmDAO
	userDAO       dao.UserDAO
	roleDAO       dao.RoleDAO
	mapper        mapper.UserMapper
	AuthService
}

// Performs authentication, registration and activation services using
// "local" datastore data access objects.
func NewLocalAuthService(app *app.App, permissionDAO dao.PermissionDAO,
	regDAO dao.RegistrationDAO, orgDAO dao.OrganizationDAO, farmDAO dao.FarmDAO,
	userDAO dao.UserDAO, roleDAO dao.RoleDAO, userMapper mapper.UserMapper) AuthService {

	return &LocalAuthService{
		app:           app,
		idGenerator:   util.NewIdGenerator(app.DataStoreEngine),
		permissionDAO: permissionDAO,
		regDAO:        regDAO,
		orgDAO:        orgDAO,
		farmDAO:       farmDAO,
		userDAO:       userDAO,
		roleDAO:       roleDAO,
		mapper:        userMapper}
}

// Looks up the specified user from the data store by email address
// func (service *LocalAuthService) Get(orgID uint64, email string) (common.UserAccount, error) {
// 	userID := service.idGenerator.NewID(email)
// 	userEntity, err := service.userDAO.Get(orgID, userID)
// 	if err != nil && err != ErrRecordNotFound {
// 		return nil, ErrInvalidCredentials
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return service.mapper.MapUserEntityToModel(userEntity), nil
// }

// ResetPassword looks the user up from the database, encrypts the UserCredentials.Password
// and updates the database with the encrypted value.
func (service *LocalAuthService) ResetPassword(userCredentials *UserCredentials) error {
	userID := service.idGenerator.NewStringID(userCredentials.Email)
	userEntity, err := service.userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	encrypted, err := service.encryptPassword(userCredentials.Password)
	if err != nil {
		return err
	}
	userEntity.SetPassword(string(encrypted))
	return service.userDAO.Save(userEntity)
}

// Login takes a set of credentials and returns a list of organizations with the farms
// the user has permission to access, minimally populated. No device or workflow data
// will be contained with the farm(s).
func (service *LocalAuthService) Login(userCredentials *UserCredentials) (common.UserAccount,
	[]*config.Organization, []*config.Farm, error) {

	service.app.Logger.Debugf("Authenticating user: %s", userCredentials.Email)

	userID := service.idGenerator.NewStringID(userCredentials.Email)
	userEntity, err := service.userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		return nil, nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, nil, err
	}

	match, err := service.comparePassword(userCredentials.Password, userEntity.GetPassword())
	if err != nil {
		return nil, nil, nil, err
	}
	if match == false {
		return nil, nil, nil, ErrInvalidCredentials
	}
	userEntity.RedactPassword()

	organizations, err := service.permissionDAO.GetOrganizations(userEntity.GetID(), common.CONSISTENCY_LOCAL)
	if err != nil {
		service.app.Logger.Errorf("Error looking up organization user: %s", err)
		return nil, nil, nil, err
	}

	farms, err := service.farmDAO.GetByUserID(userEntity.GetID(), common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, nil, nil, err
	}

	return service.mapper.MapUserConfigToModel(userEntity), organizations, farms, nil
}

// Registers a new user and sends an email with an activation button in HTML format.
func (service *LocalAuthService) Register(userCredentials *UserCredentials,
	baseURI string) (common.UserAccount, error) {

	if !service.app.EnableRegistrations {
		return nil, ErrRegistrationDisabled
	}

	pattern := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !pattern.MatchString(userCredentials.Email) {
		return nil, ErrInvalidEmailAddress
	}

	userID := service.idGenerator.NewStringID(userCredentials.Email)
	persistedUser, err := service.userDAO.Get(userID, common.CONSISTENCY_LOCAL)

	if err != nil && err.Error() != ErrRecordNotFound.Error() {
		service.app.Logger.Errorf("%s", err.Error())
		return nil, fmt.Errorf("Unexpected error: %s", err.Error())
	}
	if persistedUser != nil {
		return nil, ErrUserAlreadyExists
	}
	encrypted, err := service.encryptPassword(userCredentials.Password)
	if err != nil {
		return nil, err
	}

	registrationID := service.idGenerator.NewStringID(userCredentials.Email)
	registration := config.CreateRegistration(registrationID)
	registration.SetEmail(userCredentials.Email)
	registration.SetPassword(string(encrypted))

	if userCredentials.OrgName != "" {
		// if service.app.Mode == common.CONFIG_MODE_SERVER {
		// 	return nil, ErrOrgRegistrationUnsupported
		// }
		userCredentials.OrgID = service.idGenerator.NewStringID(userCredentials.OrgName)
		persistedOrg, err := service.orgDAO.Get(userCredentials.OrgID, common.CONSISTENCY_LOCAL)
		if err != nil && err.Error() != ErrRecordNotFound.Error() {
			service.app.Logger.Errorf("%s", err.Error())
			return nil, fmt.Errorf("Unexpected error: %s", err.Error())
		}
		if err != nil {
			return nil, err
		}
		if persistedOrg.GetID() != 0 {
			return nil, ErrOrgAlreadyExists
		}
		registration.SetOrganizationName(userCredentials.OrgName)
	}

	if err := service.regDAO.Save(registration); err != nil {
		return nil, err
	}

	mailer := NewMailer(service.app)
	mailer.SetRecipient(userCredentials.Email)
	tmpl := fmt.Sprintf("%s/%s", common.HTTP_PUBLIC_HTML, common.EMAIL_REGISTRATION)
	subject := fmt.Sprintf("%s Registration", service.app.Name)
	activationURL := fmt.Sprintf("%s/api/v1/register/activate/%d", baseURI, registrationID)
	unsubscribeURL := fmt.Sprintf("%s/api/v1/register/unsubscribe/%d", baseURI, registrationID)
	templateData := struct {
		BaseURI        template.URL
		ActivationURL  template.URL
		UnsubscribeURL template.URL
	}{
		BaseURI:        template.URL(baseURI),
		ActivationURL:  template.URL(activationURL),
		UnsubscribeURL: template.URL(unsubscribeURL),
	}
	mailer.SendHtml(tmpl, subject, templateData)

	return nil, err
}

// Activates a pending registration by creating a new user account,
// role assignment and deleting the registration record. If FARM_ACCESS_ALL
// has been configured, the user will be given permission to access all farms
// otherwise the user will need to be explicitly added to farm(s).
func (service *LocalAuthService) Activate(registrationID uint64) (common.UserAccount, error) {

	registration, err := service.regDAO.Get(registrationID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}

	userConfig := &config.User{
		ID:       service.idGenerator.NewStringID(registration.GetEmail()),
		Email:    registration.GetEmail(),
		Password: registration.GetPassword()}

	defaultRole, err := service.roleDAO.GetByName(service.app.DefaultRole, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}

	userAccount := &model.User{
		ID:       userConfig.GetID(),
		Email:    registration.GetEmail(),
		Password: registration.GetPassword()}

	orgName := registration.GetOrganizationName()
	if orgName != "" {

		org := &config.Organization{
			ID:   service.idGenerator.NewStringID(orgName),
			Name: orgName}
		service.orgDAO.Save(org)

		if service.app.DefaultPermission == common.FARM_ACCESS_ALL {
			orgs, err := service.orgDAO.GetAll(common.CONSISTENCY_QUORUM)
			orgIds := make([]uint64, len(orgs))
			if err != nil {
				return nil, err
			}
			for i, org := range orgs {
				for _, farmConfig := range org.GetFarms() {
					permission := config.NewPermission()
					permission.SetOrgID(org.GetID())
					permission.SetFarmID(farmConfig.GetID())
					permission.SetUserID(userConfig.GetID())
					permission.SetRoleID(defaultRole.GetID())
					service.permissionDAO.Save(permission)
				}
				orgIds[i] = org.GetID()
			}
			userConfig.SetOrganizationRefs(orgIds)
		}
	} else {
		if service.app.DefaultPermission == common.FARM_ACCESS_ALL {
			farms, err := service.farmDAO.GetAll(common.CONSISTENCY_LOCAL)
			farmIds := make([]uint64, len(farms))
			if err != nil {
				return nil, err
			}
			for i, farmConfig := range farms {
				permission := config.NewPermission()
				permission.SetOrgID(0)
				permission.SetFarmID(farmConfig.GetID())
				permission.SetUserID(userConfig.GetID())
				permission.SetRoleID(defaultRole.GetID())
				service.permissionDAO.Save(permission)

				farmIds[i] = farmConfig.GetID()
			}
			userConfig.SetFarmRefs(farmIds)
			// Set a default role if this is a new user
			// and there arent any farms in the system.
			if len(farms) == 0 {
				userConfig.SetRoles([]*config.Role{defaultRole})
			}
		}
	}

	if err = service.userDAO.Save(userConfig); err != nil {
		return nil, err
	}

	if err = service.regDAO.Delete(registration); err != nil {
		return userAccount, err
	}

	return userAccount, nil
}

func (service *LocalAuthService) encryptPassword(password string) (string, error) {
	hasher := util.CreatePasswordHasher(service.app.PasswordHasherParams)
	return hasher.Encrypt(password)
}

func (service *LocalAuthService) comparePassword(password, hash string) (bool, error) {
	hasher := util.CreatePasswordHasher(service.app.PasswordHasherParams)
	return hasher.Compare(password, hash)
}
