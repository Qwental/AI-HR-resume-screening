package auth

import (
	"ai-hr-service/internal/models"
	"ai-hr-service/internal/utils"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log/slog"
)

type Service interface {
	Register(req RegisterRequest) (*LoginResponse, error)
	Login(req LoginRequest) (*LoginResponse, error)
	RefreshTokens(req RefreshTokenRequest) (*RefreshTokenResponse, error)
	Logout(req LogoutRequest) error
	LogoutAll(userID uint) error
	GetProfile(userID uint) (*models.User, error)
}

type service struct {
	repo         Repository
	tokenService TokenServiceInterface
	logger       *slog.Logger
}

func NewService(repo Repository, tokenService TokenServiceInterface) Service {
	return &service{
		repo:         repo,
		tokenService: tokenService,
		logger:       slog.Default(),
	}
}

func (s *service) Register(req RegisterRequest) (*LoginResponse, error) {
	logger := s.logger.With("operation", "register", "username", req.Username, "email", req.Email)

	if err := req.Validate(); err != nil {
		logger.Warn("Validation failed", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Проверяем существование пользователя по email
	if _, err := s.repo.GetUserByEmail(req.Email); err == nil {
		logger.Warn("Registration attempt with existing email")
		return nil, errors.New("user with this email already exists")
	}

	// Проверяем существование пользователя по username
	if _, err := s.repo.GetUserByUsername(req.Username); err == nil {
		logger.Warn("Registration attempt with existing username")
		return nil, errors.New("user with this username already exists")
	}

	// Хешируем пароль
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		return nil, errors.New("failed to hash password")
	}

	// Создаем пользователя
	user := &models.User{
		Username:     req.Username,
		Surname:      req.Surname,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         "hr_specialist",
		IsActive:     true,
	}

	if err := s.repo.CreateUser(user); err != nil {
		logger.Error("Failed to create user", "error", err)
		return nil, errors.New("failed to create user")
	}

	// Создаем токены
	accessToken, refreshToken, err := s.tokenService.CreateTokenPair(*user)
	if err != nil {
		logger.Error("Failed to generate tokens", "error", err)
		return nil, errors.New("failed to generate tokens")
	}

	logger.Info("User registered successfully", "user_id", user.ID)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800,
		User:         *user,
	}, nil
}

func (s *service) Login(req LoginRequest) (*LoginResponse, error) {
	logger := s.logger.With("operation", "login", "email", req.Email)

	if err := req.Validate(); err != nil {
		logger.Warn("Validation failed", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Находим пользователя
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Login attempt with non-existent email")
			return nil, errors.New("invalid credentials")
		}
		logger.Error("Failed to get user", "error", err)
		return nil, err
	}

	// Проверяем активность пользователя
	if !user.IsActive {
		logger.Warn("Login attempt with inactive account", "user_id", user.ID)
		return nil, errors.New("account is deactivated")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		logger.Warn("Login attempt with invalid password", "user_id", user.ID)
		return nil, errors.New("invalid credentials")
	}

	// Создаем токены
	accessToken, refreshToken, err := s.tokenService.CreateTokenPair(*user)
	if err != nil {
		logger.Error("Failed to generate tokens", "error", err, "user_id", user.ID)
		return nil, errors.New("failed to generate tokens")
	}

	logger.Info("User logged in successfully", "user_id", user.ID)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800,
		User:         *user,
	}, nil
}

func (s *service) RefreshTokens(req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	logger := s.logger.With("operation", "refresh_tokens")

	if err := req.Validate(); err != nil {
		logger.Warn("Validation failed", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	accessToken, refreshToken, err := s.tokenService.RefreshTokens(req.RefreshToken)
	if err != nil {
		logger.Warn("Failed to refresh tokens", "error", err)
		return nil, err
	}

	logger.Info("Tokens refreshed successfully")

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800,
	}, nil
}

func (s *service) Logout(req LogoutRequest) error {
	logger := s.logger.With("operation", "logout")

	if err := req.Validate(); err != nil {
		logger.Warn("Validation failed", "error", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	err := s.tokenService.RevokeRefreshToken(req.RefreshToken)
	if err != nil {
		logger.Error("Failed to revoke refresh token", "error", err)
		return err
	}

	logger.Info("User logged out successfully")
	return nil
}

func (s *service) LogoutAll(userID uint) error {
	logger := s.logger.With("operation", "logout_all", "user_id", userID)

	err := s.tokenService.RevokeAllUserTokens(userID)
	if err != nil {
		logger.Error("Failed to revoke all user tokens", "error", err, "user_id", userID)
		return err
	}

	logger.Info("All user tokens revoked successfully", "user_id", userID)
	return nil
}

func (s *service) GetProfile(userID uint) (*models.User, error) {
	logger := s.logger.With("operation", "get_profile", "user_id", userID)

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Profile requested for non-existent user", "user_id", userID)
			return nil, errors.New("user not found")
		}
		logger.Error("Failed to get user profile", "error", err, "user_id", userID)
		return nil, err
	}

	return user, nil
}
