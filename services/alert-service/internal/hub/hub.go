package hub

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/digital-twin/platform/services/alert-service/internal/store"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // non-browser clients
		}
		return strings.HasPrefix(origin, "http://"+r.Host) || strings.HasPrefix(origin, "https://"+r.Host)
	},
}

type Message struct {
	Type    string      `json:"type"`
	Payload store.Alert `json:"payload"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

func New() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]bool)}
}

func (h *Hub) Broadcast(msgType string, alert store.Alert) {
	data, err := json.Marshal(Message{Type: msgType, Payload: alert})
	if err != nil {
		slog.Error("marshal ws message", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			slog.Warn("ws write failed", "error", err)
		}
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, initial []store.Alert) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("ws upgrade failed", "error", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		_ = conn.Close()
	}()

	for _, alert := range initial {
		data, _ := json.Marshal(Message{Type: "alert.raised", Payload: alert})
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}
