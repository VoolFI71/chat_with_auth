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
    //"log"
    "chat/internal/handlers"
    //"chat/internal/db"
)

func main() {
    router := gin.Default()

    db, err := sql.Open("postgres", "postgresql://postgres:1234@localhost:5432/go?sslmode=disable") 
    if err != nil {
      panic(err)
    }
    defer db.Close()

    store, err := postgres.NewStore(db, []byte("secret"))
    if err != nil {
        panic(err)
    }    
    defer db.Close()

    router.Use(sessions.Sessions("mysession", store))


    router.GET(`/`, handlers.MainPage)
    router.POST("/register", handlers.Register)
    router.GET("/incr", func(c *gin.Context) {
        session := sessions.Default(c)
        var count int
        v := session.Get("count")
        if v == nil {
          count = 0
        } else {
          count = v.(int)
          count++
        }
        session.Set("count", count)
        session.Save()
        c.JSON(200, gin.H{"count": count})
      })
    if err := router.Run(":8080"); err != nil {
        panic(err)
    }
} 

// func register(c *gin.Context) {
//     var user User
//     if err := c.ShouldBindJSON(&user); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//         return
//     }

//     // Проверка, существует ли пользователь
//     var exists bool
//     err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", user.Username).Scan(&exists)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
//         return
//     }
//     if exists {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
//         return
//     }

//     // Сохранение пользователя
//     _, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not register user"})
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
// }