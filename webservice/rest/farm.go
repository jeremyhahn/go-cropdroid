package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type FarmRestService interface {
	Config(w http.ResponseWriter, r *http.Request)
	State(w http.ResponseWriter, r *http.Request)
	SetPermission(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultFarmRestService struct {
	publicKey   string
	farmFactory service.FarmFactory
	userService service.UserService
	middleware  service.Middleware
	jsonWriter  common.HttpWriter
	FarmRestService
}

func NewFarmRestService(publicKey string, farmFactory service.FarmFactory,
	userService service.UserService, middleware service.Middleware,
	jsonWriter common.HttpWriter) FarmRestService {
	return &DefaultFarmRestService{
		publicKey:   publicKey,
		farmFactory: farmFactory,
		userService: userService,
		middleware:  middleware,
		jsonWriter:  jsonWriter}
}

func (restService *DefaultFarmRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	// /farms/{farmID}
	//farmEndpoint := fmt.Sprintf("%s/%s", baseURI, restService.farmService.GetConfig().GetID())
	farmsEndpoint := fmt.Sprintf("%s/farms", baseURI)
	// /farms/{farmID}/users
	usersEndpoint := fmt.Sprintf("%s/users", baseFarmURI)
	// /farms/{farmID}/users/{userID}
	userEndpoint := fmt.Sprintf("%s/{userID}", usersEndpoint)
	// /farms/{farmID}/users/{userID}/role
	userRoleEndpoint := fmt.Sprintf("%s/role", userEndpoint)
	// /farms/{farmID}/config
	configEndpoint := fmt.Sprintf("%s/config", baseFarmURI)
	// /farms/{farmID}/config/{deviceID}/{key}?value=foo
	setDeviceConfigEndpoint := fmt.Sprintf("%s/{deviceID}/{key}", configEndpoint)
	// /farms/{farmID}/state
	stateEndpoint := fmt.Sprintf("%s/state", baseFarmURI)
	// /farms/{farmID}/pubkey
	pubKeyEndpoint := fmt.Sprintf("%s/pubkey", baseFarmURI)

	// /farms
	router.Handle(farmsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetFarms)),
	)).Methods("GET")
	// /farms/{farmID}/config
	router.Handle(configEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetConfig)),
	)).Methods("GET")
	// /farms/{farmID}/state
	router.Handle(stateEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetState)),
	)).Methods("GET")
	// /farms/{farmID}/config/{deviceID}/{key}?value=foo
	router.Handle(setDeviceConfigEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.SetDeviceConfig)),
	))

	// /farms/{farmID}/users
	router.Handle(usersEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.GetFarmUsers)),
	)).Methods("GET")

	// /farms/{farmID}/users/{userID}
	router.Handle(userEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.UpdateFarmUser)),
	)).Methods("POST")
	router.Handle(userEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.DeleteFarmUser)),
	)).Methods("DELETE")

	// /farms/{farmID}/users/{userID}/role
	router.Handle(userRoleEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.SetPermission)),
	)).Methods("POST")

	router.Handle(pubKeyEndpoint, http.HandlerFunc(restService.PublicKey))

	return []string{baseFarmURI, farmsEndpoint, usersEndpoint, userRoleEndpoint,
		userEndpoint, configEndpoint, stateEndpoint,
		setDeviceConfigEndpoint}
}

// Updates a user at the farm level
func (restService *DefaultFarmRestService) SetPermission(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var permission config.Permission
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&permission); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	logger.Debugf(
		"session.requestedFarmID=%d, permission=%d",
		session.GetRequestedFarmID(), permission)

	err = restService.userService.SetPermission(session, &permission)
	if err != nil {
		logger.Error(err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Success200(w, nil)
}

// Returns a list of all of the non-organization owned farms this user
// has permissions to access.
func (restService *DefaultFarmRestService) GetFarms(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	farms, err := restService.farmFactory.GetFarms(session)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("farms=%v+", farms)

	restService.jsonWriter.Write(w, http.StatusOK, farms)
}

// Returns a list of users with permission to access the requested farm
func (restService *DefaultFarmRestService) GetFarmUsers(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	farmService := session.GetFarmService()
	if farmService == nil {
		err := errors.New("farm not found")
		restService.jsonWriter.Error400(w, err)
		return
	}

	users := farmService.GetConfig().GetUsers()
	for _, user := range users {
		user.RedactPassword()
	}

	session.GetLogger().Debugf("users=%v+", users)

	restService.jsonWriter.Success200(w, users)
}

// Updates a user within the requested farm. Provides password reset feature.
func (restService *DefaultFarmRestService) UpdateFarmUser(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	var userCredentials service.UserCredentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userCredentials); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	logger.Debugf(
		"session.requestedFarmID=%d, session.user.id=%d, user.email=%s",
		session.GetRequestedFarmID(), session.GetUser().GetID(),
		userCredentials.Email)

	err = restService.userService.ResetPassword(&userCredentials)
	if err != nil {
		logger.Error(err)
		restService.jsonWriter.Error500(w, err)
		return
	}

	restService.jsonWriter.Success200(w, nil)
}

// Deletes a user from the requested farm
func (restService *DefaultFarmRestService) DeleteFarmUser(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	params := mux.Vars(r)
	uid := params["userID"]

	userID, err := strconv.ParseUint(uid, 10, 64)
	if err != nil {
		logger.Error(err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	logger.Debugf(
		"session.requestedFarmID=%d, userID=%d",
		session.GetRequestedFarmID(), userID)

	err = restService.userService.DeletePermission(session, userID)
	if err != nil {
		logger.Error(err)
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Success200(w, nil)
}

// Returns the farm configuration from the current session
func (restService *DefaultFarmRestService) GetConfig(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	logger := session.GetLogger()

	logger.Debugf("REST service /config request email=%s", session.GetUser().GetEmail())

	acceptHeader := r.Header.Get("accept")
	if acceptHeader == "application/yaml" || acceptHeader == "text/yaml" {
		NewYamlWriter().Success200(w, session.GetFarmService().GetConfig())
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, session.GetFarmService().GetConfig())
}

// Returns the current farm state from the current session
func (restService *DefaultFarmRestService) GetState(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debugf("REST service /farms/{farmID}/state request email=%s", session.GetUser().GetEmail())

	restService.jsonWriter.Write(w, http.StatusOK, session.GetFarmService().GetState())
}

// Saves a device configuration item
func (restService *DefaultFarmRestService) SetDeviceConfig(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	farmID := params["farmID"]
	deviceID := params["deviceID"]
	key := params["key"]
	value := r.FormValue("value")

	//restService.session.GetLogger().Debugf("deviceID=%s, key=%s, value=%s, params=%+v", deviceID, key, value, params)

	uint64FarmID, err := strconv.ParseUint(farmID, 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	uint64DeviceID, err := strconv.ParseUint(deviceID, 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	/*
		if err := restService.configService.SetValue(session, int(intFarmID), int(intDeviceID), key, value); err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}*/
	if err := session.GetFarmService().SetConfigValue(session, uint64FarmID, uint64DeviceID, key, value); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}

// Returns the RSA public key used to encrypt the JWT token
func (restService *DefaultFarmRestService) PublicKey(w http.ResponseWriter, r *http.Request) {

	/*
		session, err := restService.middleware.CreateSession(w, r)
		if err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}
		defer session.Close()

		session.GetLogger().Debugf("REST service /farms/{farmID}/pubkey request email=%s", session.GetUser().GetEmail())

		restService.jsonWriter.Write(w, http.StatusOK, session.GetFarmService().GetPublicKey())
	*/

	restService.jsonWriter.Write(w, http.StatusOK, restService.publicKey)
}
