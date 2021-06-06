package rest

import (
	"encoding/json"
	"net/http"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type RegisterResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRestService interface {
	Register(w http.ResponseWriter, r *http.Request)
}

type RegisterRestServiceImpl struct {
	app         *app.App
	userService service.UserService
	jsonWriter  common.HttpWriter
}

func NewRegisterRestService(app *app.App, userService service.UserService, jsonWriter common.HttpWriter) RegisterRestService {
	return &RegisterRestServiceImpl{
		app:         app,
		userService: userService,
		jsonWriter:  jsonWriter}
}

func (restService *RegisterRestServiceImpl) Register(w http.ResponseWriter, r *http.Request) {
	restService.app.Logger.Debugf("[RegisterRestService.Register]")
	var response RegisterResponse
	var request RegisterRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		restService.jsonWriter.Write(w, http.StatusBadRequest, JsonResponse{
			Success: false,
			Error:   err.Error()})
		return
	}
	userCredentials := &service.UserCredentials{
		Email:    request.Username,
		Password: request.Password}
	_, err := restService.userService.Register(userCredentials)
	if err != nil {
		response.Error = err.Error()
		response.Success = false
	} else {
		response.Success = true
	}
	restService.jsonWriter.Write(w, http.StatusOK, response)
}
