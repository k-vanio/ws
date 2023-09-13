package ws

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
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
	stop       chan bool
}

func NewServer() *server {
	return &server{
		handler:    nil,
		mu:         &sync.Mutex{},
		broadcast:  make(chan []byte),
		clients:    make(map[Client]bool),
		register:   make(chan Client),
		unregister: make(chan Client),
		stop:       make(chan bool),
	}
}

func (s *server) Connect(fn func(client Client)) {
	s.handler = fn
}

func (s *server) Start(addr, pattern string) error {
	defer func() {
		close(s.broadcast)
		close(s.register)
		close(s.unregister)
		close(s.stop)
	}()

	if s.handler == nil {
		return ErrNotFoundHandler
	}

	go func() {
		r := mux.NewRouter()
		r.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			conn, err := Upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}

			client := NewClient(s, conn)

			s.Add(client)
			s.handler(client)
		})
		log.Fatalln(http.ListenAndServe(addr, r))
	}()

stop:
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
		case <-s.stop:
			break stop
		}
	}

	return nil
}

func (s *server) Stop() {
	s.mu.Lock()
	defer s.mu.Lock()
	for c := range s.clients {
		s.Remove(c)
	}

	s.stop <- true
}

func (s *server) Add(client Client) {
	s.register <- client
}

func (s *server) Remove(client Client) {
	s.unregister <- client
}

func (s *server) Size() int {
	defer s.mu.Unlock()
	s.mu.Lock()
	return len(s.clients)
}
