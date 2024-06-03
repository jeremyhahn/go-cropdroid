package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type ChannelRestServicer interface {
	SetService(channelService service.ChannelServicer)
	SetMiddleware(middleware middleware.JsonWebTokenMiddleware)
	List(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type ChannelRestService struct {
	channelService service.ChannelServicer
	middleware     middleware.JsonWebTokenMiddleware
	httpWriter     response.HttpWriter
	ChannelRestServicer
}

func NewChannelRestService(
	channelService service.ChannelServicer,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) ChannelRestServicer {

	return &ChannelRestService{
		channelService: channelService,
		middleware:     middleware,
		httpWriter:     httpWriter}
}

// Dependency injection to set mocked channel service
func (restService *ChannelRestService) SetService(channelService service.ChannelServicer) {
	restService.channelService = channelService
}

// Dependency injection to set mocked JWT middleware
func (restService *ChannelRestService) SetMiddleware(middleware middleware.JsonWebTokenMiddleware) {
	restService.middleware = middleware
}

// Returns all channels for the requested deviceID HTTP GET parameter
func (restService *ChannelRestService) List(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	params := mux.Vars(r)
	sDeviceID := params["deviceID"]
	deviceID, err := strconv.ParseUint(sDeviceID, 0, 64)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	channels, err := restService.channelService.GetByDeviceID(session, deviceID)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, channels)
}

// Updates a device channel
func (restService *ChannelRestService) Update(w http.ResponseWriter, r *http.Request) {
	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	defer session.Close()
	var channel model.ChannelStruct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&channel); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err = restService.channelService.Update(session, &channel); err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, nil)
}
