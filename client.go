package ws

import (
	"github.com/gorilla/websocket"
)

type Client interface {
}

func NewClient(server *server, conn *websocket.Conn) Client {
	return &struct{}{}
}
