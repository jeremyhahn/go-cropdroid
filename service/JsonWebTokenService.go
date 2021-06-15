package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

// https://gist.github.com/soulmachine/b368ce7292ddd7f91c15accccc02b8df

type JsonWebTokenServiceImpl struct {
	app              *app.App
	farmDAO          dao.FarmDAO
	deviceMapper mapper.DeviceMapper
	serviceRegistry  ServiceRegistry
	expiration       time.Duration
	rsaKeyPair       app.KeyPair
	jsonWriter       common.HttpWriter
	JsonWebTokenService
	Middleware
}

type farmClaim struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	//Interval int    `json:"interval"`
	//Mode     string `json:"mode"`
	//Devices []string `json:"devices"`
	Roles []string `json:"roles"`
}

type organizationClaim struct {
	ID      int                  `json:"id"`
	Name    string               `json:"name"`
	Farms   []*farmClaim         `json:"farms"`
	Roles   []string             `json:"roles"`
	License config.LicenseConfig `json:"license"`
}

type deviceClaim struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

type JsonWebTokenClaims struct {
	ServerID      int    `json:"sid"`
	UserID        int    `json:"uid"`
	Email         string `json:"email"`
	Organizations string `json:"organizations"`
	jwt.StandardClaims
}

func NewJsonWebTokenService(_app *app.App, farmDAO dao.FarmDAO, deviceMapper mapper.DeviceMapper,
	serviceRegistry ServiceRegistry, jsonWriter common.HttpWriter) (JsonWebTokenService, error) {

	keypair, err := app.NewRsaKeyPair(_app.Logger, _app.KeyDir)
	if err != nil {
		return nil, err
	}
	return CreateJsonWebTokenService(_app, farmDAO, deviceMapper, serviceRegistry, jsonWriter, 60, keypair), nil // 1 hour expiration
}

func CreateJsonWebTokenService(_app *app.App, farmDAO dao.FarmDAO, deviceMapper mapper.DeviceMapper,
	serviceRegistry ServiceRegistry, jsonWriter common.HttpWriter, expiration int64, rsaKeyPair app.KeyPair) JsonWebTokenService {

	return &JsonWebTokenServiceImpl{
		app:              _app,
		farmDAO:          farmDAO,
		deviceMapper: deviceMapper,
		serviceRegistry:  serviceRegistry,
		jsonWriter:       jsonWriter,
		expiration:       time.Duration(expiration),
		rsaKeyPair:       rsaKeyPair}
}

func (service *JsonWebTokenServiceImpl) CreateSession(w http.ResponseWriter, r *http.Request) (Session, error) {
	_, claims, err := service.parseToken(w, r)
	if err != nil {
		//http.Error(w, "Invalid token", http.StatusBadRequest)
		return nil, err
	}
	service.app.Logger.Debugf("Claims: %+v", claims)

	var roles []config.Role
	//var farmConfig *config.Farm
	var farmConfig config.FarmConfig
	var isFarmMember = false

	params := mux.Vars(r)
	orgID := 0

	if params["organizationID"] != "" {
		organizationID, err := strconv.Atoi(params["organizationID"])
		if err != nil {
			errmsg := fmt.Errorf("Missing expected organizationID HTTP GET parameter: %s", params["organizationID"])
			service.app.Logger.Error(errmsg)
			return nil, fmt.Errorf("%s", errmsg)
		}
		orgID = organizationID
	}

	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
	if err != nil {
		errmsg := fmt.Sprintf("Missing expected farmID HTTP GET parameter: %s", params["farmID"])
		service.app.Logger.Error(errmsg)
		return nil, fmt.Errorf("%s", errmsg)
	}

	var organizations []organizationClaim
	err = json.Unmarshal([]byte(claims.Organizations), &organizations)
	if err != nil {
		errmsg := fmt.Errorf("Error deserializing organization claims: %s", err)
		service.app.Logger.Error(errmsg)
		return nil, fmt.Errorf("%s", errmsg)
	}

	if orgID == 0 {

		// Try to fet the farm from global app config first
		for _, farm := range service.app.Config.GetFarms() {
			if farm.GetID() == farmID {
				farmConfig = &farm
				break
			}
		}

		// Look it up from the database
		if farmConfig == nil {
			farm, err := service.farmDAO.Get(farmID)
			if err != nil {
				service.app.Logger.Errorf("Error: %s", err)
				if err.Error() == "record not found" {
					return nil, errors.New("Farm not found")
				}
				return nil, errors.New("Internal server error")
			}
			farmConfig = farm
		}

		for _, user := range farmConfig.GetUsers() {
			if user.GetEmail() == claims.Email {
				isFarmMember = true
				roles = user.GetRoles()
				break
			}
		}

	} else {

		var isOrgMember = false

		for _, org := range organizations {
			if orgID == org.ID {
				isOrgMember = true
				break
			}
		}
		if !isOrgMember {
			errmsg := fmt.Sprintf("Not a member of this organization. Your access request has been logged.")
			service.app.Logger.Errorf("[UNAUTHORIZED] Unauthorized access attempt to organization: user=%s, org=%d", claims.Email, orgID)
			http.Error(w, errmsg, http.StatusBadRequest)
			return nil, fmt.Errorf("%s", errmsg)
		}
		orgs := service.app.Config.GetOrganizations()
		//var roles []config.RoleConfig
		//currentConfiguredOrgIDs := make([]int, 0)
		//currentConfiguredFarmIDs := make([]int, 0)
		for _, org := range orgs {
			if org.GetID() == orgID {
				for _, user := range org.GetUsers() {
					if user.GetEmail() == claims.Email {
						roles = user.GetRoles()
						break
					}
				}
				for _, farm := range org.GetFarms() {
					if farm.GetID() == farmID {
						farmConfig = &farm
						isFarmMember = true
						for _, user := range farm.GetUsers() {
							if user.GetEmail() == claims.Email {
								roles = user.GetRoles()
								break
							}
						}
					}
					//currentConfiguredFarmIDs = append(currentConfiguredFarmIDs, farm.GetID())
				}
			}
			//currentConfiguredOrgIDs = append(currentConfiguredOrgIDs, org.GetID())
		}
	}

	if !isFarmMember {
		errmsg := fmt.Sprintf("Not a member of this farm. Your access request has been logged.")
		service.app.Logger.Errorf("[UNAUTHORIZED] Unauthorized access attempt to farm: user=%s, farm=%d", claims.Email, farmID)
		http.Error(w, errmsg, http.StatusBadRequest)
	}

	commonRoles := make([]common.Role, len(roles))
	for i, role := range roles {
		commonRoles[i] = &model.Role{
			ID:   role.GetID(),
			Name: role.GetName()}
	}

	user := &model.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Roles: commonRoles}

	// TODO map[index] instead of loop
	for _, farmService := range service.serviceRegistry.GetFarmServices() {
		if farmService.GetFarmID() == farmID {
			return CreateSession(service.app.Logger, farmService, user), nil
		}
	}
	/*
		farmService, err := CreateFarmService(service.app, service.farmDAO, service.app.FarmStore,
			 farmConfig, service.deviceMapper, service.serviceRegistry)
		if err != nil {
			service.app.Logger.Errorf("Error: %s", err)
			return nil, err
		}
		return CreateSession(service.app.Logger, farmService, user), nil
	*/

	return nil, errors.New("Farm not found in service registry")
}

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
		service.app.Logger.Errorf("Error: %s", errmsg)
		return nil, nil, errors.New(errmsg)
	}
	service.app.Logger.Debugf("token=%+v", token)
	return token, claims, err
}

// GenerateToken parses an authentication HTTP request for UserCredentials and returns a JWT token or error if unsuccessful
func (service *JsonWebTokenServiceImpl) GenerateToken(w http.ResponseWriter, req *http.Request) {

	service.app.Logger.Debugf("url: %s, method: %s, remoteAddress: %s, requestUri: %s",
		req.URL.Path, req.Method, req.RemoteAddr, req.RequestURI)

	var user UserCredentials
	err := json.NewDecoder(req.Body).Decode(&user)

	service.app.Logger.Debugf("Decoded requested user: %v+", user)

	if err != nil {
		service.app.Logger.Errorf("Error: %s", err)
		service.jsonWriter.Write(w, http.StatusBadRequest, viewmodel.JsonWebToken{Error: "Bad request"})
		return
	}

	userAccount, orgs, err := service.serviceRegistry.GetUserService().Login(&user, nil)
	if err != nil {
		service.jsonWriter.Write(w, http.StatusForbidden, viewmodel.JsonWebToken{Error: "Invalid credentials"})
		return
	}

	service.app.Logger.Errorf("user: %+v", user)
	service.app.Logger.Errorf("userAccount: %+v", userAccount)
	service.app.Logger.Errorf("orgs: %+v", orgs)
	service.app.Logger.Errorf("org.len: %+v", len(orgs))
	service.app.Logger.Errorf("org[0].Farms: %+v", orgs[0].GetFarms())

	roleClaims := make([]string, len(userAccount.GetRoles()))
	for j, role := range userAccount.GetRoles() {
		roleClaims[j] = role.GetName()
	}

	orgClaims := make([]*organizationClaim, len(orgs))
	for i, org := range orgs {
		farmClaims := make([]*farmClaim, len(org.GetFarms()))
		for j, farm := range org.GetFarms() {
			farmClaims[j] = &farmClaim{
				ID:   farm.GetID(),
				Name: farm.GetName()}
			// Not sending roles here to keep JWT compact; imposes
			// logic to default farm roles to org roles on the client
			//Roles: roleClaims}
		}
		orgClaims[i] = &organizationClaim{
			ID:    org.GetID(),
			Name:  org.GetName(),
			Farms: farmClaims,
			Roles: roleClaims}
	}

	/*
		if err = service.app.ValidateLicense(); err != nil {
			// Running in unlicensed / free mode
		}
	*/

	//service.app.Logger.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	//service.app.Logger.Debugf("%+v", service.app.Config)
	//service.app.Logger.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	orgClaimsJson, err := json.Marshal(orgClaims)
	if err != nil {
		service.jsonWriter.Write(w, http.StatusInternalServerError, viewmodel.JsonWebToken{Error: "Error marshaling organization"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, JsonWebTokenClaims{
		//ServerID: service.app.Config.GetID(),
		UserID: userAccount.GetID(),
		Email:  userAccount.GetEmail(),
		//Organizations: orgClaims,
		Organizations: string(orgClaimsJson),
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

	tokenDTO := viewmodel.JsonWebToken{Value: tokenString}
	service.jsonWriter.Write(w, http.StatusOK, tokenDTO)
}

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
			/*
				var organizations []organizationClaim
				err = json.Unmarshal([]byte(claims.Organizations), &organizations)
				if err != nil {
					service.app.Logger.Errorf("Error deserializing organization claims: %s", err)
					return
				}
				if len(organizations) == 0 {
					errmsg := "Invalid request. organization claim required"
					service.app.Logger.Errorf("%s", errmsg)
					service.app.Logger.Errorf("token: %+v", token.Raw)
					http.Error(w, errmsg, http.StatusBadRequest)
					return
				}
			*/
			next(w, r)
		} else {
			service.app.Logger.Errorf("Invalid token: %s", token.Raw)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
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
