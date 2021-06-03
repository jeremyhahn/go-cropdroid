// +build cluster

package rest

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/service"
)

type RaftRestService interface {
	TransferLeader(w http.ResponseWriter, r *http.Request)
}

type RaftRestServiceImpl struct {
	app         *app.App
	userService service.UserService
	jsonWriter  common.HttpWriter
}

func NewRaftRestService(app *app.App, jsonWriter common.HttpWriter) RaftRestService {
	return &RaftRestServiceImpl{
		app:        app,
		jsonWriter: jsonWriter}
}

func (restService *RaftRestServiceImpl) TransferLeader(w http.ResponseWriter, r *http.Request) {

	restService.app.Logger.Debugf("[RaftRestService.TransferLeader]")

	params := mux.Vars(r)
	farmID, err := strconv.Atoi(params["clusterID"])
	if err != nil {
		restService.app.Logger.Error(err.Error())
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	/*
		var response RaftResponse
		var request RaftRequest
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
		_, err := restService.userService.Raft(userCredentials)
		if err != nil {
			response.Error = err.Error()
			response.Success = false
		} else {
			response.Success = true
		}*/

	restService.jsonWriter.Write(w, http.StatusOK, farmID)
}
