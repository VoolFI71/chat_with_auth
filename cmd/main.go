package main

import (
    //"time"
	//"fmt"
    //"encoding/json"
    "database/sql"
    _ "github.com/jackc/pgx/v4/stdlib"
    //"github.com/golang-jwt/jwt/v4"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/postgres"
    //"github.com/gin-contrib/sessions/cookie"
    "log"
    "chat/internal/handlers"
    //"chat/internal/db"
)

var db *sql.DB

func main() {
    router := gin.Default()

    db, err := sql.Open("postgres", "postgresql://postgres:1234@localhost:5432/go?sslmode=disable") 
    if err != nil {
      panic(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal("Ошибка при проверке соединения:", err)
    }

    // Создание таблицы (если она не существует)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS g (
        username VARCHAR(50) UNIQUE,
        password VARCHAR(100),
        balance DECIMAL(10, 2)
    )`)
    if err != nil {
        log.Fatal("Ошибка при создании таблицы:", err)
    }

    store, err := postgres.NewStore(db, []byte("secret"))
    if err != nil {
        panic(err)
    }    
    defer db.Close()

    
    router.Use(sessions.Sessions("mysession", store))


    router.GET(`/`, handlers.MainPage)
    router.POST("/register", handlers.Register(db))
    router.GET("/incr", handlers.Incr)
    if err := router.Run(":8080"); err != nil {
        panic(err)
    }
} 

