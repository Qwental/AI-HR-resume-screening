package utils

import (
	"ai-hr-service/internal/config"
	"ai-hr-service/internal/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims - структура данных которые храним в JWT токене
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
	jwt.RegisteredClaims
}

var jwtConfig *config.JWTConfig

// InitJWT инициализирует JWT конфигурацию
func InitJWT(cfg *config.JWTConfig) {
	jwtConfig = cfg
}

// GenerateToken генерирует JWT access token для пользователя
func GenerateToken(user models.User) (string, error) {
	if jwtConfig == nil {
		return "", errors.New("JWT config not initialized")
	}

	// Парсим TTL
	duration, err := time.ParseDuration(jwtConfig.AccessTokenTTL)
	if err != nil {
		duration = 30 * time.Minute // fallback
	}

	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		IsActive: user.IsActive,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Email,
			Issuer:    "ai-hr-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.Secret))
}

// ValidateToken проверяет и парсит JWT токен
func ValidateToken(tokenString string) (*Claims, error) {
	if jwtConfig == nil {
		return nil, errors.New("JWT config not initialized")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Дополнительная проверка активности пользователя
		if !claims.IsActive {
			return nil, errors.New("user account is deactivated")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetTokenRemainingTime возвращает оставшееся время жизни токена
func GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	if claims.ExpiresAt == nil {
		return 0, errors.New("token has no expiration time")
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0, errors.New("token has expired")
	}

	return remaining, nil
}
