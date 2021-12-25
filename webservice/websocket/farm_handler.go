package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	logging "github.com/op/go-logging"
)

type FarmHandler struct {
	logger            *logging.Logger
	hub               *FarmHub
	middlewareService service.Middleware
}

func NewFarmHandler(logger *logging.Logger, hub *FarmHub,
	notificationService service.NotificationService, middlewareService service.Middleware) *FarmHandler {

	return &FarmHandler{
		logger:            logger,
		hub:               hub,
		middlewareService: middlewareService}
}

func (ph *FarmHandler) OnConnect(w http.ResponseWriter, r *http.Request) {
	var user model.User
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ph.logger.Error(err.Error())
	}
	if conn == nil {
		ph.logger.Error("[FarmHandler.OnConnect] Unable to establish webservice connection")
		return
	}

	err = conn.ReadJSON(&user)
	if err != nil {
		ph.logger.Errorf("[FarmHandler.OnConnect] webservice Read Error: %v", err)
		ph.logger.Errorf("[FarmHandler.OnConnect] body=%s", r.Body)
		conn.Close()
		return
	}

	session, err := ph.middlewareService.CreateSession(w, r)
	if err != nil {
		ph.logger.Errorf("[FarmHandler.OnConnect] Error: Unable to create JsonWebTokenService session: %s", err)
		return
	}
	session.SetLogger(ph.logger)

	ph.logger.Debug("[FarmHandler.OnConnect] Accepting connection from ", conn.RemoteAddr())

	client := &FarmClient{
		logger:           ph.logger,
		hub:              ph.hub,
		conn:             conn,
		send:             make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE),
		state:            make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		deviceState:      make(chan map[string]state.DeviceStateMap, common.BUFFERED_CHANNEL_SIZE),
		deviceStateDelta: make(chan map[string]state.DeviceStateDeltaMap, common.BUFFERED_CHANNEL_SIZE),
		user:             session.GetUser()}

	client.hub.register <- client
	go client.writePump()
	go client.readPump()
	//go client.keepAlive()
}
