package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	upgrader  websocket.Upgrader
	clients   map[int]*websocket.Conn
	clientsMu sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // В production нужно реализовать проверку origin
			},
		},
		clients: make(map[int]*websocket.Conn),
	}
}

func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request, userID int) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}

	// Регистрируем соединение
	s.clientsMu.Lock()
	s.clients[userID] = conn
	s.clientsMu.Unlock()

	log.Printf("Connected to user %d", userID)

	// Обработка закрытия соединения
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				s.clientsMu.Lock()
				delete(s.clients, userID)
				s.clientsMu.Unlock()
				conn.Close()
				break
			}
		}
	}()
}

func (s *Server) BroadcastToUser(userID int, message interface{}) error {
	s.clientsMu.RLock()
	conn, ok := s.clients[userID]
	s.clientsMu.RUnlock()

	if !ok {
		return nil // Пользователь не подключен
	}

	return conn.WriteJSON(message)
}
