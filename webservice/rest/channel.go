package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type ChannelRestService interface {
	SetChannel(w http.ResponseWriter, r *http.Request)
	RestService
}

type DefaultChannelRestService struct {
	channelService service.ChannelService
	channelMapper  mapper.ChannelMapper
	middleware     service.Middleware
	jsonWriter     common.HttpWriter
	ChannelRestService
}

func NewChannelRestService(channelService service.ChannelService, mapper mapper.ChannelMapper,
	middleware service.Middleware, jsonWriter common.HttpWriter) ChannelRestService {

	return &DefaultChannelRestService{
		channelService: channelService,
		middleware:     middleware,
		jsonWriter:     jsonWriter}
}

func (restService *DefaultChannelRestService) RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string {
	channelsEndpoint := fmt.Sprintf("%s/channels", baseFarmURI)
	router.Handle(channelsEndpoint, negroni.New(
		negroni.HandlerFunc(restService.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(restService.SetChannel)),
	)).Methods("PUT")
	return []string{channelsEndpoint}
}

func (restService *DefaultChannelRestService) SetChannel(w http.ResponseWriter, r *http.Request) {

	session, err := restService.middleware.CreateSession(w, r)
	if err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}
	defer session.Close()

	session.GetLogger().Debug("[ChannelRestService.SetChannel] Decoding JSON request")

	var channel model.Channel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&channel); err != nil {
		BadRequestError(w, r, err, restService.jsonWriter)
		return
	}

	session.GetLogger().Debugf("[ChannelRestService.SetChannel] channel=%+v", channel)

	if err = restService.channelService.Update(session, &channel); err != nil {
		session.GetLogger().Errorf("[ChannelRestService.Set] Error: ", err)
		restService.jsonWriter.Error200(w, err)
		return
	}

	restService.jsonWriter.Write(w, http.StatusOK, nil)
}
