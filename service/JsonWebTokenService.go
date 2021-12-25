package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

// https://gist.github.com/soulmachine/b368ce7292ddd7f91c15accccc02b8df

type JsonWebTokenServiceImpl struct {
	app             *app.App
	idGenerator     util.IdGenerator
	orgDAO          dao.OrganizationDAO
	farmDAO         dao.FarmDAO
	deviceMapper    mapper.DeviceMapper
	serviceRegistry ServiceRegistry
	expiration      time.Duration
	rsaKeyPair      app.KeyPair
	jsonWriter      common.HttpWriter
	farmProvisioner provisioner.FarmProvisioner
	defaultRole     config.RoleConfig
	JsonWebTokenService
	Middleware
}

// Claim structs are condensed models concerned only
// with users, roles, permissions, and licensing between
// the client and server. They get exchanged with every
// request and are used to generate a "Session" for working
// with services in the "service" package.
type farmClaim struct {
	ID    uint64   `json:"id"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

type organizationClaim struct {
	ID      uint64               `json:"id"`
	Name    string               `json:"name"`
	Farms   []farmClaim          `json:"farms"`
	Roles   []string             `json:"roles"`
	License config.LicenseConfig `json:"license"`
}

type JsonWebTokenClaims struct {
	ServerID      int    `json:"sid"`
	UserID        uint64 `json:"uid"`
	Email         string `json:"email"`
	Organizations string `json:"organizations"`
	Farms         string `json:"farms"`
	jwt.StandardClaims
}

// Creates a new JsonWebTokenService with default configuration
func NewJsonWebTokenService(_app *app.App, idGenerator util.IdGenerator,
	orgDAO dao.OrganizationDAO, farmDAO dao.FarmDAO,
	defaultRole config.RoleConfig, deviceMapper mapper.DeviceMapper,
	serviceRegistry ServiceRegistry, jsonWriter common.HttpWriter) (JsonWebTokenService, error) {

	keypair, err := app.NewRsaKeyPair(_app.Logger, _app.KeyDir)
	if err != nil {
		return nil, err
	}
	return CreateJsonWebTokenService(_app, idGenerator, orgDAO, farmDAO, defaultRole,
		deviceMapper, serviceRegistry, jsonWriter, 60, keypair), nil // 1 hour expiration
}

// Createa a new JsonWebBokenService with custom configuration
func CreateJsonWebTokenService(_app *app.App, idGenerator util.IdGenerator,
	orgDAO dao.OrganizationDAO, farmDAO dao.FarmDAO,
	defaultRole config.RoleConfig, deviceMapper mapper.DeviceMapper,
	serviceRegistry ServiceRegistry, jsonWriter common.HttpWriter,
	expiration int64, rsaKeyPair app.KeyPair) JsonWebTokenService {

	return &JsonWebTokenServiceImpl{
		app:             _app,
		idGenerator:     idGenerator,
		farmDAO:         farmDAO,
		deviceMapper:    deviceMapper,
		serviceRegistry: serviceRegistry,
		jsonWriter:      jsonWriter,
		expiration:      time.Duration(expiration),
		rsaKeyPair:      rsaKeyPair,
		defaultRole:     defaultRole}
}

// Creates a new "Session" object that stores the requested organization/farm context
// and membership.
func (service *JsonWebTokenServiceImpl) CreateSession(w http.ResponseWriter,
	r *http.Request) (Session, error) {

	_, claims, err := service.parseToken(w, r)
	if err != nil {
		return nil, err
	}
	service.app.Logger.Debugf("Claims: %+v", claims)

	orgClaims, err := service.parseOrganizationClaims(claims.Organizations)
	if err != nil {
		return nil, err
	}

	farmClaims, err := service.parseFarmClaims(claims.Farms)
	if err != nil {
		return nil, err
	}

	roles := make([]config.RoleConfig, 0)
	var isFarmMember = false

	// The organizationID the user is requesting to operate on
	requestedOrgID := uint64(0)

	// If an organizationID REST parameter is present, scope the session
	// to the requested organization
	params := mux.Vars(r)
	orgIdParam := params["organizationID"]
	if orgIdParam != "" {
		organizationID, err := strconv.ParseUint(orgIdParam, 10, 64)
		if err != nil {
			errmsg := fmt.Errorf("Missing expected organizationID HTTP GET parameter: %s",
				orgIdParam)
			service.app.Logger.Error(errmsg)
			return nil, errmsg
		}
		requestedOrgID = organizationID
	}

	// Make sure user belongs to the requested organization
	if requestedOrgID > 0 && !service.isOrgMember(orgClaims, requestedOrgID) {
		return nil, fmt.Errorf("not a member of requested org %d", requestedOrgID)
	}

	// If a farmID REST parameter is present, scope the session to the
	// requested farm. Errors are ignored in case the farmID is null
	// as is the case with a new system with only a default user and
	// without any farms configured.
	requestedFarmID, _ := strconv.ParseUint(params["farmID"], 10, 64)

	// If the user doesn't belong to any organizations but does belong to farms,
	// get the roles the user has been assigned for the farm being requested
	if requestedOrgID == 0 && requestedFarmID > 0 {
		farmService := service.serviceRegistry.GetFarmService(requestedFarmID)
		if farmService == nil {
			return nil, fmt.Errorf("Farm not found: %d", requestedFarmID)
		}
		for _, user := range farmService.GetConfig().GetUsers() {
			if user.GetEmail() == claims.Email {
				isFarmMember = true
				roles = user.GetRoles()
				break
			}
		}
	} else {
		// Get the roles the user has been assigned from the organization
		// and all of the farms the user has been granted permissions within
	ORGS_LOOP:
		for _, org := range orgClaims {
			if org.ID == requestedOrgID {
				for _, role := range org.Roles {
					roles = append(roles, &config.Role{Name: role})
				}
				for _, farm := range org.Farms {
					if farm.ID == requestedFarmID {
						isFarmMember = true
						for _, role := range farm.Roles {
							roles = append(roles, &config.Role{
								ID:   service.idGenerator.NewID(role),
								Name: role})
						}
						break ORGS_LOOP
					}
				}
			}
		}
	}

	// Assign a default role if no roles were found in orgs or farms
	// The default user always gets admin, everyone else is assigned
	// the system configurable default role.
	if len(roles) == 0 {
		if claims.Email == common.DEFAULT_USER {
			roles = append(roles, &config.Role{
				ID:   service.idGenerator.NewID(common.ROLE_ADMIN),
				Name: common.ROLE_ADMIN})
		} else {
			roles = append(roles, service.defaultRole)
		}
	}

	// Make sure the user is a member of the farm being requested
	if !isFarmMember && requestedFarmID > 0 {
		service.app.Logger.Errorf("[UNAUTHORIZED] Unauthorized access attempt to farm: user=%s, farm=%d", claims.Email, requestedFarmID)
		return nil, errors.New("Not a member of this farm. Your access request has been logged.")
	}

	// Create the session
	commonRoles := make([]common.Role, len(roles))
	for i, role := range roles {
		roleName := role.GetName()
		commonRoles[i] = &model.Role{
			ID:   service.idGenerator.NewID(roleName),
			Name: roleName}
	}
	user := &model.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Roles: commonRoles}

	farmService := service.serviceRegistry.GetFarmService(requestedFarmID)

	return CreateSession(service.app.Logger, orgClaims,
		farmClaims, farmService, requestedOrgID, requestedFarmID, user), nil
}

// GenerateToken parses an authentication HTTP request for UserCredentials and returns
// a JWT token or error if unsuccessful
func (service *JsonWebTokenServiceImpl) GenerateToken(w http.ResponseWriter, req *http.Request) {

	service.app.Logger.Debugf("url: %s, method: %s, remoteAddress: %s, requestUri: %s",
		req.URL.Path, req.Method, req.RemoteAddr, req.RequestURI)

	var user UserCredentials
	err := json.NewDecoder(req.Body).Decode(&user)

	service.app.Logger.Debugf("Decoded userCredentials: %v+", user)

	if err != nil {
		service.app.Logger.Errorf("Error: %s", err)
		service.jsonWriter.Write(w, http.StatusBadRequest,
			viewmodel.JsonWebToken{Error: "Bad request"})
		return
	}

	userService := service.serviceRegistry.GetUserService()
	userAccount, orgs, farms, err := userService.Login(&user)
	if err != nil {
		service.app.Logger.Errorf("GenerateToken login error: %s", err)
		service.jsonWriter.Write(w, http.StatusForbidden,
			viewmodel.JsonWebToken{Error: "Invalid credentials"})
		return
	}

	if len(userAccount.GetRoles()) == 0 {
		// Must be a new user that hasn't been assigned to any roles yet
		userAccount.SetRoles([]common.Role{service.defaultRole.(*config.Role)})
	}

	service.app.Logger.Debugf("user: %+v", user)
	service.app.Logger.Debugf("userAccount: %+v", userAccount)
	service.app.Logger.Debugf("orgs: %+v", orgs)
	service.app.Logger.Debugf("org.len: %+v", len(orgs))
	service.app.Logger.Debugf("farms.len: %+v", len(farms))

	roleClaims := make([]string, len(userAccount.GetRoles()))
	for j, role := range userAccount.GetRoles() {
		roleClaims[j] = role.GetName()
	}

	orgClaims := make([]organizationClaim, len(orgs))
	for i, org := range orgs {
		farmClaims := make([]farmClaim, len(org.GetFarms()))
		for j, farm := range org.GetFarms() {
			farmClaims[j] = farmClaim{
				ID:   farm.GetID(),
				Name: farm.GetName()}
			// Not sending roles here to keep JWT compact; imposes
			// logic to default farm roles to org roles on the client
			//Roles: roleClaims}
		}
		orgClaims[i] = organizationClaim{
			ID:    org.GetID(),
			Name:  org.GetName(),
			Farms: farmClaims,
			Roles: roleClaims}
	}
	orgClaimsJson, err := json.Marshal(orgClaims)
	if err != nil {
		service.jsonWriter.Write(w, http.StatusInternalServerError, viewmodel.JsonWebToken{Error: "Error marshaling organization"})
		return
	}

	farmClaims := make([]farmClaim, len(farms))
	for i, farm := range farms {
		farmClaims[i] = farmClaim{
			ID:   farm.GetID(),
			Name: farm.GetName()}
	}
	farmClaimsJson, err := json.Marshal(farmClaims)
	if err != nil {
		service.jsonWriter.Write(w, http.StatusInternalServerError, viewmodel.JsonWebToken{Error: "Error marshaling farms"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, JsonWebTokenClaims{
		ServerID:      service.app.Config.GetID(),
		UserID:        userAccount.GetID(),
		Email:         userAccount.GetEmail(),
		Organizations: string(orgClaimsJson),
		Farms:         string(farmClaimsJson),
		StandardClaims: jwt.StandardClaims{
			Issuer:    common.APPNAME,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * service.expiration).Unix()}})

	tokenString, err := token.SignedString(service.rsaKeyPair.GetPrivateKey())
	if err != nil {
		service.jsonWriter.Write(w, http.StatusInternalServerError, viewmodel.JsonWebToken{Error: "Error signing token"})
		return
	}

	service.app.Logger.Debugf("Genearted JSON token: %s", tokenString)

	jwtViewModel := viewmodel.JsonWebToken{Value: tokenString}
	service.jsonWriter.Write(w, http.StatusOK, jwtViewModel)
}

// Renews a JWT token with a fresh expiration date
func (service *JsonWebTokenServiceImpl) RefreshToken(w http.ResponseWriter, req *http.Request) {
	token, claims, err := service.parseToken(w, req)
	if err == nil {
		if token.Valid {
			userService := service.serviceRegistry.GetUserService()
			userAccount, orgs, farms, err := userService.Refresh(claims.UserID)
			if err != nil {
				service.app.Logger.Errorf("Error refreshing token: %s", err)
				service.jsonWriter.Write(w, http.StatusUnauthorized, viewmodel.JsonWebToken{Error: "Invalid token"})
				return
			}

			if len(userAccount.GetRoles()) == 0 {
				// Must be a new user
				userAccount.SetRoles([]common.Role{service.defaultRole.(*config.Role)})
			}

			roleClaims := make([]string, len(userAccount.GetRoles()))
			for j, role := range userAccount.GetRoles() {
				roleClaims[j] = role.GetName()
			}

			orgClaims := make([]*organizationClaim, len(orgs))
			for i, org := range orgs {
				farmClaims := make([]farmClaim, len(org.GetFarms()))
				for j, farm := range org.GetFarms() {

					roles := make([]string, 0)
					for _, user := range farm.GetUsers() {
						if user.GetID() == userAccount.GetID() {
							for _, role := range user.GetRoles() {
								roles = append(roles, role.GetName())
							}
						}
					}
					farmClaims[j] = farmClaim{
						ID:    farm.GetID(),
						Name:  farm.GetName(),
						Roles: roles}
				}
				orgClaims[i] = &organizationClaim{
					ID:    org.GetID(),
					Name:  org.GetName(),
					Farms: farmClaims,
					Roles: roleClaims}
			}
			orgClaimsJson, err := json.Marshal(orgClaims)
			if err != nil {
				service.jsonWriter.Write(w, http.StatusInternalServerError, viewmodel.JsonWebToken{Error: "Error marshaling organization"})
				return
			}

			farmClaims := make([]farmClaim, len(farms))
			for i, farm := range farms {
				roles := make([]string, 0)
				for _, user := range farm.GetUsers() {
					if user.GetID() == userAccount.GetID() {
						for _, role := range user.GetRoles() {
							roles = append(roles, role.GetName())
						}
					}
				}
				farmClaims[i] = farmClaim{
					ID:    farm.GetID(),
					Name:  farm.GetName(),
					Roles: roles}
			}
			farmClaimsJson, err := json.Marshal(farmClaims)
			if err != nil {
				service.jsonWriter.Write(w, http.StatusInternalServerError, viewmodel.JsonWebToken{Error: "Error marshaling farms"})
				return
			}

			token := jwt.NewWithClaims(jwt.SigningMethodRS256, JsonWebTokenClaims{
				ServerID:      service.app.Config.GetID(),
				UserID:        userAccount.GetID(),
				Email:         userAccount.GetEmail(),
				Organizations: string(orgClaimsJson),
				Farms:         string(farmClaimsJson),
				StandardClaims: jwt.StandardClaims{
					Issuer:    common.APPNAME,
					IssuedAt:  time.Now().Unix(),
					ExpiresAt: time.Now().Add(time.Minute * service.expiration).Unix()}})

			tokenString, err := token.SignedString(service.rsaKeyPair.GetPrivateKey())
			if err != nil {
				service.jsonWriter.Write(w, http.StatusInternalServerError, viewmodel.JsonWebToken{Error: "Error signing token"})
				return
			}

			service.app.Logger.Debugf("Refreshed JSON token: %s", tokenString)

			tokenDTO := viewmodel.JsonWebToken{Value: tokenString}
			service.jsonWriter.Write(w, http.StatusOK, tokenDTO)

		} else {
			service.app.Logger.Errorf("Invalid token: %s", token.Raw)
			service.jsonWriter.Write(w, http.StatusUnauthorized, viewmodel.JsonWebToken{Error: "Invalid token"})
		}
	} else {
		errmsg := err.Error()
		if errmsg == "no token present in request" {
			errmsg = "Authentication required"
		}
		service.app.Logger.Errorf("Error: %s", errmsg)
		http.Error(w, errmsg, http.StatusBadRequest)
	}
}

// Validates the raw JWT token to ensure it's not expired or contains any invalid claims
func (service *JsonWebTokenServiceImpl) Validate(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	token, claims, err := service.parseToken(w, r)
	if err == nil {
		if token.Valid {
			if claims.UserID <= 0 {
				errmsg := "Invalid request. id claim required."
				service.app.Logger.Errorf("%s", errmsg)
				service.app.Logger.Errorf("token: %+v", token.Raw)
				http.Error(w, errmsg, http.StatusBadRequest)
				return
			}
			if claims.Email == "" {
				errmsg := "Invalid request. email claim required"
				service.app.Logger.Errorf("%s", errmsg)
				service.app.Logger.Errorf("token: %+v", token.Raw)
				http.Error(w, errmsg, http.StatusBadRequest)
				return
			}
			next(w, r)
		} else {
			service.app.Logger.Errorf("invalid token: %s", token.Raw)
			http.Error(w, "invalid token", http.StatusUnauthorized)
		}
	} else {
		errmsg := err.Error()
		if errmsg == "no token present in request" {
			errmsg = "Authentication required"
		}
		service.app.Logger.Errorf("Error: %s", errmsg)
		http.Error(w, errmsg, http.StatusBadRequest)
	}
}

// Used to determine if the specified organization is a member of any of the specified organizationClaims
func (service *JsonWebTokenServiceImpl) isOrgMember(orgClaims []organizationClaim, orgID uint64) bool {
	for _, org := range orgClaims {
		if org.ID == orgID {
			return true
		}
	}
	return false
}

// Parses a list of organizationClaims from a json string
func (service *JsonWebTokenServiceImpl) parseOrganizationClaims(orgJson string) ([]organizationClaim, error) {
	var orgClaims []organizationClaim
	reader := strings.NewReader(orgJson)
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&orgClaims); err != nil {
		service.app.Logger.Errorf("parseOrganizationClaims error: %s", err)
		return []organizationClaim{}, err
	}
	return orgClaims, nil
}

// Parses a list of organizationClaims from a json string
func (service *JsonWebTokenServiceImpl) parseFarmClaims(farmJson string) ([]farmClaim, error) {
	var farmClaims []farmClaim
	reader := strings.NewReader(farmJson)
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&farmClaims); err != nil {
		service.app.Logger.Errorf("parseFarmClaims error: %s", err)
		return []farmClaim{}, err
	}
	return farmClaims, nil
}

// Parses the JsonWebTokenClaims from the HTTP request
func (service *JsonWebTokenServiceImpl) parseClaims(r *http.Request, extractor request.Extractor) (*jwt.Token, *JsonWebTokenClaims, error) {
	token, err := request.ParseFromRequest(r, extractor,
		func(token *jwt.Token) (interface{}, error) {
			return service.rsaKeyPair.GetPublicKey(), nil
		})
	if err != nil {
		return nil, nil, err
	}
	claims := &JsonWebTokenClaims{}
	_, err = jwt.ParseWithClaims(token.Raw, claims,
		func(token *jwt.Token) (interface{}, error) {
			return service.rsaKeyPair.GetPublicKey(), nil
		})
	if err != nil {
		return nil, nil, err
	}
	service.app.Logger.Debugf("claims: %+v", claims)
	return token, claims, nil
}

// Parses the JsonWebTokenClaims from the HTTP request using either an OAuth2 or
// Authorization header based on their presence in the HTTP request.
func (service *JsonWebTokenServiceImpl) parseToken(w http.ResponseWriter, r *http.Request) (*jwt.Token, *JsonWebTokenClaims, error) {
	var token *jwt.Token
	var claims *JsonWebTokenClaims
	var err error
	if _, ok := r.URL.Query()["access_token"]; ok {
		t, c, e := service.parseClaims(r, request.OAuth2Extractor)
		token = t
		claims = c
		err = e
	} else {
		t, c, e := service.parseClaims(r, request.AuthorizationHeaderExtractor)
		token = t
		claims = c
		err = e
	}
	if err != nil {
		errmsg := err.Error()
		service.app.Logger.Errorf("parseToken error: %s", errmsg)
		return nil, nil, errors.New(errmsg)
	}
	service.app.Logger.Debugf("token=%+v", token)
	return token, claims, err
}
