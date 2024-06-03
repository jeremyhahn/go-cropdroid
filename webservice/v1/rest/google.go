package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type GoogleAuthRequest struct {
	Token          string `json:"idToken"`
	ServerAuthCode string `json:"serverAuthCode"`
}

type GoogleRestServicer interface {
	Login(w http.ResponseWriter, r *http.Request)
	SetGoogle(w http.ResponseWriter, r *http.Request)
	RestService
}

type GoogleRestService struct {
	authService service.AuthServicer
	middleware  middleware.AuthMiddleware
	httpWriter  response.HttpWriter
	GoogleRestServicer
}

func NewGoogleRestService(
	googleAuthService service.AuthServicer,
	middleware middleware.AuthMiddleware,
	httpWriter response.HttpWriter) GoogleRestServicer {

	return &GoogleRestService{
		authService: googleAuthService,
		middleware:  middleware,
		httpWriter:  httpWriter}
}

func (restService *GoogleRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	googleEndpoint := fmt.Sprintf("%s/google", baseURI)
	googleLoginEndpoint := fmt.Sprintf("%s/login", googleEndpoint)
	//router.HandleFunc(googleLoginEndpoint, restService.Login).Methods("POST")
	router.HandleFunc(googleLoginEndpoint, restService.middleware.GenerateToken).Methods("POST")
	return []string{googleLoginEndpoint}
}

// Google login stores the idToken in "email" field and serverAuthCode in "password".
func (restService *GoogleRestService) Login(w http.ResponseWriter, r *http.Request) {

	//var authRequest GoogleAuthRequest
	var userCredentials service.UserCredentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userCredentials); err != nil {
		restService.httpWriter.Error400(w, r, err)
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
	restService.middleware.GenerateToken(w, r)
}
