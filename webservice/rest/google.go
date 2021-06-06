// +build cluster

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type GoogleAuthRequest struct {
	Token          string `json:"idToken"`
	ServerAuthCode string `json:"serverAuthCode"`
}

type GoogleRestService interface {
	SetGoogle(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultGoogleRestService struct {
	authService service.AuthService
	jwtService  service.JsonWebTokenService
	jsonWriter  common.HttpWriter
	GoogleRestService
}

func NewGoogleRestService(googleAuthService service.AuthService, jwtService service.JsonWebTokenService,
	jsonWriter common.HttpWriter) GoogleRestService {

	return &DefaultGoogleRestService{
		authService: googleAuthService,
		jwtService:  jwtService,
		jsonWriter:  jsonWriter}
}

func (restService *DefaultGoogleRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	googleEndpoint := fmt.Sprintf("%s/google", baseURI)
	googleLoginEndpoint := fmt.Sprintf("%s/login", googleEndpoint)
	//router.HandleFunc(googleLoginEndpoint, restService.Login).Methods("POST")
	router.HandleFunc(googleLoginEndpoint, restService.jwtService.GenerateToken).Methods("POST")
	return []string{googleLoginEndpoint}
}

// Google login stores the idToken in "email" field and serverAuthCode in "password".
func (restService *DefaultGoogleRestService) Login(w http.ResponseWriter, r *http.Request) {

	//var authRequest GoogleAuthRequest
	var userCredentials service.UserCredentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userCredentials); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	fmt.Printf("%+v\n", userCredentials)

	/*
		// returns userAccount, orgs, err
		userAccount, _, err := restService.authService.Login(&userCredentials)
		if err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}

		var localCredentials service.UserCredentials
		localCredentials.Email = userAccount.GetEmail()
		localCredentials.Password = userAccount.GetPassword()
		localCredentials.AuthType = common.AUTH_TYPE_GOOGLE
		bytes, err := json.Marshal(localCredentials)
		if err != nil {
			BadRequestError(w, r, err, restService.jsonWriter)
			return
		}
		r.Body = ioutil.NopCloser(strings.NewReader(string(bytes)))
	*/

	// Forward the request onto the JsonWebTokenService
	restService.jwtService.GenerateToken(w, r)
}
