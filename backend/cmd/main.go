package main

import (
	//"encoding/json"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"

	//"net/http"
	"chat/internal/handlers/db"
	"chat/internal/handlers/db/cassandra"

	"chat/internal/handlers"
	"chat/internal/middleware"
	"chat/internal/websocket"

	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, continuing with environment variables")
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "123"
	}
	var jwtSecret = []byte(secret)

	err := db.Connect() //postgres
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	database := db.GetDB()

	cassDB := cassandra.NewDB("cassandra", "chat")
	defer cassDB.Close()

	router := gin.Default()
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://127.0.0.1"
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},                                // адрес фронтенда
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Разрешенные методы
		AllowHeaders:     []string{"Authorization", "Content-Type"},           // Разрешенные заголовки
		ExposeHeaders:    []string{"Content-Length"},                          // Заголовки, которые могут быть доступны клиенту
		AllowCredentials: true,                                                // Разрешить отправку учетных данных
	}))

	sessionsOptions := sessions.Options{
		MaxAge:   1000,
		HttpOnly: true,
	}

	store, err := postgres.NewStore(db.GetDB(), []byte("secret"))
	if err != nil {
		log.Fatalf("Ошибка создания хранилища сессий: %v", err)
	}

	router.Use(sessions.Sessions("mysession", store))
	router.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		session.Options(sessionsOptions)
		c.Next()
	})

	// print hello world

	go websocket.HandleMessages()

	router.GET("/gt", middleware.AuthMiddleware(), handlers.GT)
	router.GET(`/`, handlers.MainPage)
	router.GET("/ws", websocket.SendMsg())

	router.GET("/getmsg", websocket.GetMessagesHandler(cassDB.Session))
	router.POST("/savemsg", middleware.AuthMiddleware(), websocket.SaveMsg(cassDB.Session))
	router.POST("/saveimage", middleware.AuthMiddleware(), websocket.SaveImage(cassDB.Session))
	router.POST("/saveaudio", middleware.AuthMiddleware(), websocket.SaveAudio(cassDB.Session))

	router.POST("/sendmail", handlers.Sendmail(database))
	router.POST("/login", handlers.Login(database))
	router.POST("/reg", handlers.Reg(database))
	router.GET("/userinfo", func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		// Удаляем "Bearer " из токена, если он присутствует
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
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

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"username": username,
		})
	})

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
