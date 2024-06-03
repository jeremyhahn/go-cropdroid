package rest

import (
	"crypto/rsa"
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
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

// https://gist.github.com/soulmachine/b368ce7292ddd7f91c15accccc02b8df

type JsonWebTokenServicer interface {
	ParseToken(r *http.Request, extractor request.Extractor) (*jwt.Token, *JsonWebTokenClaims, error)
	middleware.AuthMiddleware
	middleware.JsonWebTokenMiddleware
}

type JWTService struct {
	app             *app.App
	idGenerator     util.IdGenerator
	deviceMapper    mapper.DeviceMapper
	serviceRegistry service.ServiceRegistry
	expiration      time.Duration
	responseWriter  response.HttpWriter
	defaultRole     *config.RoleStruct
	publicKey       *rsa.PublicKey
	JsonWebTokenServicer
	middleware.JsonWebTokenMiddleware
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
func NewJsonWebTokenService(
	app *app.App,
	idGenerator util.IdGenerator,
	defaultRole *config.RoleStruct,
	deviceMapper mapper.DeviceMapper,
	serviceRegistry service.ServiceRegistry,
	responseWriter response.HttpWriter) (JsonWebTokenServicer, error) {

	return CreateJsonWebTokenService(app, idGenerator, defaultRole,
		deviceMapper, serviceRegistry, responseWriter, app.JwtExpiration)
}

// Createa a new JsonWebBokenService with custom expiration
func CreateJsonWebTokenService(
	app *app.App,
	idGenerator util.IdGenerator,
	defaultRole *config.RoleStruct,
	deviceMapper mapper.DeviceMapper,
	serviceRegistry service.ServiceRegistry,
	responseWriter response.HttpWriter,
	expiration int) (JsonWebTokenServicer, error) {

	publicKey, err := app.CA.PublicKey("ca")
	if err != nil {
		return nil, err
	}

	// Don't store the private key, just attempt to load it
	// for now so an error is returned if it doesn't exist
	// and it's safe to ignore errors later.
	if _, err = app.CA.PrivateKey(app.Domain); err != nil {
		return nil, err
	}

	return &JWTService{
		app:             app,
		idGenerator:     idGenerator,
		deviceMapper:    deviceMapper,
		serviceRegistry: serviceRegistry,
		responseWriter:  responseWriter,
		expiration:      time.Duration(expiration),
		defaultRole:     defaultRole,
		publicKey:       publicKey}, nil
}

// Returns the RSA private key for the web server, ignoring errors. The
// constructor already makes sure the private key exists when the service
// is instantiated.
func (jwtService *JWTService) privateKey() *rsa.PrivateKey {
	privateKey, _ := jwtService.app.CA.PrivateKey("ca")
	return privateKey
}

func (jwtService *JWTService) CreateSession(w http.ResponseWriter,
	r *http.Request) (service.Session, error) {

	jwtService.app.Logger.Debugf("url: %s, method: %s, remoteAddress: %s, requestUri: %s",
		r.URL.Path, r.Method, r.RemoteAddr, r.RequestURI)

	_, claims, err := jwtService.parseToken(w, r)
	if err != nil {
		return nil, err
	}
	jwtService.app.Logger.Debugf("Claims: %+v", claims)

	orgClaims, err := jwtService.parseOrganizationClaims(claims.Organizations)
	if err != nil {
		return nil, err
	}

	FarmClaims, err := jwtService.parseFarmClaims(claims.Farms)
	if err != nil {
		return nil, err
	}

	roles := make([]*config.RoleStruct, 0)
	var isFarmMember = false
	var consistencyLevel = common.CONSISTENCY_LOCAL

	// The organizationID the user is requesting to operate on
	var requestedOrgID = uint64(0)

	// If an organizationID REST parameter is present, scope the session
	// to the requested organization
	params := mux.Vars(r)
	orgIdParam := params["organizationID"]
	if orgIdParam != "" {
		organizationID, err := strconv.ParseUint(orgIdParam, 10, 64)
		if err != nil {
			errmsg := fmt.Errorf("missing expected organizationID HTTP GET parameter: %s",
				orgIdParam)
			jwtService.app.Logger.Error(errmsg)
			return nil, errmsg
		}
		requestedOrgID = organizationID
	}

	// Make sure user belongs to the requested organization
	if requestedOrgID > 0 && !jwtService.isOrgMember(orgClaims, requestedOrgID) {
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
		farmService := jwtService.serviceRegistry.GetFarmService(requestedFarmID)
		if farmService == nil {
			return nil, fmt.Errorf("farm not found: %d", requestedFarmID)
		}
		farmConfig := farmService.GetConfig()
		for _, user := range farmConfig.GetUsers() {
			if user.GetEmail() == claims.Email {
				isFarmMember = true
				roles = user.GetRoles()
				break
			}
		}
		consistencyLevel = farmConfig.GetConsistencyLevel()
	} else {
		// Get the roles the user has been assigned from the organization
		// and all of the farms the user has been granted permissions within
	ORGS_LOOP:
		for _, org := range orgClaims {
			if org.ID == requestedOrgID {
				for _, role := range org.Roles {
					roles = append(roles, &config.RoleStruct{Name: role})
				}
				for _, farm := range org.Farms {
					if farm.ID == requestedFarmID {
						isFarmMember = true
						for _, role := range farm.Roles {
							roles = append(roles, &config.RoleStruct{
								ID:   jwtService.idGenerator.NewStringID(role),
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
			roles = append(roles, &config.RoleStruct{
				ID:   jwtService.idGenerator.NewStringID(common.ROLE_ADMIN),
				Name: common.ROLE_ADMIN})
		} else {
			roles = append(roles, jwtService.defaultRole)
		}
	}

	// Make sure the user is a member of the farm being requested
	if !isFarmMember && requestedFarmID > 0 {
		jwtService.app.Logger.Errorf("[UNAUTHORIZED] Unauthorized access attempt to farm: user=%s, farm=%d", claims.Email, requestedFarmID)
		return nil, errors.New("not a member of the requested farm, access request has been logged")
	}

	// Create the session
	commonRoles := make([]model.Role, len(roles))
	for i, role := range roles {
		roleName := role.GetName()
		commonRoles[i] = &model.RoleStruct{
			ID:   jwtService.idGenerator.NewStringID(roleName),
			Name: roleName}
	}
	user := &model.UserStruct{
		ID:    claims.UserID,
		Email: claims.Email,
		Roles: commonRoles}

	farmService := jwtService.serviceRegistry.GetFarmService(requestedFarmID)

	return service.CreateSession(jwtService.app.Logger, orgClaims,
		FarmClaims, farmService, requestedOrgID, requestedFarmID, consistencyLevel, user), nil
}

func (jwtService *JWTService) GenerateToken(w http.ResponseWriter, req *http.Request) {

	jwtService.app.Logger.Debugf("url: %s, method: %s, remoteAddress: %s, requestUri: %s",
		req.URL.Path, req.Method, req.RemoteAddr, req.RequestURI)

	var user service.UserCredentials
	err := json.NewDecoder(req.Body).Decode(&user)
	// jwtService.app.Logger.Debugf("Decoded userCredentials: %v+", user)
	if err != nil {
		jwtService.responseWriter.Error400(w, req, err)
		return
	}

	userService := jwtService.serviceRegistry.GetUserService()
	userAccount, orgs, farms, err := userService.Login(&user)
	if err != nil {
		jwtService.app.Logger.Errorf("GenerateToken login error: %s", err)
		jwtService.responseWriter.Write(w, req, http.StatusForbidden,
			viewmodel.JsonWebToken{Error: "Invalid credentials"})
		return
	}

	if len(userAccount.GetRoles()) == 0 {
		// Must be a new user that hasn't been assigned to any roles yet
		userAccount.SetRoles([]model.Role{
			&model.RoleStruct{
				ID:   jwtService.defaultRole.ID,
				Name: jwtService.defaultRole.Name}})
	}

	jwtService.app.Logger.Debugf("user: %+v", user)
	jwtService.app.Logger.Debugf("userAccount: %+v", userAccount)
	jwtService.app.Logger.Debugf("orgs: %+v", orgs)
	jwtService.app.Logger.Debugf("org.len: %+v", len(orgs))
	jwtService.app.Logger.Debugf("farms.len: %+v", len(farms))

	roleClaims := make([]string, len(userAccount.GetRoles()))
	for j, role := range userAccount.GetRoles() {
		roleClaims[j] = role.GetName()
	}

	orgClaims := make([]service.OrganizationClaim, len(orgs))
	for i, org := range orgs {
		FarmClaims := make([]service.FarmClaim, len(org.GetFarms()))
		for j, farm := range org.GetFarms() {
			FarmClaims[j] = service.FarmClaim{
				ID:   farm.ID,
				Name: farm.GetName()}
			// Not sending roles here to keep JWT compact; imposes
			// logic to default farm roles to org roles on the client
			//Roles: roleClaims}
		}
		orgClaims[i] = service.OrganizationClaim{
			ID:    org.Identifier(),
			Name:  org.GetName(),
			Farms: FarmClaims,
			Roles: roleClaims}
	}
	orgClaimsJson, err := json.Marshal(orgClaims)
	if err != nil {
		jwtService.responseWriter.Write(w, req, http.StatusInternalServerError,
			viewmodel.JsonWebToken{Error: "Error marshaling organization"})
		return
	}

	farmClaims := make([]service.FarmClaim, len(farms))
	for i, farm := range farms {
		farmClaims[i] = service.FarmClaim{
			ID:   farm.Identifier(),
			Name: farm.GetName()}
	}
	farmClaimsJson, err := json.Marshal(farmClaims)
	if err != nil {
		jwtService.responseWriter.Write(w, req, http.StatusInternalServerError,
			viewmodel.JsonWebToken{Error: "Error marshaling farms"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, JsonWebTokenClaims{
		ServerID:      int(jwtService.app.NodeID),
		UserID:        userAccount.Identifier(),
		Email:         userAccount.GetEmail(),
		Organizations: string(orgClaimsJson),
		Farms:         string(farmClaimsJson),
		StandardClaims: jwt.StandardClaims{
			Issuer:    common.APPNAME,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * jwtService.expiration).Unix()}})

	tokenString, err := token.SignedString(jwtService.privateKey())
	if err != nil {
		jwtService.responseWriter.Write(w, req,
			http.StatusInternalServerError, viewmodel.JsonWebToken{Error: "Error signing token"})
		return
	}

	jwtService.app.Logger.Debugf("Generated JSON token: %s", tokenString)

	jwtViewModel := viewmodel.JsonWebToken{Value: tokenString}
	jwtService.responseWriter.Write(w, req, http.StatusOK, jwtViewModel)
}

func (jwtService *JWTService) RefreshToken(w http.ResponseWriter, req *http.Request) {

	jwtService.app.Logger.Debugf("url: %s, method: %s, remoteAddress: %s, requestUri: %s",
		req.URL.Path, req.Method, req.RemoteAddr, req.RequestURI)

	token, claims, err := jwtService.parseToken(w, req)
	if err == nil {
		if token.Valid {
			userService := jwtService.serviceRegistry.GetUserService()
			userAccount, orgs, farms, err := userService.Refresh(claims.UserID)
			if err != nil {
				jwtService.app.Logger.Errorf("Error refreshing token: %s", err)
				jwtService.responseWriter.Write(w, req, http.StatusUnauthorized,
					viewmodel.JsonWebToken{Error: "Invalid token"})
				return
			}

			if len(userAccount.GetRoles()) == 0 {
				// Must be a new user
				userAccount.SetRoles([]model.Role{
					&model.RoleStruct{
						ID:   jwtService.defaultRole.ID,
						Name: jwtService.defaultRole.Name}})
			}

			roleClaims := make([]string, len(userAccount.GetRoles()))
			for j, role := range userAccount.GetRoles() {
				roleClaims[j] = role.GetName()
			}

			orgClaims := make([]*service.OrganizationClaim, len(orgs))
			for i, org := range orgs {
				FarmClaims := make([]service.FarmClaim, len(org.GetFarms()))
				for j, farm := range org.GetFarms() {

					roles := make([]string, 0)
					for _, user := range farm.GetUsers() {
						if user.ID == userAccount.Identifier() {
							for _, role := range user.GetRoles() {
								roles = append(roles, role.GetName())
							}
						}
					}
					FarmClaims[j] = service.FarmClaim{
						ID:    farm.Identifier(),
						Name:  farm.GetName(),
						Roles: roles}
				}
				orgClaims[i] = &service.OrganizationClaim{
					ID:    org.Identifier(),
					Name:  org.GetName(),
					Farms: FarmClaims,
					Roles: roleClaims}
			}
			orgClaimsJson, err := json.Marshal(orgClaims)
			if err != nil {
				jwtService.responseWriter.Write(w, req, http.StatusInternalServerError,
					viewmodel.JsonWebToken{Error: "Error marshaling organization"})
				return
			}

			FarmClaims := make([]service.FarmClaim, len(farms))
			for i, farm := range farms {
				roles := make([]string, 0)
				for _, user := range farm.GetUsers() {
					if user.ID == userAccount.Identifier() {
						for _, role := range user.GetRoles() {
							roles = append(roles, role.GetName())
						}
					}
				}
				FarmClaims[i] = service.FarmClaim{
					ID:    farm.Identifier(),
					Name:  farm.GetName(),
					Roles: roles}
			}
			FarmClaimsJson, err := json.Marshal(FarmClaims)
			if err != nil {
				jwtService.responseWriter.Write(w, req, http.StatusInternalServerError,
					viewmodel.JsonWebToken{Error: "Error marshaling farms"})
				return
			}

			token := jwt.NewWithClaims(jwt.SigningMethodRS256, JsonWebTokenClaims{
				ServerID:      int(jwtService.app.NodeID),
				UserID:        userAccount.Identifier(),
				Email:         userAccount.GetEmail(),
				Organizations: string(orgClaimsJson),
				Farms:         string(FarmClaimsJson),
				StandardClaims: jwt.StandardClaims{
					Issuer:    common.APPNAME,
					IssuedAt:  time.Now().Unix(),
					ExpiresAt: time.Now().Add(time.Minute * jwtService.expiration).Unix()}})

			tokenString, err := token.SignedString(jwtService.privateKey())
			if err != nil {
				jwtService.responseWriter.Write(w, req, http.StatusInternalServerError,
					viewmodel.JsonWebToken{Error: "Error signing token"})
				return
			}

			jwtService.app.Logger.Debugf("Refreshed JSON token: %s", tokenString)

			tokenDTO := viewmodel.JsonWebToken{Value: tokenString}
			jwtService.responseWriter.Write(w, req, http.StatusOK, tokenDTO)

		} else {
			jwtService.app.Logger.Errorf("Invalid token: %s", token.Raw)
			jwtService.responseWriter.Write(w, req, http.StatusUnauthorized,
				viewmodel.JsonWebToken{Error: "Invalid token"})
		}
	} else {
		errmsg := err.Error()
		if errmsg == "no token present in request" {
			errmsg = "Authentication required"
		}
		jwtService.app.Logger.Errorf("Error: %s", errmsg)
		http.Error(w, errmsg, http.StatusBadRequest)
	}
}

// Validates the raw JWT token to ensure it's not expired or contains any invalid claims. This
// is used by the negroni middleware to enforce authenticated access to procted resources.
func (jwtService *JWTService) Validate(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	jwtService.app.Logger.Debugf("url: %s, method: %s, remoteAddress: %s, requestUri: %s",
		r.URL.Path, r.Method, r.RemoteAddr, r.RequestURI)

	token, claims, err := jwtService.parseToken(w, r)
	if err == nil {
		if token.Valid {
			if claims.UserID <= 0 {
				errmsg := "Invalid request. id claim required."
				jwtService.app.Logger.Errorf("%s", errmsg)
				jwtService.app.Logger.Errorf("token: %+v", token.Raw)
				http.Error(w, errmsg, http.StatusBadRequest)
				return
			}
			if claims.Email == "" {
				errmsg := "Invalid request. email claim required"
				jwtService.app.Logger.Errorf("%s", errmsg)
				jwtService.app.Logger.Errorf("token: %+v", token.Raw)
				http.Error(w, errmsg, http.StatusBadRequest)
				return
			}
			next(w, r)
		} else {
			jwtService.app.Logger.Errorf("invalid token: %s", token.Raw)
			http.Error(w, "invalid token", http.StatusUnauthorized)
		}
	} else {
		errmsg := err.Error()
		if errmsg == "no token present in request" {
			errmsg = "Authentication required"
		}
		jwtService.app.Logger.Errorf("Error: %s", errmsg)
		http.Error(w, errmsg, http.StatusBadRequest)
	}
}

// Used to determine if the specified organization is a member of any of the specified OrganizationClaims
func (jwtService *JWTService) isOrgMember(orgClaims []service.OrganizationClaim, orgID uint64) bool {
	for _, org := range orgClaims {
		if org.ID == orgID {
			return true
		}
	}
	return false
}

// Parses a list of OrganizationClaims from a json string
func (jwtService *JWTService) parseOrganizationClaims(orgJson string) ([]service.OrganizationClaim, error) {
	var orgClaims []service.OrganizationClaim
	reader := strings.NewReader(orgJson)
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&orgClaims); err != nil {
		jwtService.app.Logger.Errorf("parseOrganizationClaims error: %s", err)
		return []service.OrganizationClaim{}, err
	}
	return orgClaims, nil
}

// Parses a list of OrganizationClaims from a json string
func (jwtService *JWTService) parseFarmClaims(farmJson string) ([]service.FarmClaim, error) {
	var FarmClaims []service.FarmClaim
	reader := strings.NewReader(farmJson)
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&FarmClaims); err != nil {
		jwtService.app.Logger.Errorf("parseFarmClaims error: %s", err)
		return []service.FarmClaim{}, err
	}
	return FarmClaims, nil
}

// Parses the JsonWebTokenClaims from the HTTP request
func (jwtService *JWTService) parseClaims(r *http.Request, extractor request.Extractor) (*jwt.Token, *JsonWebTokenClaims, error) {
	token, err := request.ParseFromRequest(r, extractor,
		func(token *jwt.Token) (interface{}, error) {
			return jwtService.publicKey, nil
		})
	if err != nil {
		return nil, nil, err
	}
	claims := &JsonWebTokenClaims{}
	_, err = jwt.ParseWithClaims(token.Raw, claims,
		func(token *jwt.Token) (interface{}, error) {
			return jwtService.publicKey, nil
		})
	if err != nil {
		return nil, nil, err
	}
	jwtService.app.Logger.Debugf("claims: %+v", claims)
	return token, claims, nil
}

// Parses the JsonWebTokenClaims from the HTTP request using either an OAuth2 or
// Authorization header based on their presence in the HTTP request.
func (jwtService *JWTService) parseToken(w http.ResponseWriter, r *http.Request) (*jwt.Token, *JsonWebTokenClaims, error) {
	var token *jwt.Token
	var claims *JsonWebTokenClaims
	var err error
	if _, ok := r.URL.Query()["access_token"]; ok {
		t, c, e := jwtService.parseClaims(r, request.OAuth2Extractor)
		token = t
		claims = c
		err = e
	} else {
		t, c, e := jwtService.parseClaims(r, request.AuthorizationHeaderExtractor)
		token = t
		claims = c
		err = e
	}
	if err != nil {
		errmsg := err.Error()
		jwtService.app.Logger.Errorf("parseToken error: %s", errmsg)
		return nil, nil, errors.New(errmsg)
	}
	jwtService.app.Logger.Debugf("token=%+v", token)
	return token, claims, err
}
