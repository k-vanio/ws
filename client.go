package ws

import (
	"github.com/gorilla/websocket"
)

type Client interface {
	Server() *server
	Conn() *websocket.Conn
}

type client struct {
	server *server
	conn   *websocket.Conn
}

func (c *client) Server() *server {
	return c.server
}

func (c *client) Conn() *websocket.Conn {
	return c.conn
}

func NewClient(server *server, conn *websocket.Conn) Client {
	return &client{
		server: server,
		conn:   conn,
	}
}
