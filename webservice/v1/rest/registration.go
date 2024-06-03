package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type RegistrationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegistrationResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type RegistrationRestServicer interface {
	Register(w http.ResponseWriter, r *http.Request)
	Activate(w http.ResponseWriter, r *http.Request)
}

type RegistrationRestService struct {
	app         *app.App
	userService service.UserServicer
	jsonWriter  response.HttpWriter
}

func NewRegistrationRestService(
	app *app.App,
	userService service.UserServicer,
	jsonWriter response.HttpWriter) RegistrationRestServicer {

	return &RegistrationRestService{
		app:         app,
		userService: userService,
		jsonWriter:  jsonWriter}
}

// Register a new user and send an activation email
func (restService *RegistrationRestService) Register(w http.ResponseWriter, r *http.Request) {
	var userCredentials service.UserCredentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userCredentials); err != nil {
		restService.jsonWriter.Error400(w, r, err)
		return
	}
	serverURI := fmt.Sprintf("http://%s", r.Host)
	_, err := restService.userService.Register(&userCredentials, serverURI)
	if err != nil {
		restService.jsonWriter.Error400(w, r, err)
		return
	}
	restService.jsonWriter.Write(w, r, http.StatusOK, RegistrationResponse{Success: true})
}

// Activates a pending registration
func (restService *RegistrationRestService) Activate(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	registrationID, err := strconv.ParseUint(params["token"], 10, 64)
	if err != nil {
		restService.jsonWriter.Error400(w, r, err)
		return
	}

	//userAccount, err := restService.userService.Activate(registrationID)
	_, err = restService.userService.Activate(registrationID)
	if err != nil {
		restService.jsonWriter.Error500(w, r, err)
		return
	}

	// Load the email activation HTML template
	tmpl := fmt.Sprintf("%s/%s", common.HTTP_PUBLIC_HTML, common.EMAIL_ACTIVATION)
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		restService.jsonWriter.Error500(w, r, err)
		return
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, nil); err != nil {
		restService.jsonWriter.Error500(w, r, err)
		return
	}

	// Write the rendered HTML file
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, buf.String())
}
