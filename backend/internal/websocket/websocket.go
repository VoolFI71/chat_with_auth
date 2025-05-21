package websocket

import (
	//"database/sql"
	"fmt"
	"io"
	"strconv"

	//"strconv"

	//"strconv"

	//"log"
	"net/http"
	"strings"

	"encoding/base64"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	//"github.com/go-redis/redis/v8"
	"context"
	"log"

	"github.com/gocql/gocql"
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
    CreatedAt int64 `json:"created_at"` // тут должно храниться время как int64 но в вид Unix time
	Image     string `json:"image"`
	Audio     string `json:"audio"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan ChatMessage)



func SendMsg()  gin.HandlerFunc { // функция для вебсокета
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


func SaveMsg(session *gocql.Session) gin.HandlerFunc {
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
            if session == nil {
                log.Fatalf("Сессия не инициализирована")
            }
            query := session.Query("INSERT INTO messages (chat_id, username, message, created_at) VALUES (?, ?, ?, ?)", 1, username, message,  time.Now().Unix())
            fmt.Println("130", time.Now().Unix())
            fmt.Println("133", message)

            if err := query.Exec(); err != nil {
                log.Fatalf("Ошибка при добавлении сообщения в базу данных: %v", err)
            }
        }()
        fmt.Println("Сообщение добавлено в базу данных")
        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username":  username})
    }
}


func SaveImage(session *gocql.Session) gin.HandlerFunc {
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
        imageUrl := uuid.New().String() 

        _, err = minioClient.PutObject(ctx, bucketName, imageUrl, file, imageHeader.Size, minio.PutObjectOptions{})
        if err != nil {
            fmt.Println("Error uploading image:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
            return
        }


        fmt.Println(imageUrl)
        go func() {
            query := session.Query("INSERT INTO messages (chat_id, username, message, image, created_at, audio_data) VALUES (?, ?, ?, ?, ?, ?)", 1, username, "", imageUrl, time.Now().Unix(), nil)
            if err := query.Exec(); err != nil {
                log.Fatalf("Ошибка при добавлении изображения в базу данных %v", err)
            }
        }()

        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username":  username})
    }
}

func SaveAudio(session *gocql.Session) gin.HandlerFunc {
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
        audioUrl := uuid.New().String() // Имя загруженного файла

        // Загружаем изображение в MinIO
        _, err = minioClient.PutObject(ctx, bucketName, audioUrl, file, audioFile.Size, minio.PutObjectOptions{})
        if err != nil {
            fmt.Println("Error uploading image:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
            return
        }

        //fmt.Println(audio)
		go func() {
            query := session.Query("INSERT INTO messages (chat_id, username, message, image, created_at, audio_data) VALUES (?, ?, ?, ?, ?, ?)", 1, username, "", "", time.Now().Unix(), audioUrl)
            err := query.Exec()
            if err != nil {
                fmt.Println(err)
            }
        }()

        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username":  username})
    }
}


func GetMessagesHandler(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
        var lastID int64
        var err error
        ID := c.Query("id")
        if (ID == "0") {
            lastID = time.Now().Unix()
        } else {
            lastID, err = strconv.ParseInt(ID, 10, 64)
            if err != nil {
                fmt.Println("преобразование не удалось")
            }
        }

        fmt.Println("lastID:", lastID)

        messages, newLastID, err := GetLastMessages(session, lastID)
		if err != nil {
			fmt.Println("Error fetching messages:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch messages"})
			return
		}
        fmt.Println(999)

        fmt.Println(newLastID)
        c.JSON(http.StatusOK, gin.H{
            "messages":     messages,
            "id": newLastID,
        })	
    }
}

func GetLastMessages(session *gocql.Session, lastID int64) ([]ChatMessage,  int64, error) {
	var newLastID int64
    fmt.Println(lastID, "392")
	query := "SELECT username, message, created_at, image, audio_data FROM messages WHERE chat_id = 1 LIMIT 5"
    //query := "SELECT username, message, created_at, image, audio_data FROM messages WHERE chat_id = 1 AND created_at < ? LIMIT 5"

    iter := session.Query(query).Iter()
    defer iter.Close()

	var messages []ChatMessage

    minioClient, err := minio.New("minio:9000", &minio.Options{
        Creds:  credentials.NewStaticV4("123123123", "123123123", ""),
        Secure: false,
    })
    if err != nil {
        fmt.Println("Error creating MinIO client:", err)
        return nil, 0, err
    }
    for {
        var msg ChatMessage
		var createdAt int64
        var imageUrl string
        var audioUrl string
        var message string
        fmt.Println()
        if !iter.Scan(&msg.Username, &msg.Message, &msg.CreatedAt, &imageUrl, &audioUrl) {
            // Проверяем, есть ли ошибка
            if err := iter.Close(); err != nil {
                fmt.Println("Error closing iterator:", err)
                return nil, 0, fmt.Errorf("error scanning row: %v", err)

            }
            fmt.Println(message, "<- сообщение")
            fmt.Println(createdAt, "<- сообщение")

            break
        }
        fmt.Println(createdAt, "created_at413")
        msg.Message = message // Присваиваем строку напрямую
        fmt.Println(message, "<- сообщение")

        if imageUrl != ""  {
            ctx := context.Background()
            object, err := minioClient.GetObject(ctx, "chat-files", imageUrl, minio.GetObjectOptions{})
            if err != nil {
                fmt.Println("Error getting object:", err)
                return nil, 0, err

            }
            defer object.Close()

            imageData, err := io.ReadAll(object)
            if err != nil {
                fmt.Println("Error reading image data:", err)
                return nil, 0, err

            }
            msg.Image = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageData)
        }

        if audioUrl != "" {
            ctx := context.Background()
            object, err := minioClient.GetObject(ctx, "chat-files", audioUrl, minio.GetObjectOptions{})
            if err != nil {
                fmt.Println("Error getting object:", err)
                return nil, 0, err
            }
            defer object.Close()
            
            audioData, err := io.ReadAll(object)
            if err != nil {
                fmt.Println("Error reading image data:", err)
                return nil, 0, err
            }
            
            msg.Audio = "data:audio/wav;base64," + base64.StdEncoding.EncodeToString(audioData)
        }
		newLastID = createdAt
        fmt.Println(newLastID, 453)
        messages = append(messages, msg)
    }

    fmt.Println(newLastID , "created_at3")

    return messages, newLastID, nil
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