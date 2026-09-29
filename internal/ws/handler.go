package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Local dev / same-origin dashboard: allow any origin. Tighten this once
	// the Flutter app's deployment origin is known.
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Handler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws: upgrade failed: %v", err)
			return
		}
		hub.ServeClient(conn)
	}
}
