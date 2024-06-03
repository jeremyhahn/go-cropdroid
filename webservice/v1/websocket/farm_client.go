package websocket

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/state"
	logging "github.com/op/go-logging"
)

type FarmClient struct {
	logger           *logging.Logger
	hub              *FarmHub
	conn             *websocket.Conn
	send             chan config.Farm
	state            chan state.FarmStateMap
	deviceState      chan map[string]state.DeviceStateMap
	deviceStateDelta chan map[string]state.DeviceStateDeltaMap
	user             model.User
}

func (c *FarmClient) getUser() model.User {
	return c.user
}

func (c *FarmClient) disconnect() {
	c.conn.Close()
}

func (c *FarmClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	//c.conn.SetReadDeadline(time.Now().Add(pongWait))
	//c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		c.logger.Debug("[FarmClient.readPump] pumping...")
		var configuration config.Farm
		err := c.conn.ReadJSON(&configuration)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			c.logger.Errorf("[FarmClient.readPump] Error: %s", err.Error())
			break
		}
		//c.hub.broadcast <- configuration
	}
}

func (c *FarmClient) writePump() {
	//ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.logger.Error("[FarmClient.writePump] Closing connection")
		c.conn.Close()
	}()
	for {
		c.logger.Debug("[FarmClient.writePump] pumping...")
		select {
		case message, ok := <-c.send:
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.logger.Warning("[FarmClient.writePump] hub closed the channel")
				c.hub.unregister <- c
				return
			}

			err := c.conn.WriteJSON(message)
			if err != nil {
				c.logger.Errorf("[FarmClient.writePump] Error: %s", err.Error())
				return
			}
			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				c.conn.WriteJSON(<-c.send)
			}
			/*
				case <-ticker.C:
					c.conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
			*/

		case message, ok := <-c.state:
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.logger.Warning("[FarmClient.writePump] hub closed the channel")
				c.hub.unregister <- c
				return
			}
			err := c.conn.WriteJSON(message)
			if err != nil {
				c.logger.Errorf("[FarmClient.writePump] Error: %s", err.Error())
				return
			}
			// Add queued messages
			n := len(c.state)
			for i := 0; i < n; i++ {
				c.conn.WriteJSON(<-c.state)
			}

		case message, ok := <-c.deviceState:
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.logger.Warning("[FarmClient.writePump] hub closed the channel")
				c.hub.unregister <- c
				return
			}
			err := c.conn.WriteJSON(message)
			if err != nil {
				c.logger.Errorf("[FarmClient.writePump] Error: %s", err.Error())
				return
			}
			// Add queued messages
			n := len(c.deviceState)
			for i := 0; i < n; i++ {
				c.conn.WriteJSON(<-c.deviceState)
			}

		case message, ok := <-c.deviceStateDelta:
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.logger.Warning("[FarmClient.writePump] hub closed the channel")
				c.hub.unregister <- c
				return
			}
			err := c.conn.WriteJSON(message)
			if err != nil {
				c.logger.Errorf("[FarmClient.writePump] Error: %s", err.Error())
				return
			}
			// Add queued messages
			n := len(c.deviceStateDelta)
			for i := 0; i < n; i++ {
				c.conn.WriteJSON(<-c.deviceStateDelta)
			}
		}

	}
}
