package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/jeremyhahn/cropdroid/state"
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
		hub:                  ph.hub,
		conn:                 conn,
		send:                 make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE),
		state:                make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		controllerState:      make(chan map[string]state.ControllerStateMap, common.BUFFERED_CHANNEL_SIZE),
		controllerStateDelta: make(chan map[string]state.ControllerStateDeltaMap, common.BUFFERED_CHANNEL_SIZE),
		session:              session}

	client.hub.register <- client
	go client.writePump()
	go client.readPump()
	//go client.keepAlive()
}
