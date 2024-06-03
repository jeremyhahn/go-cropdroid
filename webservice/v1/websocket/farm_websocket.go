package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	logging "github.com/op/go-logging"
)

type FarmWebSocket struct {
	logger         *logging.Logger
	hub            *FarmHub
	middleware     middleware.JsonWebTokenMiddleware
	responseWriter response.HttpWriter
	WebSocket
}

func NewFarmWebSocket(
	logger *logging.Logger,
	hub *FarmHub,
	middleware middleware.JsonWebTokenMiddleware,
	responseWriter response.HttpWriter) *FarmWebSocket {

	return &FarmWebSocket{
		logger:         logger,
		hub:            hub,
		middleware:     middleware,
		responseWriter: responseWriter}
}

// Upgrades the HTTP connection to a websocket, parses the user from the initial
// websocket client message (required by this protocol), creates a new session with
// the parsed user information, and starts streaming messages.
func (farmHandler *FarmWebSocket) OnConnect(w http.ResponseWriter, r *http.Request) {
	//var user model.User
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		farmHandler.logger.Error(err.Error())
		farmHandler.responseWriter.Error500(w, r, err)
		return
	}
	if conn == nil {
		farmHandler.logger.Error("[FarmWebSocketUpgrader.OnConnect] Unable to establish webservice connection")
		return
	}
	// defer func() {
	// 	farmHandler.logger.Error("[FarmWebSocketUpgrader.OnConnect] closing connection")
	// 	conn.Close()
	// }()
	//err = conn.ReadJSON(&user)
	if err != nil {
		farmHandler.logger.Errorf("[FarmWebSocketUpgrader.OnConnect] webservice Read Error: %v", err)
		farmHandler.logger.Errorf("[FarmWebSocketUpgrader.OnConnect] body=%s", r.Body)
		farmHandler.responseWriter.Error400(w, r, err)
		conn.Close()
		return
	}

	session, err := farmHandler.middleware.CreateSession(w, r)
	if err != nil {
		farmHandler.logger.Errorf("[FarmWebSocketUpgrader.OnConnect] Error: Unable to create JsonWebTokenService session: %s", err)
		farmHandler.responseWriter.Error400(w, r, err)
		conn.Close()
		return
	}
	session.SetLogger(farmHandler.logger)

	farmHandler.logger.Debug("[FarmWebSocketUpgrader.OnConnect] Accepting connection from ", conn.RemoteAddr())

	client := &FarmClient{
		logger:           farmHandler.logger,
		hub:              farmHandler.hub,
		conn:             conn,
		send:             make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE),
		state:            make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		deviceState:      make(chan map[string]state.DeviceStateMap, common.BUFFERED_CHANNEL_SIZE),
		deviceStateDelta: make(chan map[string]state.DeviceStateDeltaMap, common.BUFFERED_CHANNEL_SIZE),
		user:             session.GetUser()}

	client.hub.register <- client
	go client.writePump()
	go client.readPump()
	//go client.keepAlive()
}
