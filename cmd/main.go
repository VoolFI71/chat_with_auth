package main

import (
    //"time"
	//"fmt"
    //"encoding/json"
    //"database/sql"
    _ "github.com/jackc/pgx/v4/stdlib"
    //"github.com/golang-jwt/jwt/v4"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/postgres"
    //"github.com/gin-contrib/sessions/cookie"
    //"log"
    "chat/internal/handlers"
    "chat/internal/db"
)

func main() {
    router := gin.Default()

    database, err := db.Connect()
    if (err!=nil){
        panic(err)
    }
    sessionsOptions := sessions.Options{
        MaxAge:   4, // Время жизни сессии в секундах (например, 1 час)
        HttpOnly: true, // Запрет на доступ к cookie через JavaScript
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

    router.GET("/incr", handlers.Incr)

    router.GET(`/`, handlers.MainPage)
    router.POST("/register", handlers.Register(database))

    if err := router.Run(":8080"); err != nil {
        panic(err)
    }
} 

