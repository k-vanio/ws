package ws

type Message struct {
	Action string      `json:"a"`
	Data   interface{} `json:"d"`
}

func NewMessage(action string, data interface{}) *Message {
	return &Message{
		Action: action,
		Data:   data,
	}
}
