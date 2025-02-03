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
        session.Options(sessionsOptions) // Установка параметров сессии
        c.Next()
    })
    // router.Use(middleware.CORSMiddleware())
    router.Use(cors.Default()) // Разрешает все источники

    router.GET("/incr", handlers.Incr)

    router.GET("/gt", middleware.AuthMiddleware(), handlers.GT)

    router.GET(`/`, handlers.MainPage)
    router.POST("/register", handlers.Register(database))
    router.POST("/login", handlers.Login(database))
    if err := router.Run(":8080"); err != nil {
        panic(err)
    }
} 

