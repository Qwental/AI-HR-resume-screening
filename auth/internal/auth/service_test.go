package auth

import (
	"ai-hr-service/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MockRepository для тестирования service слоя
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetUserByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockTokenService для тестирования
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) CreateTokenPair(user models.User) (string, string, error) {
	args := m.Called(user)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockTokenService) RefreshTokens(refreshToken string) (string, string, error) {
	args := m.Called(refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockTokenService) RevokeRefreshToken(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

func (m *MockTokenService) RevokeAllUserTokens(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockTokenService) ValidateRefreshToken(refreshToken string) (*models.User, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// ДОБАВЛЕНО: CleanupExpiredTokens
func (m *MockTokenService) CleanupExpiredTokens() error {
	args := m.Called()
	return args.Error(0)
}

// ДОБАВЛЕНО: GetUserActiveTokensCount
func (m *MockTokenService) GetUserActiveTokensCount(userID uint) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

// Test fixtures
func createTestUser() *models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	return &models.User{
		ID:           1,
		Username:     "testuser",
		Surname:      "TestSurname",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Role:         "hr_specialist",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name           string
		req            RegisterRequest
		setupMocks     func(*MockRepository, *MockTokenService)
		expectedError  string
		expectedResult bool
	}{
		{
			name: "успешная регистрация",
			req: RegisterRequest{
				Username: "newuser",
				Surname:  "NewSurname",
				Email:    "new@example.com",
				Password: "password123",
			},
			setupMocks: func(mockRepo *MockRepository, mockTokenService *MockTokenService) {
				// Пользователь не существует
				mockRepo.On("GetUserByEmail", "new@example.com").Return(nil, gorm.ErrRecordNotFound)
				mockRepo.On("GetUserByUsername", "newuser").Return(nil, gorm.ErrRecordNotFound)

				// Успешное создание пользователя
				mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

				// Успешное создание токенов
				mockTokenService.On("CreateTokenPair", mock.AnythingOfType("models.User")).
					Return("access_token", "refresh_token", nil)
			},
			expectedResult: true,
		},
		{
			name: "пользователь уже существует по email",
			req: RegisterRequest{
				Username: "newuser",
				Surname:  "NewSurname",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(mockRepo *MockRepository, mockTokenService *MockTokenService) {
				existingUser := createTestUser()
				existingUser.Email = "existing@example.com"
				mockRepo.On("GetUserByEmail", "existing@example.com").Return(existingUser, nil)
			},
			expectedError: "user with this email already exists",
		},
		{
			name: "пользователь уже существует по username",
			req: RegisterRequest{
				Username: "existinguser",
				Surname:  "NewSurname",
				Email:    "new@example.com",
				Password: "password123",
			},
			setupMocks: func(mockRepo *MockRepository, mockTokenService *MockTokenService) {
				mockRepo.On("GetUserByEmail", "new@example.com").Return(nil, gorm.ErrRecordNotFound)

				existingUser := createTestUser()
				existingUser.Username = "existinguser"
				mockRepo.On("GetUserByUsername", "existinguser").Return(existingUser, nil)
			},
			expectedError: "user with this username already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockRepository)
			mockTokenService := new(MockTokenService)
			service := NewService(mockRepo, mockTokenService)

			tt.setupMocks(mockRepo, mockTokenService)

			// Act
			result, err := service.Register(tt.req)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "access_token", result.AccessToken)
				assert.Equal(t, "refresh_token", result.RefreshToken)
				assert.Equal(t, "Bearer", result.TokenType)
			}

			mockRepo.AssertExpectations(t)
			mockTokenService.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	tests := []struct {
		name           string
		req            LoginRequest
		setupMocks     func(*MockRepository, *MockTokenService)
		expectedError  string
		expectedResult bool
	}{
		{
			name: "успешный логин",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(mockRepo *MockRepository, mockTokenService *MockTokenService) {
				user := createTestUser()
				mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)
				mockTokenService.On("CreateTokenPair", *user).
					Return("access_token", "refresh_token", nil)
			},
			expectedResult: true,
		},
		{
			name: "неверные учетные данные - пользователь не найден",
			req: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupMocks: func(mockRepo *MockRepository, mockTokenService *MockTokenService) {
				mockRepo.On("GetUserByEmail", "nonexistent@example.com").
					Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "invalid credentials",
		},
		{
			name: "неверные учетные данные - неправильный пароль",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(mockRepo *MockRepository, mockTokenService *MockTokenService) {
				user := createTestUser()
				mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)
			},
			expectedError: "invalid credentials",
		},
		{
			name: "неактивный аккаунт",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(mockRepo *MockRepository, mockTokenService *MockTokenService) {
				user := createTestUser()
				user.IsActive = false
				mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)
			},
			expectedError: "account is deactivated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockRepository)
			mockTokenService := new(MockTokenService)
			service := NewService(mockRepo, mockTokenService)

			tt.setupMocks(mockRepo, mockTokenService)

			// Act
			result, err := service.Login(tt.req)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "access_token", result.AccessToken)
				assert.Equal(t, "refresh_token", result.RefreshToken)
			}

			mockRepo.AssertExpectations(t)
			mockTokenService.AssertExpectations(t)
		})
	}
}

func TestService_RefreshTokens(t *testing.T) {
	var tests []struct {
		name          string
		req           RefreshTokenRequest
		setupMocks    func(*MockTokenService)
		expectedError string
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockRepository)
			mockTokenService := new(MockTokenService)
			service := NewService(mockRepo, mockTokenService)

			tt.setupMocks(mockTokenService)

			// Act
			result, err := service.RefreshTokens(tt.req)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "new_access_token", result.AccessToken)
				assert.Equal(t, "new_refresh_token", result.RefreshToken)
			}

			mockTokenService.AssertExpectations(t)
		})
	}
}

func TestService_LogoutAll(t *testing.T) {
	mockTokenService := new(MockTokenService)
	service := NewService(new(MockRepository), mockTokenService)

	userID := uint(1)
	mockTokenService.On("RevokeAllUserTokens", userID).Return(nil)

	err := service.LogoutAll(userID)

	assert.NoError(t, err)
	mockTokenService.AssertExpectations(t)
}

func TestService_GetProfile(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint
		setupMocks    func(*MockRepository)
		expectedError string
	}{
		{
			name:   "успешное получение профиля",
			userID: 1,
			setupMocks: func(mockRepo *MockRepository) {
				user := createTestUser()
				mockRepo.On("GetUserByID", uint(1)).Return(user, nil)
			},
		},
		{
			name:   "пользователь не найден",
			userID: 999,
			setupMocks: func(mockRepo *MockRepository) {
				mockRepo.On("GetUserByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockRepository)
			service := NewService(mockRepo, new(MockTokenService))

			tt.setupMocks(mockRepo)

			// Act
			result, err := service.GetProfile(tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "testuser", result.Username)
				assert.Equal(t, "test@example.com", result.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
