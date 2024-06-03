package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/pki/ca"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type FarmRestServicer interface {
	Devices(w http.ResponseWriter, r *http.Request)
	Farms(w http.ResponseWriter, r *http.Request)
	FarmUsers(w http.ResponseWriter, r *http.Request)
	SetPermission(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
	DeleteFarmUser(w http.ResponseWriter, r *http.Request)
	Config(w http.ResponseWriter, r *http.Request)
	State(w http.ResponseWriter, r *http.Request)
	SetDeviceConfig(w http.ResponseWriter, r *http.Request)
	PublicKey(w http.ResponseWriter, r *http.Request)
	SendMessage(w http.ResponseWriter, r *http.Request)
}

type FarmRestService struct {
	domain               string
	certificateAuthority ca.CertificateAuthority
	farmFactory          service.FarmFactory
	userService          service.UserServicer
	notificationService  service.NotificationServicer
	middleware           middleware.JsonWebTokenMiddleware
	httpWriter           response.HttpWriter
	FarmRestServicer
}

func NewFarmRestService(
	domain string,
	certificateAuthority ca.CertificateAuthority,
	farmFactory service.FarmFactory,
	userService service.UserServicer,
	notificationService service.NotificationServicer,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) FarmRestServicer {

	return &FarmRestService{
		domain:               domain,
		certificateAuthority: certificateAuthority,
		farmFactory:          farmFactory,
		userService:          userService,
		middleware:           middleware,
		httpWriter:           httpWriter}
}

// Returns a list of all non-organization owned farms this requested user has permissions to access
func (restService *FarmRestService) Farms(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	farms, err := restService.farmFactory.GetFarms(session)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Write(w, r, http.StatusOK, farms)
}

// Returns a list of users with access to the farm
func (restService *FarmRestService) FarmUsers(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	farmService := session.GetFarmService()
	if farmService == nil {
		restService.httpWriter.Error400(w, r, errors.New("farm not found"))
		return
	}
	users := farmService.GetConfig().GetUsers()
	for _, user := range users {
		user.RedactPassword()
	}
	restService.httpWriter.Success200(w, r, users)
}

// Updates a users role
func (restService *FarmRestService) SetPermission(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var permission *config.PermissionStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(permission); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	err = restService.userService.SetPermission(session, permission)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

// Updates a user within the requested farm, including the password.
func (restService *FarmRestService) ResetPassword(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var userCredentials service.UserCredentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userCredentials); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	err = restService.userService.ResetPassword(&userCredentials)
	if err != nil {
		restService.httpWriter.Error500(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

// Deletes a user from the requested farm
func (restService *FarmRestService) DeleteFarmUser(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()

	params := mux.Vars(r)
	uid := params["userID"]
	userID, err := strconv.ParseUint(uid, 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	err = restService.userService.DeletePermission(session, userID)
	if err != nil {
		restService.httpWriter.Error500(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

// Returns the farm configuration from the current session
func (restService *FarmRestService) GetConfig(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	restService.httpWriter.Success200(w, r, session.GetFarmService().GetConfig())
}

// Returns the current farm state from the current session
func (restService *FarmRestService) GetState(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	restService.httpWriter.Write(w, r, http.StatusOK, session.GetFarmService().GetState())
}

func (restService *FarmRestService) Devices(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	//devices, err := restService.deviceFactory.Devices(session)
	devices, err := session.GetFarmService().Devices()
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, devices)
}

// Saves a device configuration item
func (restService *FarmRestService) SetDeviceConfig(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
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
		restService.httpWriter.Error400(w, r, err)
		return
	}
	uint64DeviceID, err := strconv.ParseUint(deviceID, 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err := session.GetFarmService().SetConfigValue(session, uint64FarmID, uint64DeviceID, key, value); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}

// Returns the web server RSA public key used to encrypt the JWT
func (restService *FarmRestService) PublicKey(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	farmID := params["farmID"]

	// Parse to uint64 to make sure its a valid farm ID
	_, err := strconv.ParseUint(farmID, 10, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}

	// Try to load a certificate issued to the farm
	cert, err := restService.certificateAuthority.PEM(farmID)
	if err == ca.ErrCertNotFound {
		// Fall back to the web server certificate
		cert, err = restService.certificateAuthority.PEM(restService.domain)
		if err == ca.ErrCertNotFound {
			restService.httpWriter.Error404(w, r, ca.ErrCertNotFound)
			return
		}
	} else if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, string(cert))
}

func (restService *FarmRestService) SendMessage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	_type := params["type"]
	message := params["message"]
	priority, err := strconv.Atoi(params["priority"])
	if err != nil {
		priority = common.NOTIFICATION_PRIORITY_LOW
	}
	restService.notificationService.Enqueue(&model.NotificationStruct{
		Device:    "webserver",
		Priority:  priority,
		Type:      _type,
		Message:   message,
		Timestamp: time.Now()})
}
