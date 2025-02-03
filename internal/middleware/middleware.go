package middleware

import (
    _ "github.com/jackc/pgx/v4/stdlib"
    "github.com/gin-gonic/gin"
    "net/http"
    "github.com/golang-jwt/jwt/v4"
	"strings"
)

var jwtSecret = []byte("123")

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
            c.Abort()
            return
        }

        // Проверка формата заголовка
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
            c.Abort()
            return
        }

        tokenString := parts[1]

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()	
            return
        }

        c.Next()
    }
}

// func CORSMiddleware() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         // Устанавливаем необходимые заголовки CORS
//         c.Header("Access-Control-Allow-Origin", "*") // Разрешаем все источники
//         c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Разрешаем методы
//         c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept") // Разрешаем заголовки

//         // Обрабатываем preflight запросы
//         if c.Request.Method == http.MethodOptions {
//             c.AbortWithStatus(http.StatusNoContent) // Возвращаем статус 204 No Content
//             return
//         }
        
//         c.Next() // Продолжаем обработку запроса
//     }
// }