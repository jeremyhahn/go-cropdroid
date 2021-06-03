package websocket

import (
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/service"
	logging "github.com/op/go-logging"
)

type NotificationHub struct {
	logger              *logging.Logger
	clients             map[*NotificationClient]bool
	broadcast           chan common.Notification
	register            chan *NotificationClient
	unregister          chan *NotificationClient
	notificationService service.NotificationService
}

func NewNotificationHub(logger *logging.Logger, notificationService service.NotificationService) *NotificationHub {
	return &NotificationHub{
		broadcast:           make(chan common.Notification),
		register:            make(chan *NotificationClient),
		unregister:          make(chan *NotificationClient),
		clients:             make(map[*NotificationClient]bool),
		logger:              logger,
		notificationService: notificationService}
}

func (h *NotificationHub) Run() {
	for {

		h.logger.Debugf("Notification hub running... %d clients connected. %d items in the queue.",
			len(h.clients), h.notificationService.QueueSize())

		select {
		case client := <-h.register:
			client.session.GetLogger().Debugf("[NotificationHub.run] Registering new client: address=%s, user=%s",
				client.conn.RemoteAddr(), client.session.GetUser().GetEmail())
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.logger.Debugf("[NotificationHub.run] Unregistering client address=%s, user=%s",
					client.conn.RemoteAddr(), client.session.GetUser().GetEmail())
				client.disconnect()
				delete(h.clients, client)
				close(client.send)
			}

		case notification := <-h.notificationService.Dequeue():
			h.doBroadcast(notification)

		case notification := <-h.broadcast:
			h.doBroadcast(notification)
		}

	}
}

func (h *NotificationHub) doBroadcast(notification common.Notification) {
	for client := range h.clients {
		h.logger.Debugf("[NotificationHub.doBroadcast] Notification: %+v\n", notification)
		select {
		case client.send <- notification:
		default:
			h.logger.Errorf("[NotificationHub.doBroadcast] Unable to send notification to client: ", client.conn.RemoteAddr)
			close(client.send)
			delete(h.clients, client)
		}
	}
}
