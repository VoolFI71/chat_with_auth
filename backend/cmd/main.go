package main

import (
    //"time"
	//"fmt"
    //"encoding/json"
    //"database/sql"
    _ "github.com/jackc/pgx/v4/stdlib"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/postgres"
    //"github.com/gin-contrib/sessions/cookie"
    "log"
    //"net/http"
    //"github.com/golang-jwt/jwt/v4"
    "chat/internal/handlers"
    "chat/internal/db"
    "chat/internal/middleware"
    "github.com/gin-contrib/cors"
    //"os"
    "github.com/joho/godotenv"
)


func main() {
    err := godotenv.Load() // Путь к .env файлу

    if err != nil {
        log.Fatalf("Ошибка загрузки .env файла: %v", err)
    }

    router := gin.Default()

    database, err := db.Connect()

    if (err!=nil){
        panic(err)
    }

    sessionsOptions := sessions.Options{
        MaxAge:   4,
        HttpOnly: true, 
    }


    store, err := postgres.NewStore(database, []byte("secret"))
    if err != nil {
        panic(err)
    }    
    defer database.Close()

    
    router.Use(sessions.Sessions("mysession", store))
    router.Use(func(c *gin.Context) {
        session := sessions.Default(c)
        session.Options(sessionsOptions)
        c.Next()
    })
    // router.Use(middleware.CORSMiddleware())

    //router.Use(cors.Default()) // Разрешает все источники
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://127.0.0.1:5500"}, 
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Разрешенные методы
        AllowHeaders:     []string{"Authorization", "Content-Type"}, // Разрешенные заголовки
        ExposeHeaders:    []string{"Content-Length"}, // Заголовки, которые могут быть доступны клиенту
        AllowCredentials: true, // Разрешить отправку учетных данных
    }))

    router.GET("/gt", middleware.AuthMiddleware(), handlers.GT)
    router.GET(`/`, handlers.MainPage)
    router.POST("/sendmail", handlers.Sendmail(database))
    router.POST("/login", handlers.Login(database))
    router.POST("/reg", handlers.Reg(database))
    //router.POST("/sendmail", handlers.Sendmail)
    if err := router.Run(":8080"); err != nil {
        panic(err)
    }
} 

