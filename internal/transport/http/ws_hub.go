package http

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{clients: make(map[*websocket.Conn]bool)}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (hub *WebSocketHub) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	hub.mu.Lock()
	hub.clients[conn] = true
	hub.mu.Unlock()
	go func() {
		defer func() {
			hub.mu.Lock()
			delete(hub.clients, conn)
			hub.mu.Unlock()
			conn.Close()
		}()
		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	}()
}

func (hub *WebSocketHub) Broadcast(message []byte) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for conn := range hub.clients {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}
