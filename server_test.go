package ws_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k-vanio/ws"
	"github.com/stretchr/testify/assert"
)

func TestNewServerUnregisteredHandler(t *testing.T) {
	server := ws.NewServer()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "unregistered handler\n", w.Body.String())
}
