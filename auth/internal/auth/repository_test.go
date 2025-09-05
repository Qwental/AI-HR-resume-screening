package auth

import (
	"ai-hr-service/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// RepositoryTestSuite - test suite для интеграционного тестирования repository
type RepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo Repository
}

// SetupTest выполняется перед каждым тестом
func (suite *RepositoryTestSuite) SetupTest() {
	// Используем in-memory SQLite для тестов
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Миграции
	err = db.AutoMigrate(&models.User{}, &models.Token{})
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewRepository(db)
}

// TearDownTest выполняется после каждого теста
func (suite *RepositoryTestSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (suite *RepositoryTestSuite) createTestUser() *models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	return &models.User{
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

func (suite *RepositoryTestSuite) TestCreateUser() {
	// Arrange
	user := suite.createTestUser()

	// Act
	err := suite.repo.CreateUser(user)

	// Assert
	suite.NoError(err)
	suite.NotZero(user.ID)

	// Проверяем, что пользователь действительно создан в БД
	var createdUser models.User
	err = suite.db.First(&createdUser, user.ID).Error
	suite.NoError(err)
	suite.Equal(user.Username, createdUser.Username)
	suite.Equal(user.Email, createdUser.Email)
}

func (suite *RepositoryTestSuite) TestCreateUser_DuplicateEmail() {
	// Arrange
	user1 := suite.createTestUser()
	user2 := suite.createTestUser()
	user2.Username = "anotheruser"

	// Act
	err1 := suite.repo.CreateUser(user1)
	err2 := suite.repo.CreateUser(user2) // Same email

	// Assert
	suite.NoError(err1)
	suite.Error(err2) // Should fail due to unique constraint
}

func (suite *RepositoryTestSuite) TestCreateUser_DuplicateUsername() {
	// Arrange
	user1 := suite.createTestUser()
	user2 := suite.createTestUser()
	user2.Email = "another@example.com"

	// Act
	err1 := suite.repo.CreateUser(user1)
	err2 := suite.repo.CreateUser(user2) // Same username

	// Assert
	suite.NoError(err1)
	suite.Error(err2) // Should fail due to unique constraint
}

func (suite *RepositoryTestSuite) TestGetUserByEmail() {
	// Arrange
	originalUser := suite.createTestUser()
	err := suite.repo.CreateUser(originalUser)
	suite.Require().NoError(err)

	tests := []struct {
		name        string
		email       string
		expectError bool
		checkResult func(*models.User)
	}{
		{
			name:        "существующий активный пользователь",
			email:       "test@example.com",
			expectError: false,
			checkResult: func(user *models.User) {
				suite.Equal(originalUser.ID, user.ID)
				suite.Equal(originalUser.Username, user.Username)
				suite.Equal(originalUser.Email, user.Email)
				suite.True(user.IsActive)
			},
		},
		{
			name:        "несуществующий пользователь",
			email:       "nonexistent@example.com",
			expectError: true,
		},
		{
			name:        "неактивный пользователь",
			email:       "inactive@example.com",
			expectError: true,
		},
	}

	// Создаем неактивного пользователя
	inactiveUser := &models.User{
		Username:     "inactive",
		Surname:      "Inactive",
		Email:        "inactive@example.com",
		PasswordHash: "hash",
		Role:         "hr_specialist",
		IsActive:     false,
	}
	err = suite.db.Create(inactiveUser).Error
	suite.Require().NoError(err)

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Act
			user, err := suite.repo.GetUserByEmail(tt.email)

			// Assert
			if tt.expectError {
				suite.Error(err)
				suite.Nil(user)
			} else {
				suite.NoError(err)
				suite.NotNil(user)
				if tt.checkResult != nil {
					tt.checkResult(user)
				}
			}
		})
	}
}

func (suite *RepositoryTestSuite) TestGetUserByUsername() {
	// Arrange
	originalUser := suite.createTestUser()
	err := suite.repo.CreateUser(originalUser)
	suite.Require().NoError(err)

	tests := []struct {
		name        string
		username    string
		expectError bool
		checkResult func(*models.User)
	}{
		{
			name:        "существующий активный пользователь",
			username:    "testuser",
			expectError: false,
			checkResult: func(user *models.User) {
				suite.Equal(originalUser.ID, user.ID)
				suite.Equal(originalUser.Username, user.Username)
				suite.Equal(originalUser.Email, user.Email)
			},
		},
		{
			name:        "несуществующий пользователь",
			username:    "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Act
			user, err := suite.repo.GetUserByUsername(tt.username)

			// Assert
			if tt.expectError {
				suite.Error(err)
				suite.Nil(user)
			} else {
				suite.NoError(err)
				suite.NotNil(user)
				if tt.checkResult != nil {
					tt.checkResult(user)
				}
			}
		})
	}
}

func (suite *RepositoryTestSuite) TestGetUserByID() {
	// Arrange
	originalUser := suite.createTestUser()
	err := suite.repo.CreateUser(originalUser)
	suite.Require().NoError(err)

	tests := []struct {
		name        string
		userID      uint
		expectError bool
		checkResult func(*models.User)
	}{
		{
			name:        "существующий активный пользователь",
			userID:      originalUser.ID,
			expectError: false,
			checkResult: func(user *models.User) {
				suite.Equal(originalUser.ID, user.ID)
				suite.Equal(originalUser.Username, user.Username)
				suite.Equal(originalUser.Email, user.Email)
				suite.True(user.IsActive)
			},
		},
		{
			name:        "несуществующий пользователь",
			userID:      99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Act
			user, err := suite.repo.GetUserByID(tt.userID)

			// Assert
			if tt.expectError {
				suite.Error(err)
				suite.Nil(user)
			} else {
				suite.NoError(err)
				suite.NotNil(user)
				if tt.checkResult != nil {
					tt.checkResult(user)
				}
			}
		})
	}
}

// Benchmark тесты
//func (suite *RepositoryTestSuite) TestBenchmarkCreateUser() {
//	// Подготавливаем тестовые данные
//	users := make([]*models.User, 100)
//	for i := 0; i < 100; i++ {
//		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
//		users[i] = &models.User{
//			Username:     fmt.Sprintf("user%d", i),
//			Surname:      fmt.Sprintf("Surname%d", i),
//			Email:        fmt.Sprintf("user%d@example.com", i),
//			PasswordHash: string(hashedPassword),
//			Role:         "hr_specialist",
//			IsActive:     true,
//		}
//	}
//
//	// Создаем пользователей и измеряем производительность
//	start := time.Now()
//	for _, user := range users {
//		err := suite.repo.CreateUser(user)
//		suite.NoError(err)
//	}
//	duration := time.Since(start)
//
//	suite.T().Logf("Создание 100 пользователей заняло: %v", duration)
//	suite.T().Logf("Среднее время на пользователя: %v", duration/100)
//}

// Запуск test suite
//func TestRepositoryTestSuite(t *testing.T) {
//	suite.Run(t, new(RepositoryTestSuite))
//}

// Unit тесты без suite (для простых случаев)
func TestNewRepository(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	repo := NewRepository(db)
	assert.NotNil(t, repo)
	assert.IsType(t, &repository{}, repo)
}
