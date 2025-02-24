package websocket

import (
	"database/sql"
	"fmt"
	"io"

	//"log"
	"net/http"
	"strings"

	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	//"github.com/go-redis/redis/v8"
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
	Image     string `json:"image"`
	Audio     string `json:"audio"`
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
			
			go func(message ChatMessage) {
				broadcast <- message
			}(msg)
		}
	}
}


func SaveMsg(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var jwtSecret = []byte("123")

        tokenString := c.GetHeader("Authorization")
        if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        // Извлечение логина пользователя из токена
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }
		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token claims"})
			return
		}

		var messageRequest ChatMessage
        if err := c.BindJSON(&messageRequest); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
            return
        }

        message := messageRequest.Message
        go func() {
            _, err = db.Exec("INSERT INTO chat (chat_id, username, message) VALUES ($1, $2, $3)", 1, username, message)
            if err != nil {
                fmt.Println("Failed to save message:", err)
                return
            }
        }()

        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username":  username})
    }
}


func SaveImage(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var jwtSecret = []byte("123")

        tokenString := c.GetHeader("Authorization")
        if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        // Извлечение логина пользователя из токена
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }
		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token claims"})
			return
		}

        imageHeader, err := c.FormFile("image")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
            return
        }

        file, err := imageHeader.Open()
        if err != nil {
            fmt.Println("Error opening file:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
            return
        }
        defer file.Close()

        // Читаем содержимое файла
        // image, err := io.ReadAll(file)
        // if err != nil {
        //     fmt.Println("Error reading file:", err)
        //     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
        //     return
        // }

        minioClient, err := minio.New("minio:9000", &minio.Options{
            Creds:  credentials.NewStaticV4("123123123", "123123123", ""),
            Secure: false,
        })
        if err != nil {
            fmt.Println("Error creating MinIO client:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MinIO client"})
            return
        }
    
        // Имя бакета
        bucketName := "chat-files"
        ctx := context.Background()

        // Создаем бакет, если он не существует 
        exists, err := minioClient.BucketExists(ctx, bucketName)
        if err != nil {
            fmt.Println("Error checking bucket existence:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bucket existence"})
            return
        }
        if !exists {
            err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
            if err != nil {
                fmt.Println("Error creating bucket:", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bucket"})
                return
            }
        }
        imageUrl := uuid.New().String() // Имя загруженного файла

        // Загружаем изображение в MinIO
        _, err = minioClient.PutObject(ctx, bucketName, imageUrl, file, imageHeader.Size, minio.PutObjectOptions{})
        if err != nil {
            fmt.Println("Error uploading image:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
            return
        }

        fmt.Println(imageUrl)
        go func() {
            _, err = db.Exec("INSERT INTO chat (chat_id, username, image) VALUES ($1, $2, $3)", 1, username, imageUrl)
            if err != nil {
                fmt.Println("Failed to save message:", err)
                return
            }
        }()

        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username":  username})
    }
}

func SaveAudio(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var jwtSecret = []byte("123")

        tokenString := c.GetHeader("Authorization")
        if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        // Извлечение логина пользователя из токена
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }
		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token claims"})
			return
		}

		audioFile, err := c.FormFile("audio")
		if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "No audio file provided"})
            return
        }
		file, err := audioFile.Open()
		if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open audio file"})
            return
        }
        defer file.Close()


		audio, err := io.ReadAll(file)
		if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audio file"})
            return
        }

        //fmt.Println(audio)
		go func() {
            _, err = db.Exec("INSERT INTO chat (chat_id, username, audio_data) VALUES ($1, $2, $3)", 1, username, audio)
            if err != nil {
                fmt.Println("Failed to save audio:", err)
                return
            }
        }()

        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username":  username})
    }
}


func GetMessagesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		messages, err := GetLastMessages(db)
		if err != nil {
			fmt.Println("Error fetching messages:", err)

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch messages"})
			return
		}
		c.JSON(http.StatusOK, messages)
	}
}

func GetLastMessages(db *sql.DB) ([]ChatMessage, error) {
	rows, err := db.Query("SELECT username, message, created_at, image, audio_data FROM chat WHERE chat_id=1 ORDER BY created_at DESC LIMIT 75")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
        var msg ChatMessage
		var imageUrl sql.NullString // Используем sql.NullString для обработки NULL значений
		var audioData []byte
        var message sql.NullString // Используем sql.NullString для обработки NULL значений

        if err := rows.Scan(&msg.Username, &message, &msg.CreatedAt, &imageUrl, &audioData); err != nil {
			fmt.Println("Error scanning row:", err) // Логируем ошибку

            return nil, err
        }

        if message.Valid {
            msg.Message = message.String // Присваиваем строку, если значение не NULL
        } else {
            msg.Message = "" // Или присваиваем пустую строку, если значение NULL
        }
        fmt.Println("Attempting to get object with name:", imageUrl)

        
        if imageUrl.Valid {
			minioClient, err := minio.New("minio:9000", &minio.Options{
				Creds:  credentials.NewStaticV4("123123123", "123123123", ""),
				Secure: false,
			})
			if err != nil {
				fmt.Println("Error creating MinIO client:", err)
				return nil, err
			}
			bucketName := "chat-files"
			ctx := context.Background()

			// Логируем имя объекта
			//fmt.Println("Attempting to get object with name:", imageUrl.String)

			object, err := minioClient.GetObject(ctx, bucketName, imageUrl.String, minio.GetObjectOptions{})
			if err != nil {
				fmt.Println("Error getting object:", err)
				return nil, err
			}
			defer object.Close()

			imageData, err := io.ReadAll(object)
			if err != nil {
				fmt.Println("Error reading image data:", err)
				return nil, err
			}
            //fmt.Println(imageData)
            msg.Image = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageData)

		}
        
        if len(audioData) > 0 {
            msg.Audio = "data:audio/wav;base64," + base64.StdEncoding.EncodeToString(audioData)
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
		go func(message ChatMessage) {
			for client := range clients {
				err := client.WriteJSON(message)
				if err != nil {
					fmt.Println("Error while writing message:", err)
					client.Close()
					delete(clients, client)
				}
			}
		}(msg)
	}
}