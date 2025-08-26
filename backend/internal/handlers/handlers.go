package handlers

import (
	//"net/http"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	//"github.com/gin-contrib/sessions"
	//"github.com/gin-contrib/sessions/cookie"
)

var jwtSecret = []byte("123")

func MainPage(c *gin.Context) {
	response := map[int]int{5: 5}
	c.JSON(200, response)
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Code     string `json:"code"`
}

var (
	redisClient *redis.Client
	once        sync.Once
	ctx         = context.Background()
)

func createRedisClient() *redis.Client {
	once.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     "redis:6379", // Укажите адрес вашего Redis сервера
			Password: "",           // Укажите пароль, если он есть
			DB:       0,            // Используйте базу данных по умолчанию
		})
	})
	return redisClient
}

func Sendmailfunc(user *User) error { //Если эта функция успешно возвратила nil. То код был отправлен на почту и код появился в редисе

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	confirmationCode := 100000 + r.Intn(899999)
	codeStr := strconv.Itoa(confirmationCode)

	m := gomail.NewMessage()

	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, continuing with environment variables")
	}

	m.SetHeader("From", os.Getenv("MAILCODESEND"))
	m.SetHeader("To", user.Email)
	fromEmail := os.Getenv("MAILCODESEND")
	if fromEmail == "" {
		fmt.Println("MAILCODESEND is not set")
	}
	m.SetHeader("Subject", "Подтверждение регистрации")
	m.SetBody("text/plain", "Ваш код подтверждения: "+codeStr)
	d := gomail.NewDialer("smtp.mail.ru", 465, os.Getenv("MAILCODESEND"), os.Getenv("SMTPPASSOWRD"))
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)

		return err
	}

	rdb := createRedisClient()
	pingRes, pingErr := rdb.Ping(ctx).Result()
	_ = pingRes
	if pingErr != nil {
		return pingErr
	}
	redisKey := user.Username + user.Email
	setErr := rdb.Set(ctx, redisKey, codeStr, 3*time.Minute).Err()
	if setErr != nil {
		return setErr
	}

	log.Printf("Значение для ключа %s установлено: %s\n", redisKey, codeStr)

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
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", user.Username).Scan(&exists1)

		if err != nil {
			log.Printf("Database error: %v", err)
			c.JSON(500, gin.H{"error": "Database error"})
			return
		}

		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", user.Email).Scan(&exists2)

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

		err = Sendmailfunc(&user)
		if err != nil {
			c.JSON(501, gin.H{
				"message":  "Данная почта уже используется", // тут добавить ошибку сервера и ошибку клиента
				"username": user.Username,
				"email":    user.Email,
				"error":    "Ошибка при отправке кода на почту",
			})
			return
		}

		c.JSON(200, gin.H{
			"message":  "Код подтверждения отправлен на почту",
			"username": user.Username,
			"email":    user.Email,
			"status":   "Код успешно отправлен на почту",
		})
	}
}

func Reg(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		rdb := createRedisClient()

		redisKey := user.Username + user.Email

		value, err := rdb.Get(ctx, redisKey).Result()
		if err == redis.Nil {
			log.Printf("Ключ %s не найден в Redis\n", redisKey)
			c.JSON(500, gin.H{"error": "Попробуйте ещё раз"})
			return
		} else if err != nil {
			log.Println("Ошибка при получении значения из Redis:", err)
			c.JSON(500, gin.H{"error": "Ошибка при получении данных"})
			return
		}
		if value == user.Code {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("Ошибка хеширования пароля:", err)
				c.JSON(500, gin.H{"error": "Ошибка при обработке пароля"})
				return
			}
			_, err = db.Exec("INSERT INTO users (username, password, email) VALUES ($1, $2, $3)", user.Username, string(hashedPassword), user.Email)
			if err != nil {
				log.Println("Ошибка при добавлении пользователя в базу данных:", err)
				c.JSON(500, gin.H{"error": "Ошибка при добавлении пользователя"})
				return
			}
			c.JSON(200, gin.H{"message": "Код подтверждения успешно подтверждён"})
		} else {
			c.JSON(401, gin.H{"error": "Неверный код подтверждения"})
		}
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
		err := db.QueryRow("SELECT password FROM users WHERE username = $1", user.Username).Scan(&storedPassword)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(401, gin.H{"error": "Неверно указан логин или пароль"})
				return
			}
			log.Printf("Database error: %v", err)
			c.JSON(500, gin.H{"error": "Database error"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)); err != nil {
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

		// cookie := http.Cookie{
		//     Name:     "token",
		//     Value:    tokenString,
		//     Path:     "/",
		//     MaxAge: 3000,
		//     HttpOnly: true,
		//     Secure:   false, // Установите true, если используете HTTPS
		//     SameSite: http.SameSiteStrictMode, // Разрешить доступ с другого источника
		// }
		//http.SetCookie(c.Writer, &cookie)
		c.SetCookie("token", tokenString, 3000, "/", "", false, true)
		c.JSON(200, gin.H{
			"message":  "Login successful",
			"username": user.Username,
			"token":    tokenString,
		})
	}
}

func GT(c *gin.Context) {
	c.JSON(200, gin.H{"number": 1})
}
