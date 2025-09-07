package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"interview/internal/models"
	"interview/internal/repository"
)

const (
	InterviewLinkDuration = 7 * 24 * time.Hour // Неделя
)

type InterviewService interface {
	CreateInterview(ctx context.Context, interview *models.Interview, baseURL string) error // Убрал scheduledTime
	StartInterview(ctx context.Context, token string) error
	FinishInterview(ctx context.Context, token string) error
	GetInterview(ctx context.Context, id string) (*models.Interview, error)
	GetInterviewByToken(ctx context.Context, token string) (*models.Interview, error)
	IsInterviewAccessible(ctx context.Context, token string) (bool, string, error)
	// Дополнительные полезные методы
	GetInterviewsByVacancy(ctx context.Context, vacancyID string) ([]*models.Interview, error)
}

type interviewService struct {
	repo repository.InterviewRepository
}

func NewInterviewService(repo repository.InterviewRepository) InterviewService {
	return &interviewService{repo: repo}
}

func (s *interviewService) CreateInterview(ctx context.Context, interview *models.Interview, baseURL string) error {
	if interview.VacancyID == "" {
		return errors.New("vacancy_id is required")
	}

	// Генерируем токен
	token := generateToken()
	fullURL := fmt.Sprintf("%s/api/interview/%s", baseURL, token)
	interview.URLToken = &fullURL
	interview.Status = "pending"

	// Ссылка действует неделю с момента создания
	now := time.Now()
	interview.ScheduledAt = &now // Доступна сразу

	expirationTime := now.Add(InterviewLinkDuration) // Неделя
	interview.ExpiresAt = &expirationTime

	return s.repo.Create(ctx, interview)
}

func (s *interviewService) IsInterviewAccessible(ctx context.Context, token string) (bool, string, error) {
	interview, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return false, "Interview not found", err
	}

	// Проверяем статус - если завершен, недоступен
	if interview.Status == "finished" {
		return false, "Interview has already been completed", nil
	}

	// Проверяем срок действия
	if interview.ExpiresAt != nil && time.Now().After(*interview.ExpiresAt) {
		return false, "Interview link has expired", nil
	}

	return true, "Interview is accessible", nil
}

func (s *interviewService) GetInterviewByToken(ctx context.Context, token string) (*models.Interview, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	interview, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get interview: %w", err)
	}

	return interview, nil
}

func (s *interviewService) StartInterview(ctx context.Context, token string) error {
	// Проверяем доступность
	accessible, message, err := s.IsInterviewAccessible(ctx, token)
	if err != nil {
		return err
	}
	if !accessible {
		return errors.New(message)
	}

	interview, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get interview: %w", err)
	}

	if interview.Status != "pending" {
		return fmt.Errorf("interview cannot be started from status: %s", interview.Status)
	}

	// Обновляем статус и фактическое время начала
	interview.Status = "started"
	now := time.Now()
	interview.StartedAt = &now

	return s.repo.Update(ctx, interview)
}

func (s *interviewService) FinishInterview(ctx context.Context, token string) error {
	interview, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get interview: %w", err)
	}

	if interview.Status != "started" {
		return fmt.Errorf("interview cannot be finished from status: %s", interview.Status)
	}

	// Обновляем статус на завершен
	interview.Status = "finished"

	return s.repo.Update(ctx, interview)
}

func (s *interviewService) GetInterview(ctx context.Context, id string) (*models.Interview, error) {
	if id == "" {
		return nil, errors.New("interview ID cannot be empty")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *interviewService) GetInterviewsByVacancy(ctx context.Context, vacancyID string) ([]*models.Interview, error) {
	if vacancyID == "" {
		return nil, errors.New("vacancy ID cannot be empty")
	}
	return s.repo.GetByVacancyID(ctx, vacancyID)
}

// Дополнительные utility методы

func (s *interviewService) GetTimeUntilExpiry(ctx context.Context, token string) (time.Duration, error) {
	interview, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return 0, err
	}

	if interview.ExpiresAt == nil {
		return 0, errors.New("interview has no expiry time")
	}

	now := time.Now()
	if now.After(*interview.ExpiresAt) {
		return 0, nil // Ссылка уже истекла
	}

	return interview.ExpiresAt.Sub(now), nil
}

func (s *interviewService) IsInterviewActive(ctx context.Context, token string) (bool, error) {
	interview, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return false, err
	}

	now := time.Now()

	// Проверяем срок действия
	if interview.ExpiresAt != nil && now.After(*interview.ExpiresAt) {
		return false, nil
	}

	// Проверяем статус
	return interview.Status == "pending" || interview.Status == "started", nil
}

func generateToken() string {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		panic(fmt.Sprintf("failed to generate token: %v", err))
	}
	return hex.EncodeToString(token)
}
