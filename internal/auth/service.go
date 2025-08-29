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
	GetProfile(userID uint) (*models.User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
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
		Username: req.Username,
		Surname:  req.Surname,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "hr",
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Генерируем токен
	token, err := utils.GenerateToken(*user)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &LoginResponse{
		Token: token,
		User:  *user,
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

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Генерируем токен
	token, err := utils.GenerateToken(*user)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &LoginResponse{
		Token: token,
		User:  *user,
	}, nil
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
