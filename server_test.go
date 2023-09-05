package ws_test

import (
	"testing"

	"github.com/k-vanio/ws"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	server := ws.NewServer()

	go func() {
		assert.Equal(t, 0, server.Size())
		server.Stop()
	}()

	server.Start(":7000", "/")
}
