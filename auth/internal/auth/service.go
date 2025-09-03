package auth

import (
	"ai-hr-service/internal/models"
	"ai-hr-service/internal/utils"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	tokenService *TokenService
}

func NewService(repo Repository, tokenService *TokenService) Service {
	return &service{
		repo:         repo,
		tokenService: tokenService,
	}
}

func (s *service) Register(req RegisterRequest) (*LoginResponse, error) {
	// Проверяем существование пользователя по email
	if _, err := s.repo.GetUserByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Проверяем существование пользователя по username
	if _, err := s.repo.GetUserByUsername(req.Username); err == nil {
		return nil, errors.New("user with this username already exists")
	}

	// Хешируем пароль
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Создаем пользователя
	user := &models.User{
		Username:     req.Username,
		Surname:      req.Surname,
		Email:        req.Email,
		PasswordHash: hashedPassword,  // изменено на PasswordHash
		Role:         "hr_specialist", // обновлено на новую роль
		IsActive:     true,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Создаем токены
	accessToken, refreshToken, err := s.tokenService.CreateTokenPair(*user)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800, // 30 минут для JWT
		User:         *user,
	}, nil
}

func (s *service) Login(req LoginRequest) (*LoginResponse, error) {
	// Находим пользователя
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Проверяем активность пользователя
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Создаем токены
	accessToken, refreshToken, err := s.tokenService.CreateTokenPair(*user)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800, // 30 минут для JWT
		User:         *user,
	}, nil
}

func (s *service) RefreshTokens(req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	accessToken, refreshToken, err := s.tokenService.RefreshTokens(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800, // 30 минут для JWT
	}, nil
}

func (s *service) Logout(req LogoutRequest) error {
	return s.tokenService.RevokeRefreshToken(req.RefreshToken)
}

func (s *service) LogoutAll(userID uint) error {
	return s.tokenService.RevokeAllUserTokens(userID)
}

func (s *service) GetProfile(userID uint) (*models.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}
