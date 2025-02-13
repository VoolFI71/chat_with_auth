package stream

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "net/http"
    "sync"
	"fmt"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var clients = make(map[*websocket.Conn]bool) // Подключенные клиенты
var mu sync.Mutex // Мьютекс для безопасного доступа к clients

// HandleWebSocket принимает WebSocket соединение и обрабатывает сообщения
func HandleWebSocket(conn *websocket.Conn) {
    defer conn.Close()

    mu.Lock()
    clients[conn] = true
    mu.Unlock()
	fmt.Println("New client connected")

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
			fmt.Println("Error reading video:", err)

            mu.Lock()
            delete(clients, conn)
            mu.Unlock()
			fmt.Println("Client disconnected")

            break
        }

        // Рассылаем полученное сообщение всем подключенным клиентам
        mu.Lock()
        for client := range clients {
            if err := client.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				fmt.Println("Error sending video to client:", err)

                client.Close()
                delete(clients, client)
            }
        }
        mu.Unlock()
    }
}

func Stream(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.String(http.StatusInternalServerError, "Ошибка при подключении: %v", err)
        return
    }

    HandleWebSocket(conn)
}