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
			// Deny empty Origin: browsers always send Origin on WS upgrades.
			// Non-browser clients should use the REST API or same-host Origin.
			return false
		}
		return strings.HasPrefix(origin, "http://"+r.Host) || strings.HasPrefix(origin, "https://"+r.Host)
	},
}

type Message struct {
	Type    string      `json:"type"`
	Payload store.Alert `json:"payload"`
}

type client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*client]bool
}

func New() *Hub {
	return &Hub{clients: make(map[*client]bool)}
}

func (c *client) writeMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(messageType, data)
}

func (c *client) writeControl(messageType int, data []byte, deadline time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteControl(messageType, data, deadline)
}

func (h *Hub) Broadcast(msgType string, alert store.Alert) {
	data, err := json.Marshal(Message{Type: msgType, Payload: alert})
	if err != nil {
		slog.Error("marshal ws message", "error", err)
		return
	}

	h.mu.RLock()
	clients := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()
	for _, c := range clients {
		if err := c.writeMessage(websocket.TextMessage, data); err != nil {
			slog.Warn("ws write failed", "error", err)
			h.mu.Lock()
			delete(h.clients, c)
			h.mu.Unlock()
			_ = c.conn.Close()
		}
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, initial []store.Alert) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("ws upgrade failed", "error", err)
		return
	}

	c := &client{conn: conn}
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
		_ = conn.Close()
	}()

	for _, alert := range initial {
		data, err := json.Marshal(Message{Type: "alert.raised", Payload: alert})
		if err != nil {
			slog.Error("marshal ws snapshot", "error", err)
			continue
		}
		if err := c.writeMessage(websocket.TextMessage, data); err != nil {
			return
		}
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
			if err := c.writeControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}
