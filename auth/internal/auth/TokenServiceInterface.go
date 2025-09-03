package auth

import (
	"ai-hr-service/internal/models"
)

// TokenServiceInterface - интерфейс для TokenService (для тестирования)
type TokenServiceInterface interface {
	CreateTokenPair(user models.User) (string, string, error)
	ValidateRefreshToken(refreshToken string) (*models.User, error)
	RefreshTokens(oldRefreshToken string) (string, string, error)
	RevokeRefreshToken(refreshToken string) error
	RevokeAllUserTokens(userID uint) error
	CleanupExpiredTokens() error
	GetUserActiveTokensCount(userID uint) (int64, error)
}

// Убедимся, что TokenService реализует интерфейс
var _ TokenServiceInterface = (*TokenService)(nil)
