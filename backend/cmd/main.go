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
    //"log"
    //"net/http"
    //"github.com/golang-jwt/jwt/v4"
    "chat/internal/handlers"
    "chat/internal/db"
    "chat/internal/middleware"
    "github.com/gin-contrib/cors"

)


func main() {
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
    go router.GET("/incr", handlers.Incr)
    go router.GET("/gt", middleware.AuthMiddleware(), handlers.GT)

    go router.GET(`/`, handlers.MainPage)
    go router.POST("/register", handlers.Register(database))
    go router.POST("/login", handlers.Login(database))
    if err := router.Run(":8080"); err != nil {
        panic(err)
    }
} 

