package ws

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/tomtorh96/grocery-app/internal/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for now
	},
}

// ServeWS upgrades an HTTP connection to WebSocket and registers the client
func ServeWS(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		userID := auth.GetUserID(r)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "failed to upgrade connection", http.StatusInternalServerError)
			return
		}

		client := &Client{
			conn:   conn,
			send:   make(chan []byte, 256),
			listID: listID,
			userID: userID,
		}

		hub.Register(client)

		go client.writePump()
		go client.readPump(hub)
	}
}
