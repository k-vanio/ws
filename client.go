package ws

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = int64(512)
	newline        = []byte{'\n'}
	space          = []byte{' '}
)

type Client interface {
	Server() *server
	Conn() *websocket.Conn
	Send(data []byte)
	On(name string, action func(c Client, data interface{}))
}

type client struct {
	server  *server
	conn    *websocket.Conn
	send    chan []byte                                 // Buffered channel for outbound messages.
	actions map[string]func(c Client, data interface{}) // Map of registered actions.
}

func (c *client) Server() *server {
	return c.server
}

func (c *client) Conn() *websocket.Conn {
	return c.conn
}

func (c *client) Send(data []byte) {
	c.send <- data
}

func NewClient(server *server, conn *websocket.Conn) Client {
	newClient := &client{
		server:  server,
		conn:    conn,
		send:    make(chan []byte),
		actions: make(map[string]func(c Client, data interface{})),
	}

	go newClient.read()
	go newClient.writer()

	return newClient
}

func (c *client) read() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
		close(c.send)
	}()

	c.Conn().SetReadLimit(maxMessageSize)
	c.Conn().SetReadDeadline(time.Now().Add(pongWait))
	c.Conn().SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.Conn().ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		data := new(Message)
		err = json.Unmarshal(message, data)
		if err != nil {
			return
		}

		if action, ok := c.actions[data.Action]; ok {
			action(c, data.Data)
		}
	}
}

// writer handles outgoing messages to the WebSocket connection.
func (c *client) writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			c.Conn().SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn().WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case message, ok := <-c.send:
			c.Conn().SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn().WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn().NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

// On registers an action with a given name and corresponding function.
func (c *client) On(name string, action func(c Client, data interface{})) {
	c.actions[name] = action
}

// Emit sends a message to a specific client.
func (c *client) Emit(to Client, data *Message) {
	if _, ok := c.Server().clients[to]; ok {
		dataSend, err := json.Marshal(data)
		if err == nil {
			to.Send(dataSend)
		}
	}
}

// Broadcast sends a message to all connected clients except the sender.
func (c *client) Broadcast(data *Message) {
	dataSend, err := json.Marshal(data)
	if err == nil {
		for to := range c.Server().clients {
			if to != c {
				to.Send(dataSend)
			}
		}
	}
}
