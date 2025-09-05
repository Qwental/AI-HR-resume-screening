package middleware

import (
	"ai-hr-service/internal/auth"
	"ai-hr-service/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthMiddleware struct {
	tokenService *auth.TokenService
}

func NewAuthMiddleware(db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: auth.NewTokenService(db),
	}
}

// TokenAuth проверяет JWT access token
func (am *AuthMiddleware) TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Проверяем формат: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		// Валидируем JWT токен
		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			status := http.StatusUnauthorized
			message := "Invalid or expired token"

			if err.Error() == "user account is deactivated" {
				message = "User account is deactivated"
			}

			utils.ErrorResponse(c, status, message)
			c.Abort()
			return
		}

		// Проверяем активность пользователя еще раз (на случай если токен старый)
		if !claims.IsActive {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User account is deactivated")
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("is_active", claims.IsActive)

		c.Next()
	}
}

// RequireRole проверяет роль пользователя (поддерживает несколько ролей)
func (am *AuthMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User role not found")
			c.Abort()
			return
		}

		roleStr := userRole.(string)

		// Проверяем, есть ли роль пользователя среди разрешенных
		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions")
		c.Abort()
	}
}

// AdminOnly - только для администраторов
func (am *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return am.RequireRole("admin")
}

// HROrAdmin - для HR-специалистов и администраторов
func (am *AuthMiddleware) HROrAdmin() gin.HandlerFunc {
	return am.RequireRole("hr_specialist", "admin")
}

// CandidateOnly - только для кандидатов
func (am *AuthMiddleware) CandidateOnly() gin.HandlerFunc {
	return am.RequireRole("candidate")
}

// OptionalAuth - опциональная авторизация (не обязательная)
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		// Пытаемся валидировать токен
		if claims, err := utils.ValidateToken(parts[1]); err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("is_active", claims.IsActive)
			c.Set("authenticated", true)
		} else {
			c.Set("authenticated", false)
		}

		c.Next()
	}
}
