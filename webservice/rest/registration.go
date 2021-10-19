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
)

type RegistrationResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type RegistrationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegistrationRestService interface {
	Register(w http.ResponseWriter, r *http.Request)
	Activate(w http.ResponseWriter, r *http.Request)
}

type RegistrationRestServiceImpl struct {
	app         *app.App
	userService service.UserService
	jsonWriter  common.HttpWriter
}

func NewRegistrationRestService(app *app.App, userService service.UserService, jsonWriter common.HttpWriter) RegistrationRestService {
	return &RegistrationRestServiceImpl{
		app:         app,
		userService: userService,
		jsonWriter:  jsonWriter}
}

func (restService *RegistrationRestServiceImpl) Register(w http.ResponseWriter, r *http.Request) {
	restService.app.Logger.Debugf("[RegistrationRestService.Register]")
	var response RegistrationResponse
	var userCredentials service.UserCredentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userCredentials); err != nil {
		restService.jsonWriter.Write(w, http.StatusBadRequest, JsonResponse{
			Success: false,
			Error:   err.Error()})
		return
	}

	baseURI := fmt.Sprintf("http://%s", r.Host)
	restService.app.Logger.Debugf("Activate: %s", baseURI)

	_, err := restService.userService.Register(&userCredentials, baseURI)
	if err != nil {
		response.Error = err.Error()
		response.Success = false
	} else {
		response.Success = true
	}
	restService.jsonWriter.Write(w, http.StatusOK, response)
}

func (restService *RegistrationRestServiceImpl) Activate(w http.ResponseWriter, r *http.Request) {
	restService.app.Logger.Debugf("[RegistrationRestService.Activate]")

	params := mux.Vars(r)
	registrationID, err := strconv.ParseUint(params["token"], 10, 64)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	//userAccount, err := restService.userService.Activate(registrationID)
	_, err = restService.userService.Activate(registrationID)
	if err != nil {
		restService.jsonWriter.Error500(w, err)
		return
	}

	//restService.jsonWriter.Write(w, http.StatusOK, userAccount)

	tmpl := fmt.Sprintf("%s/%s", common.HTTP_PUBLIC_HTML, common.EMAIL_ACTIVATION)
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		restService.jsonWriter.Write(w, http.StatusBadRequest, err)
		return
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, nil); err != nil {
		restService.jsonWriter.Write(w, http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, buf.String())
}
