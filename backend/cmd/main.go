package main

import (
	//"time"
	//"fmt"
	//"encoding/json"
	//"database/sql"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"

	//"github.com/gorilla/websocket"
	"net/http"

	_ "github.com/jackc/pgx/v4/stdlib"

	//"github.com/gin-contrib/sessions/cookie"
	"log"
	//"net/http"
	"chat/internal/db"
	"chat/internal/handlers"
	"chat/internal/middleware"
	"chat/internal/websocket"
	"chat/internal/stream"

	"github.com/golang-jwt/jwt/v4"

	//"os"
	"github.com/joho/godotenv"
)


func main() {
    err := godotenv.Load()
    var jwtSecret = []byte("123")

    if err != nil {
        log.Fatalf("Ошибка загрузки .env файла: %v", err)
    }

    router := gin.Default()
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://glebase.ru"}, // Укажите адрес вашего фронтенда
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Разрешенные методы
        AllowHeaders:     []string{"Authorization", "Content-Type"}, // Разрешенные заголовки
        ExposeHeaders:    []string{"Content-Length"}, // Заголовки, которые могут быть доступны клиенту
        AllowCredentials: true, // Разрешить отправку учетных данных
    }))

    database, err := db.ConnectAuth()
    if (err!=nil){
        panic(err)
    }
    databasemsg, err := db.ConnectChat()
    if (err!=nil){
        panic(err)
    }

    sessionsOptions := sessions.Options{
        MaxAge:   1000,
        HttpOnly: true, 
    }


    store, err := postgres.NewStore(database, []byte("secret"))
    if err != nil {
        panic(err)
    }    
    defer database.Close()
    defer databasemsg.Close()

    
    router.Use(sessions.Sessions("mysession", store))
    router.Use(func(c *gin.Context) {
        session := sessions.Default(c)
        session.Options(sessionsOptions)
        c.Next()
    })

    //router.Use(cors.Default()) // Разрешает все источники

    go websocket.HandleMessages()

    router.GET("/gt", middleware.AuthMiddleware(), handlers.GT)
    router.GET(`/`, handlers.MainPage)
    router.GET("/wsstream", stream.Stream) // Теперь это работает

    router.GET("/ws", websocket.SendMsg(databasemsg))
    router.GET("/getmsg", websocket.GetMessagesHandler(databasemsg))
    router.POST("/savemsg",  middleware.AuthMiddleware(), websocket.SaveMsg(databasemsg)) //отправка сообщения

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
            return
        }

        username, ok := claims["username"].(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token"})
            return
        }

        // Возвращаем имя пользователя
        c.JSON(http.StatusOK, gin.H{
            "username": username,
        })
    })


    if err := router.Run(":8080"); err != nil {
        panic(err)
    }
} 

