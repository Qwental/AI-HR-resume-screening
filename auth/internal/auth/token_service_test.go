package auth

import (
	"ai-hr-service/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TokenServiceTestSuite для интеграционного тестирования TokenService
type TokenServiceTestSuite struct {
	suite.Suite
	db           *gorm.DB
	tokenService *TokenService
}

func (suite *TokenServiceTestSuite) SetupTest() {
	// Используем in-memory SQLite для тестов
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Миграции
	err = db.AutoMigrate(&models.User{}, &models.Token{})
	suite.Require().NoError(err)

	suite.db = db
	suite.tokenService = NewTokenService(db)
}

func (suite *TokenServiceTestSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (suite *TokenServiceTestSuite) createTestUser() *models.User {
	user := &models.User{
		Username:     "testuser",
		Surname:      "TestSurname",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         "hr_specialist",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := suite.db.Create(user).Error
	suite.Require().NoError(err)
	return user
}

func (suite *TokenServiceTestSuite) TestCreateTokenPair() {
	// Arrange
	user := suite.createTestUser()

	// Act
	accessToken, refreshToken, err := suite.tokenService.CreateTokenPair(*user)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(accessToken)
	suite.NotEmpty(refreshToken)

	// Проверяем, что refresh token сохранен в БД
	var tokenCount int64
	err = suite.db.Model(&models.Token{}).Where("user_id = ?", user.ID).Count(&tokenCount).Error
	suite.NoError(err)
	suite.Equal(int64(1), tokenCount)
}

func (suite *TokenServiceTestSuite) TestCreateTokenPair_MultipleTokens() {
	// Arrange
	user := suite.createTestUser()

	// Act - создаем несколько пар токенов
	_, refreshToken1, err1 := suite.tokenService.CreateTokenPair(*user)
	_, refreshToken2, err2 := suite.tokenService.CreateTokenPair(*user)

	// Assert
	suite.NoError(err1)
	suite.NoError(err2)
	suite.NotEqual(refreshToken1, refreshToken2)

	// Проверяем количество токенов в БД
	var tokenCount int64
	err := suite.db.Model(&models.Token{}).Where("user_id = ?", user.ID).Count(&tokenCount).Error
	suite.NoError(err)
	suite.Equal(int64(2), tokenCount)
}

func (suite *TokenServiceTestSuite) TestValidateRefreshToken() {
	// Arrange
	user := suite.createTestUser()
	_, refreshToken, err := suite.tokenService.CreateTokenPair(*user)
	suite.Require().NoError(err)

	tests := []struct {
		name        string
		token       string
		expectError bool
		checkResult func(*models.User)
	}{
		{
			name:        "валидный токен",
			token:       refreshToken,
			expectError: false,
			checkResult: func(validatedUser *models.User) {
				suite.Equal(user.ID, validatedUser.ID)
				suite.Equal(user.Email, validatedUser.Email)
				suite.True(validatedUser.IsActive)
			},
		},
		{
			name:        "недействительный токен",
			token:       "invalid_token",
			expectError: true,
		},
		{
			name:        "пустой токен",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Act
			validatedUser, err := suite.tokenService.ValidateRefreshToken(tt.token)

			// Assert
			if tt.expectError {
				suite.Error(err)
				suite.Nil(validatedUser)
			} else {
				suite.NoError(err)
				suite.NotNil(validatedUser)
				if tt.checkResult != nil {
					tt.checkResult(validatedUser)
				}
			}
		})
	}
}

func (suite *TokenServiceTestSuite) TestValidateRefreshToken_InactiveUser() {
	// Arrange
	user := suite.createTestUser()
	_, refreshToken, err := suite.tokenService.CreateTokenPair(*user)
	suite.Require().NoError(err)

	// Деактивируем пользователя
	user.IsActive = false
	err = suite.db.Save(user).Error
	suite.Require().NoError(err)

	// Act
	validatedUser, err := suite.tokenService.ValidateRefreshToken(refreshToken)

	// Assert
	suite.Error(err)
	suite.Nil(validatedUser)
	suite.Contains(err.Error(), "user account is deactivated")
}

func (suite *TokenServiceTestSuite) TestRefreshTokens() {
	// Arrange
	user := suite.createTestUser()
	_, oldRefreshToken, err := suite.tokenService.CreateTokenPair(*user)
	suite.Require().NoError(err)

	// Act
	newAccessToken, newRefreshToken, err := suite.tokenService.RefreshTokens(oldRefreshToken)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(newAccessToken)
	suite.NotEmpty(newRefreshToken)
	suite.NotEqual(oldRefreshToken, newRefreshToken)

	// Проверяем, что старый токен отозван
	var token models.Token
	err = suite.db.Where("token_hash = ?", hashToken(oldRefreshToken)).First(&token).Error
	suite.NoError(err)
	suite.True(token.IsRevoked)
	suite.NotNil(token.RevokedAt)

	// Проверяем, что новый токен валиден
	validatedUser, err := suite.tokenService.ValidateRefreshToken(newRefreshToken)
	suite.NoError(err)
	suite.Equal(user.ID, validatedUser.ID)
}

func (suite *TokenServiceTestSuite) TestRefreshTokens_InvalidToken() {
	// Act
	accessToken, refreshToken, err := suite.tokenService.RefreshTokens("invalid_token")

	// Assert
	suite.Error(err)
	suite.Empty(accessToken)
	suite.Empty(refreshToken)
}

func (suite *TokenServiceTestSuite) TestRevokeRefreshToken() {
	// Arrange
	user := suite.createTestUser()
	_, refreshToken, err := suite.tokenService.CreateTokenPair(*user)
	suite.Require().NoError(err)

	// Act
	err = suite.tokenService.RevokeRefreshToken(refreshToken)

	// Assert
	suite.NoError(err)

	// Проверяем, что токен отозван
	var token models.Token
	err = suite.db.Where("token_hash = ?", hashToken(refreshToken)).First(&token).Error
	suite.NoError(err)
	suite.True(token.IsRevoked)
	suite.NotNil(token.RevokedAt)

	// Проверяем, что отозванный токен больше не валиден
	_, err = suite.tokenService.ValidateRefreshToken(refreshToken)
	suite.Error(err)
}

func (suite *TokenServiceTestSuite) TestRevokeRefreshToken_NonexistentToken() {
	// Act
	err := suite.tokenService.RevokeRefreshToken("nonexistent_token")

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "token not found")
}

func (suite *TokenServiceTestSuite) TestRevokeAllUserTokens() {
	// Arrange
	user := suite.createTestUser()

	// Создаем несколько токенов для пользователя
	_, refreshToken1, _ := suite.tokenService.CreateTokenPair(*user)
	_, refreshToken2, _ := suite.tokenService.CreateTokenPair(*user)
	_, refreshToken3, _ := suite.tokenService.CreateTokenPair(*user)

	// Act
	err := suite.tokenService.RevokeAllUserTokens(user.ID)

	// Assert
	suite.NoError(err)

	// Проверяем, что все токены отозваны
	var activeTokenCount int64
	err = suite.db.Model(&models.Token{}).Where(
		"user_id = ? AND expires_at > ? AND is_revoked = ?",
		user.ID, time.Now(), false,
	).Count(&activeTokenCount).Error
	suite.NoError(err)
	suite.Equal(int64(0), activeTokenCount)

	// Проверяем, что все токены больше не валидны
	_, err1 := suite.tokenService.ValidateRefreshToken(refreshToken1)
	_, err2 := suite.tokenService.ValidateRefreshToken(refreshToken2)
	_, err3 := suite.tokenService.ValidateRefreshToken(refreshToken3)

	suite.Error(err1)
	suite.Error(err2)
	suite.Error(err3)
}

func (suite *TokenServiceTestSuite) TestGetUserActiveTokensCount() {
	// Arrange
	user := suite.createTestUser()

	// Создаем токены
	suite.tokenService.CreateTokenPair(*user)
	suite.tokenService.CreateTokenPair(*user)

	// Один токен отзываем
	_, refreshToken3, _ := suite.tokenService.CreateTokenPair(*user)
	suite.tokenService.RevokeRefreshToken(refreshToken3)

	// Act
	count, err := suite.tokenService.GetUserActiveTokensCount(user.ID)

	// Assert
	suite.NoError(err)
	suite.Equal(int64(2), count) // 2 активных токена
}

func (suite *TokenServiceTestSuite) TestCleanupExpiredTokens() {
	// Arrange
	user := suite.createTestUser()

	// Создаем токен с коротким временем жизни
	token := &models.Token{
		UserID:    user.ID,
		TokenHash: hashToken("test_token"),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Просроченный
		IsRevoked: false,
	}
	err := suite.db.Create(token).Error
	suite.Require().NoError(err)

	// Создаем отозванный токен
	revokedToken := &models.Token{
		UserID:    user.ID,
		TokenHash: hashToken("revoked_token"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		IsRevoked: true,
		RevokedAt: &time.Time{},
	}
	*revokedToken.RevokedAt = time.Now().Add(-8 * 24 * time.Hour) // 8 дней назад
	err = suite.db.Create(revokedToken).Error
	suite.Require().NoError(err)

	// Создаем валидный токен
	_, _, err = suite.tokenService.CreateTokenPair(*user)
	suite.Require().NoError(err)

	// Act
	err = suite.tokenService.CleanupExpiredTokens()

	// Assert
	suite.NoError(err)

	// Проверяем, что просроченные и старые отозванные токены удалены
	var remainingCount int64
	err = suite.db.Model(&models.Token{}).Where("user_id = ?", user.ID).Count(&remainingCount).Error
	suite.NoError(err)
	suite.Equal(int64(1), remainingCount) // Остался только валидный токен
}

func (suite *TokenServiceTestSuite) TestTokenSecurity() {
	// Проверяем, что токены генерируются случайно и не повторяются
	user := suite.createTestUser()

	tokens := make(map[string]bool)

	for i := 0; i < 100; i++ {
		_, refreshToken, err := suite.tokenService.CreateTokenPair(*user)
		suite.NoError(err)
		suite.False(tokens[refreshToken], "Токен должен быть уникальным")
		tokens[refreshToken] = true
	}
}

// Unit тесты для вспомогательных функций
func TestGenerateRandomToken(t *testing.T) {
	token1, err1 := generateRandomToken()
	token2, err2 := generateRandomToken()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2)
	assert.Len(t, token1, 64) // 32 bytes = 64 hex chars
}

func TestHashToken(t *testing.T) {
	token := "test_token_123"

	hash1 := hashToken(token)
	hash2 := hashToken(token)
	hash3 := hashToken("different_token")

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2)    // Одинаковые токены дают одинаковые хеши
	assert.NotEqual(t, hash1, hash3) // Разные токены дают разные хеши
	assert.Len(t, hash1, 64)         // SHA256 = 64 hex chars
}
