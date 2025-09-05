package auth

import (
	"ai-hr-service/internal/models"
	"ai-hr-service/internal/utils"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TokenService struct {
	db *gorm.DB
}

func NewTokenService(db *gorm.DB) *TokenService {
	return &TokenService{db: db}
}

// CreateTokenPair создает JWT access token и refresh token в БД
func (ts *TokenService) CreateTokenPair(user models.User) (string, string, error) {
	// 1. Создаем JWT access token (короткоживущий)
	jwtToken, err := utils.GenerateToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	// 2. Создаем refresh token (долгоживущий) и сохраняем в БД
	refreshToken, err := ts.createRefreshToken(user.ID, 7*24*time.Hour) // 7 дней
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return jwtToken, refreshToken, nil
}

// createRefreshToken создает refresh token и сохраняет в БД
func (ts *TokenService) createRefreshToken(userID uint, duration time.Duration) (string, error) {
	// Генерируем случайный токен
	rawToken, err := generateRandomToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Хэшируем для хранения
	tokenHash := hashToken(rawToken)

	// Сохраняем в БД
	token := &models.Token{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(duration),
		IsRevoked: false,
	}

	if err := ts.db.Create(token).Error; err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	return rawToken, nil
}

// ValidateRefreshToken проверяет refresh token и возвращает пользователя
func (ts *TokenService) ValidateRefreshToken(refreshToken string) (*models.User, error) {
	tokenHash := hashToken(refreshToken)

	var token models.Token
	err := ts.db.Preload("User").Where(
		"token_hash = ? AND expires_at > ? AND is_revoked = ?",
		tokenHash, time.Now(), false,
	).First(&token).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	// Проверяем активность пользователя
	if !token.User.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	return &token.User, nil
}

// RefreshTokens обновляет токены: отзывает старый refresh token и создает новую пару
func (ts *TokenService) RefreshTokens(oldRefreshToken string) (string, string, error) {
	// Проверяем старый токен
	user, err := ts.ValidateRefreshToken(oldRefreshToken)
	if err != nil {
		return "", "", err
	}

	// Отзываем старый refresh token
	if err := ts.RevokeRefreshToken(oldRefreshToken); err != nil {
		return "", "", fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Создаем новую пару токенов
	return ts.CreateTokenPair(*user)
}

// RevokeRefreshToken отзывает refresh token
func (ts *TokenService) RevokeRefreshToken(refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	now := time.Now()

	result := ts.db.Model(&models.Token{}).Where("token_hash = ?", tokenHash).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": &now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to revoke token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// RevokeAllUserTokens отзывает все refresh токены пользователя
func (ts *TokenService) RevokeAllUserTokens(userID uint) error {
	now := time.Now()

	return ts.db.Model(&models.Token{}).Where("user_id = ? AND is_revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": &now,
		}).Error
}

// CleanupExpiredTokens удаляет просроченные и отозванные токены (для cron)
func (ts *TokenService) CleanupExpiredTokens() error {
	return ts.db.Where("expires_at < ? OR is_revoked = ?", time.Now().AddDate(0, 0, -7), true).
		Delete(&models.Token{}).Error
}

// GetUserActiveTokensCount возвращает количество активных токенов пользователя
func (ts *TokenService) GetUserActiveTokensCount(userID uint) (int64, error) {
	var count int64
	err := ts.db.Model(&models.Token{}).Where(
		"user_id = ? AND expires_at > ? AND is_revoked = ?",
		userID, time.Now(), false,
	).Count(&count).Error
	return count, err
}

// Вспомогательные функции

// generateRandomToken генерирует криптографически стойкий случайный токен
func generateRandomToken() (string, error) {
	bytes := make([]byte, 32) // 256 бит
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken хэширует токен для безопасного хранения
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
