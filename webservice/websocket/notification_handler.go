package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	logging "github.com/op/go-logging"
)

type NotificationHandler struct {
	logger            *logging.Logger
	hub               *NotificationHub
	middlewareService service.Middleware
}

func NewNotificationHandler(logger *logging.Logger, hub *NotificationHub, middlewareService service.Middleware) *NotificationHandler {
	return &NotificationHandler{
		logger:            logger,
		hub:               hub,
		middlewareService: middlewareService}
}

func (ph *NotificationHandler) OnConnect(w http.ResponseWriter, r *http.Request) {
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
		ph.logger.Error("[NotificationHandler.OnConnect] Unable to establish webservice connection")
		return
	}

	var user model.User
	err = conn.ReadJSON(&user)
	if err != nil {
		ph.logger.Errorf("[NotificationHandler.OnConnect] webservice Read Error: %v", err)
		ph.logger.Errorf("[NotificationHandler.OnConnect] body=%s", r.Body)
		conn.Close()
		return
	}

	session, err := ph.middlewareService.CreateSession(w, r)
	if err != nil {
		ph.logger.Errorf("[NotificationHandler.OnConnect] Error: Unable to retrieve context from JsonWebTokenService: %s", err)
		return
	}
	session.SetLogger(ph.logger)

	ph.logger.Debug("[NotificationHandler.OnConnect] Accepting connection from ", conn.RemoteAddr())

	/*
		userDAO := dao.NewUserDAO(session)
		if err != nil {
			ph.logger.Errorf("[NotificationHandler.OnConnect] Error: %s", err.Error())
			return
		}

		userService := service.NewUserService(session, userDAO, mapper.NewUserMapper())
		notificationService := service.NewNotificationService(session)
	*/

	client := &NotificationClient{
		hub:     ph.hub,
		conn:    conn,
		send:    make(chan common.Notification, common.BUFFERED_CHANNEL_SIZE),
		session: session}

	client.hub.register <- client
	go client.writePump()
	go client.readPump()
	//go client.keepAlive()
}
