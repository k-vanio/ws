package ws_test

import (
	"encoding/json"
	"testing"

	"github.com/k-vanio/ws"
	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	action := "request"
	data := map[string]string{"id": "any"}

	message := ws.NewMessage(action, data)

	assert.Equal(t, action, message.Action)
	assert.Equal(t, data, message.Data)
}

func TestMessageToJson(t *testing.T) {
	message := ws.NewMessage("request", map[string]string{"name": "any"})
	expected := []byte(`{"a":"request","d":{"name":"any"}}`)

	result, err := json.Marshal(message)

	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
