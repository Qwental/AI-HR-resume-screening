package middleware

import (
	"ai-hr-service/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет JWT токен
// Если токена нет или он говно - отправляем нахуй
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// значит токена нет
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header required")

			c.Abort() // Останавливаем выполнение дальнейших middleware
			return
		}

		// Проверяем формат: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		// Валидируем токен на просрочку
		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			// Токен просрочен
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
			c.Abort()
			return
		}

		// Если все ок - сохраняем данные пользователя в контексте
		// Чтобы другие middleware и хендлеры могли их использовать
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole проверяет роль пользователя
// Если роль не подходит - отправляем нахуй

func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем роль из контекста (должна была быть установлена AuthMiddleware)

		userRole, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User role not found")
			c.Abort()
			return
		}
		// Проверяем что роль подходит

		if userRole != requiredRole {
			utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
