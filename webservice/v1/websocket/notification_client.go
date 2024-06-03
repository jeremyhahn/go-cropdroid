package websocket

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
)

type NotificationClient struct {
	session service.Session
	//service service.NotificationService
	hub  *NotificationHub
	conn *websocket.Conn
	send chan model.Notification
}

func (c *NotificationClient) disconnect() {
	c.conn.Close()
}

func (c *NotificationClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	//c.conn.SetReadDeadline(time.Now().Add(pongWait))
	//c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		c.session.GetLogger().Debug("[NotificationClient.readPump] pumping...")
		var notification model.Notification
		err := c.conn.ReadJSON(&notification)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.hub.broadcast <- notification
	}
}

func (c *NotificationClient) writePump() {
	//ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.conn.Close()
	}()
	for {
		c.session.GetLogger().Debug("[NotificationClient.writePump] pumping...")
		message, ok := <-c.send
		//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if !ok {
			// The hub closed the channel.
			c.session.GetLogger().Warning("[NotificationClient.writePump] hub closed the channel")
			c.hub.unregister <- c
			return
		}
		err := c.conn.WriteJSON(message)
		if err != nil {
			c.session.GetLogger().Errorf("[NotificationClient.writePump] Error: %s", err.Error())
			return
		}
		// Add queued messages
		n := len(c.send)
		for i := 0; i < n; i++ {
			c.conn.WriteJSON(<-c.send)
		}
		/*
			case notification := <-c.hub.notificationService.Dequeue():
				if err := c.conn.WriteJSON(notification); err != nil {
					c.session.GetLogger().Errorf("[NotificationClient.writePump] Error: %s", err.Error())
					return
				}
				c.session.GetLogger().Debugf("[NotificationClient.writePump] Sent message: %v+", notification)
				time.Sleep(time.Second * 2) // Android rate limits notifications, 1/second
		*/
		/*
			case <-ticker.C:
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
		*/
	}
}

/*
func (c *NotificationClient) keepAlive() {
	lastResponse := time.Now()
	c.conn.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})
	go func() {
		for {
			c.session.GetLogger().Debug("[NotificationClient.keepAlive] Sending keepalive")
			err := c.conn.WriteMessage(websocket.PingMessage, []byte("keepalive"))
			if err != nil {
				c.session.GetLogger().Debugf("[NotificationClient.keepAlive] Error: %s", err.Error())
				return
			}
			time.Sleep(common.WEBSOCKET_KEEPALIVE / 2)
			if time.Now().Sub(lastResponse) > common.WEBSOCKET_KEEPALIVE {
				c.session.GetLogger().Debug("[NotificationClient.keepAlive] Closing orphaned connection")
				c.conn.Close()
				return
			}
		}
	}()
}
*/
