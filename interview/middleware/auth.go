package middleware

import (
	"fmt"
	"net/http"
	"os" // 1. Импортируем пакет 'os' для доступа к переменным окружения
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

// 2. Используем функцию init() для инициализации секрета при старте приложения
func init() {
	// os.Getenv() читает переменную окружения
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Это fallback на случай, если переменная не установлена.
		// В проде здесь лучше вызывать панику: log.Fatal("JWT_SECRET is not set")
		fmt.Println("Warning: JWT_SECRET environment variable not set. Using default secret.")
		secret = "your-default-fallback-secret-for-dev-only"
	}
	jwtSecret = []byte(secret)
}

// Claims - это структура данных, которую вы зашиваете в токен.
// Она должна быть идентична той, что в auth-service.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
	jwt.RegisteredClaims
}

// TokenAuthMiddleware проверяет JWT токен и добавляет данные в контекст.
func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}
		tokenString := parts[1]

		claims := &Claims{}
		// При парсинге токена используется jwtSecret, который мы получили в init()
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		if !claims.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User account is deactivated"})
			return
		}

		// Успех! Сохраняем данные в контекст для следующих обработчиков.
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRoleMiddleware - это вспомогательный middleware для проверки роли.
func RequireRoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role not found in token"})
			return
		}

		if role.(string) != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		c.Next()
	}
}
