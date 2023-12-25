package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Yatsok/digital-queue/internal/helper"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := helper.GetUserIDContext(r)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	s.currentUserMap[userID] = conn
	defer func() {
		s.connectedClientsMutex.Lock()
		delete(s.connectedClients, conn)
		delete(s.currentUserMap, userID)
		s.connectedClientsMutex.Unlock()
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket closed unexpectedly: %v", err)
			} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("WebSocket closed by client: %v", err)
			}
			return
		}

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println("Error writing message:", err)
			return
		}
	}
}

func (s *Server) StartWebSocketServer() {
	http.HandleFunc("/ws", s.handleWebSocket)
	s.WebSocketServer = &http.Server{
		Addr:    ":8081",
		Handler: nil,
	}

	go func() {
		fmt.Println("WebSocket server listening on :8081...")
		err := s.WebSocketServer.ListenAndServe()
		if err != nil {
			fmt.Println("WebSocket server error:", err)
		}
	}()
}
