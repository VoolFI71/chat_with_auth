package handlers

import (
    //"net/http"
    "github.com/gin-contrib/sessions"
    "database/sql"
    "log"
    "github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var jwtSecret = []byte("123") // Замените на ваш секретный ключ

func MainPage(c *gin.Context) {
    response := map[int]int{5: 5}
        c.JSON(200, response)
} 

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func Register(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var user User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(400, gin.H{"error": "Invalid input"})
            return
        }

        // Проверка существования пользователя
        var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM g WHERE username = $1)", user.Username).Scan(&exists)
		
        if err != nil {
			log.Printf("Database error: %v", err) // Логируем ошибку
			c.JSON(500, gin.H{"error": "Database error"})
			return
		}

        if exists {
            c.JSON(400, gin.H{"error": "Username already exists"})
            return
        }

		_, err = db.Exec("INSERT INTO g (username, password, balance) VALUES ($1, $2, $3)", user.Username, user.Password, 0)
        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to register user"})
            return
        }

        c.JSON(200, gin.H{
            "message": "User registered successfully",
            "username": user.Username,
        })
    }
}



func Login(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var user User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(400, gin.H{"error": "Invalid input"})
            return
        }

        var storedPassword string
        err := db.QueryRow("SELECT password FROM g WHERE username = $1", user.Username).Scan(&storedPassword)

        if err != nil {
            if err == sql.ErrNoRows {
                c.JSON(401, gin.H{"error": "Invalid username or password"})
                return
            }
            log.Printf("Database error: %v", err)
            c.JSON(500, gin.H{"error": "Database error"})
            return
        }

        if user.Password != storedPassword {
            c.JSON(401, gin.H{"error": "Invalid username or password"})
            return
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "username": user.Username,
            "exp":      time.Now().Add(time.Hour * 72).Unix(), // Токен будет действителен 72 часа
        })

        tokenString, err := token.SignedString(jwtSecret)
        if err != nil {
            log.Printf("Error signing token: %v", err)
            c.JSON(500, gin.H{"error": "Could not create token"})
            return
        }

		session := sessions.Default(c)
        session.Set("token", tokenString)
        session.Save()

        c.JSON(200, gin.H{
            "message": "Login successful",
            "username": user.Username,
            "token": tokenString,
        })
    }
}


func Incr(c *gin.Context) {
    session := sessions.Default(c)
    var count int

    v := session.Get("count")
    if v == nil {
        count = 1
        session.Set("count", count) 
        log.Println("Initializing count to 1")
    } else {
        count = v.(int) + 1
        session.Set("count", count) 
        log.Printf("Incrementing count to %d", count)
    }
	session.Save()
    c.JSON(200, gin.H{"count": count})
}

func GT(c *gin.Context) {
    c.JSON(200, gin.H{"number": 1})
}