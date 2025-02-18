package stream

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "net/http"
    "sync"
	//"fmt"
)
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true 
    },
}

var clients = make(map[*websocket.Conn]bool)
var mu sync.Mutex

func Stream(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not upgrade connection")
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}

		// Отправляем сообщение всем подключенным клиентам
		mu.Lock()
		for client := range clients {
			if err := client.WriteJSON(msg); err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}