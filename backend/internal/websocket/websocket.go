package websocket

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"database/sql"
    "github.com/golang-jwt/jwt/v4"
    "strings" 

)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


type ChatMessage struct {
	Username  string `json:"username"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"` 
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan ChatMessage)

func SendMsg(db *sql.DB)  gin.HandlerFunc { // функция для вебсокета
	return func (c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Error while upgrading connection:", err)
			return
		}
		defer conn.Close()

		clients[conn] = true

		for {
			var msg ChatMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				fmt.Println("Error while reading message:", err)
				delete(clients, conn)
				break
			}
			
			broadcast <- msg
		}
	}
}

func SaveMsg(db *sql.DB) gin.HandlerFunc{ // функция для сохранения сообщения в бд. Если успешно сохранено. То можно  выполнять функцию для вебсокета SendMsg
	return func (c *gin.Context) {
		
		var jwtSecret = []byte("123")

		tokenString := c.GetHeader("Authorization")

		if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return
        }

        // Парсим и проверяем токен
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Проверяем метод подписи
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        // Извлекаем логин пользователя из токена
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			fmt.Println(err)

            return
        }

        username, ok := claims["username"].(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token"})
			fmt.Println(err)

            return
        }


		var user ChatMessage
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(400, gin.H{"error": "Invalid input"})
			fmt.Println(err)
            return
        }

		_, err = db.Exec("INSERT INTO chat (username, message) VALUES ($1, $2)", username, user.Message)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
			fmt.Println(err)

            return
        }

        // Возвращаем успешный ответ
        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username": username})
	}
}

func GetMessagesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		messages, err := GetLastMessages(db)
		if err != nil {
			c.JSON(501, gin.H{"error": "Unable to fetch messages"})
			return
		}
		c.JSON(200, messages)
	}
}

func GetLastMessages(db *sql.DB) ([]ChatMessage, error) {
	rows, err := db.Query("SELECT username, message, created_at FROM chat ORDER BY created_at DESC LIMIT 10")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
        if err := rows.Scan(&msg.Username, &msg.Message, &msg.CreatedAt); err != nil {
			fmt.Println(err)
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		fmt.Println(err)	
		return nil, err
	}
	return messages, nil
}





func HandleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println("Error while writing message:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}