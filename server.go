package ws

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	ErrNotFoundHandler = errors.New("not found handler")
)

var Upgrader = websocket.Upgrader{
	CheckOrigin:       func(r *http.Request) bool { return true },
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

type server struct {
	handler    func(client Client)
	mu         *sync.Mutex
	broadcast  chan []byte
	clients    map[Client]bool
	register   chan Client
	unregister chan Client
}

func NewServer() *server {
	s := &server{
		handler:    nil,
		mu:         &sync.Mutex{},
		broadcast:  make(chan []byte),
		clients:    make(map[Client]bool),
		register:   make(chan Client),
		unregister: make(chan Client),
	}

	go s.start()

	return s
}

func (s *server) Connect(fn func(client Client)) {
	s.handler = fn
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.handler == nil {
		http.Error(w, "unregistered handler", http.StatusInternalServerError)
		return
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	NewClient(s, conn)
	s.handler(NewClient(s, conn))
}

func (s *server) start() {
	defer func() {
		close(s.broadcast)
		close(s.register)
		close(s.unregister)
	}()

	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
		case client := <-s.unregister:
			s.mu.Lock()
			delete(s.clients, client)
			s.mu.Unlock()
		}
	}
}

func (s *server) Size() int {
	defer s.mu.Unlock()
	s.mu.Lock()
	return len(s.clients)
}
