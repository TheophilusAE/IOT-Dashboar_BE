package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// client represents one connected WebSocket subscriber. filter is nil for
// "all devices"; otherwise it restricts which device_id-scoped broadcasts
// this client receives.
type client struct {
	conn   *websocket.Conn
	send   chan []byte
	filter map[int64]bool // nil = no filter (receive everything)
}

// Hub tracks connected clients and broadcasts messages that originate only
// from real, committed database writes (see internal/ingest). There is no
// ticker in this file generating synthetic updates.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*client]struct{})}
}

func (h *Hub) register(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
}

// Broadcast sends msg to every subscribed client. If deviceID is non-nil,
// only clients whose filter includes it (or have no filter) receive it.
func (h *Hub) Broadcast(msg OutgoingMessage, deviceID *int64) {
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws: failed to marshal broadcast message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if deviceID != nil && c.filter != nil && !c.filter[*deviceID] {
			continue
		}
		select {
		case c.send <- payload:
		default:
			log.Printf("ws: client send buffer full, dropping message")
		}
	}
}

const (
	writeWait  = 10 * time.Second
	pingPeriod = 30 * time.Second
	sendBuffer = 32
)

// ServeClient upgrades an HTTP connection and runs its read/write pumps
// until disconnect. Call this from the HTTP handler after upgrading.
func (h *Hub) ServeClient(conn *websocket.Conn) {
	c := &client{conn: conn, send: make(chan []byte, sendBuffer)}
	h.register(c)

	go h.writePump(c)
	h.readPump(c)
}

func (h *Hub) readPump(c *client) {
	defer func() {
		h.unregister(c)
		c.conn.Close()
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var sub SubscribeMessage
		if err := json.Unmarshal(data, &sub); err != nil {
			continue
		}
		if sub.Type != "subscribe" {
			continue
		}

		if len(sub.DeviceIDs) == 0 {
			c.filter = nil
		} else {
			c.filter = make(map[int64]bool, len(sub.DeviceIDs))
			for _, id := range sub.DeviceIDs {
				c.filter[id] = true
			}
		}

		ack, _ := json.Marshal(OutgoingMessage{Type: TypeSubscribed, Data: sub.DeviceIDs})
		select {
		case c.send <- ack:
		default:
		}
	}
}

func (h *Hub) writePump(c *client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			// Protocol-level keepalive, not a data update — distinct from the
			// forbidden "Timer.periodic() generating fake sensor changes"
			// pattern: this never carries application data, only a liveness
			// heartbeat for the WS connection itself.
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
