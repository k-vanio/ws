package ws_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/k-vanio/ws"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	server := ws.NewServer()
	wait := make(chan bool)

	go func() {
		time.Sleep(time.Millisecond * 2)
		assert.Equal(t, 0, server.Size())
		<-wait
		close(wait)
		server.Stop()
	}()

	go func() {
		time.Sleep(time.Millisecond * 3)
		client, _, err := websocket.DefaultDialer.Dial("ws://0.0.0.0:7000/", nil)
		assert.Nil(t, err)

		defer func() {
			client.Close()
			wait <- true
		}()
	}()

	server.Connect(func(client ws.Client) {
		assert.NotNil(t, client)
		client.Server().Remove(client)
	})
	server.Start(":7000", "/")
}

func TestNewServerErrNotFoundHandler(t *testing.T) {
	server := ws.NewServer()

	err := server.Start(":7000", "/")
	assert.Equal(t, ws.ErrNotFoundHandler, err)
}

func TestNewServerErrUpgrade(t *testing.T) {
	server := ws.NewServer()

	go func() {
		time.Sleep(time.Millisecond * 2)
		res, _ := http.DefaultClient.Get("http://0.0.0.0:7001/")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		server.Stop()
	}()

	server.Connect(func(client ws.Client) {
		assert.NotNil(t, client)
	})
	server.Start(":7001", "/")
}
