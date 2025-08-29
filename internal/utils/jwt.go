package utils

import (
	"ai-hr-service/internal/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TODO
// Секретный ключ для подписи JWT токенов
// В продакшене брать из конфига, а не хардкодить
var jwtSecret = []byte("your-secret-key") // В продакшене брать из конфига

// Claims - структура данных которые храним в JWT токене

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken генерирует JWT токен для пользователя
func GenerateToken(user models.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 часа
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken проверяет и парсит JWT токен
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err // Токен невалидный или просрочен
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil // Все найс
	}

	return nil, errors.New("invalid token")
}
