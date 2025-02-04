package handlers

import (
	//"net/http"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
    "github.com/joho/godotenv"
    "os"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gopkg.in/gomail.v2"
)

var jwtSecret = []byte("123")

func MainPage(c *gin.Context) {
    response := map[int]int{5: 5}
        c.JSON(200, response)
} 

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Email string `json:"email"`
}

func Sendmailfunc(user *User) error {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    confirmationCode := 100000 + r.Intn(899999)
    codeStr := strconv.Itoa(confirmationCode)

    m := gomail.NewMessage()

    err := godotenv.Load()

    m.SetHeader("From", os.Getenv("MAILCODESEND"))
    m.SetHeader("To", user.Email)
    m.SetHeader("Subject", "Подтверждение регистрации")
    m.SetBody("text/plain", "Ваш код подтверждения: "+codeStr)
    if err != nil {
        log.Fatal("Ошибка загрузки .env файла")
        return err
    }
    d := gomail.NewDialer("smtp.mail.ru", 465, os.Getenv("MAILCODESEND"), os.Getenv("SMTPPASSOWRD"))
    if err := d.DialAndSend(m); err != nil {
        fmt.Println(err)

        return err
    }

    return nil // Успешная отправка
}

func Sendmail(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var user User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(400, gin.H{"error": "Invalid input"})
            return
        }

        var exists1 bool
        var exists2 bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM g WHERE username = $1)", user.Username).Scan(&exists1)

        if err != nil {
			log.Printf("Database error: %v", err)
			c.JSON(500, gin.H{"error": "Database error"})
			return
		}

        err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM g WHERE email = $1)", user.Email).Scan(&exists2)

        if err != nil {
			log.Printf("Database error: %v", err)
			c.JSON(500, gin.H{"error": "Database error"})
			return
		}

        if exists1 {
            c.JSON(400, gin.H{"error": "Данный юзернейм уже используется"})
            return
        }

        if exists2 {
            c.JSON(400, gin.H{"error": "Данная почта уже используется"})
            return
        }

		// _, err = db.Exec("INSERT INTO g (username, password, balance, email) VALUES ($1, $2, $3, $4)", user.Username, user.Password, 0, user.Email)
        // if err != nil {
        //     c.JSON(500, gin.H{"error": "Ошибка при создании пользователя"})
        //     return
        // }

        // c.JSON(200, gin.H{
        //     "message": "Можно создать пользователя с таким юзернеймом и почтой",
        //     "username": user.Username,
        //     "email": user.Email,
        // })
        err = Sendmailfunc(&user)
        if err!=nil{
            c.JSON(501, gin.H{
                "message": "Можно создать пользователя с таким юзернеймом и почтой",
                "username": user.Username,
                "email": user.Email,
                "error": "Ошибка при отправке кода на почту",
            })
            return
        }
        c.JSON(200, gin.H{
            "message": "Можно создать пользователя с таким юзернеймом и почтой",
            "username": user.Username,
            "email": user.Email,
            "status": "Код успешно отправлен на почту",
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
                c.JSON(401, gin.H{"error": "Неверно указан логин или пароль"})
                return
            }
            log.Printf("Database error: %v", err)
            c.JSON(500, gin.H{"error": "Database error"})
            return
        }

        if user.Password != storedPassword {
            c.JSON(401, gin.H{"error": "Неверный логин или пароль"})
            return
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "username": user.Username,
            "exp":      time.Now().Add(time.Hour * 72).Unix(), 
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

        c.SetCookie("token", tokenString, 3600*72, "/", "127.0.0.1", false, true) 
        c.JSON(200, gin.H{
            "message": "Login successful",
            "username": user.Username,
            "token": tokenString,
        })
    }
}


func GT(c *gin.Context) {
    c.JSON(200, gin.H{"number": 1})
}
