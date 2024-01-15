package websocket

import (
	"github.com/jeremyhahn/go-cropdroid/service"
	logging "github.com/op/go-logging"
)

type FarmHub struct {
	logger  *logging.Logger
	clients map[*FarmClient]bool
	//broadcast           chan config.FarmConfig
	register            chan *FarmClient
	unregister          chan *FarmClient
	notificationService service.NotificationService
	farmService         service.FarmService
}

func NewFarmHub(logger *logging.Logger,
	notificationService service.NotificationService,
	farmService service.FarmService) *FarmHub {

	return &FarmHub{
		logger: logger,
		//broadcast:           make(chan config.FarmConfig),
		register:            make(chan *FarmClient),
		unregister:          make(chan *FarmClient),
		clients:             make(map[*FarmClient]bool),
		notificationService: notificationService,
		farmService:         farmService}
}

func (h *FarmHub) Run() {

	for {

		select {
		case client := <-h.register:
			h.logger.Debugf("[FarmHub.Run] Registering new client: address=%s, user=%s. %d clients connected to farm hub %d",
				client.conn.RemoteAddr(), client.getUser().GetEmail(), len(h.clients), h.farmService.GetFarmID())
			h.clients[client] = true
			//h.logger.Debugf("Sending config: %s", client.session.GetFarmService().GetConfig())
			client.send <- *h.farmService.GetConfig()

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.logger.Debugf("[FarmHub.Run] Unregistering client address=%s, user=%s",
					client.conn.RemoteAddr(), client.getUser().GetEmail())
				client.disconnect()
				delete(h.clients, client)
				close(client.send)
			}

		case farmConfig := <-h.farmService.WatchConfig():
			for client := range h.clients {
				select {
				case client.send <- farmConfig:
					h.logger.Errorf("[FarmHub.Run] Broadcasting configuration update for farm.id=%d, farm.name=%s\n",
						farmConfig.GetID(), farmConfig.GetName())
				default:
					h.logger.Errorf("[FarmHub.Run] Unable to send config update to client: %s", client.conn.RemoteAddr())
					close(client.send)
					delete(h.clients, client)
				}
			}

		case farmState := <-h.farmService.WatchState():
			for client := range h.clients {
				select {
				case client.state <- farmState:
					h.logger.Errorf("[FarmHub.Run] Broadcasting farm state update for farm.id=%d", h.farmService.GetFarmID())
					for k, v := range farmState.GetDevices() {
						h.logger.Errorf("[FarmHub.Run] farm.state.%s=%+v", k, v)
					}
				default:
					h.logger.Errorf("[FarmHub.Run] Unable to send state update to client: %s", client.conn.RemoteAddr())
					close(client.send)
					delete(h.clients, client)
				}
			}

		/*
			case deviceState := <-h.farmService.WatchDeviceState():
				for client := range h.clients {
					select {
					case client.deviceState <- deviceState:
						h.logger.Errorf("[FarmHub.Run] Broadcasting device state update for farm.id=%d", h.farmService.GetFarmID())
						h.logger.Errorf("[FarmHub.Run] deviceState=%+v", deviceState)
					default:
						h.logger.Errorf("[FarmHub.Run] Unable to send state update to client: %s", client.conn.RemoteAddr())
						close(client.send)
						delete(h.clients, client)
					}
				}*/

		case deviceStateDelta := <-h.farmService.WatchDeviceDeltas():
			for client := range h.clients {
				select {
				case client.deviceStateDelta <- deviceStateDelta:
					h.logger.Errorf("[FarmHub.Run] Broadcasting device state delta update for farm.id=%d", h.farmService.GetFarmID())
					for k, v := range deviceStateDelta {
						h.logger.Errorf("[FarmHub.Run] deviceStateDelta: k=%s, v=%+v", k, v)
					}
				default:
					h.logger.Errorf("[FarmHub.Run] Unable to send state update to client: %s", client.conn.RemoteAddr())
					close(client.send)
					delete(h.clients, client)
				}
			}
		}

	}
}
