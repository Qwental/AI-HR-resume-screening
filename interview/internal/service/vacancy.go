package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"interview/internal/models"
	"interview/internal/repository"
	"interview/internal/storage"
)

const MaxVacancyFileSize = 10 * 1024 * 1024 // 10MB

type VacancyService interface {
	CreateVacancy(ctx context.Context, vacancy *models.Vacancy, file io.Reader, filename string) error
	GetVacancy(ctx context.Context, id string) (*models.Vacancy, error)
	GetVacancyWithFileURL(ctx context.Context, id string) (*models.Vacancy, error)

	GetAllVacancies(ctx context.Context) ([]*models.Vacancy, error) // без StorageKey

	UpdateVacancy(ctx context.Context, vacancy *models.Vacancy) error
	UpdateVacancyWithFile(ctx context.Context, vacancy *models.Vacancy, file io.Reader, filename string) error
	DeleteVacancy(ctx context.Context, id string) error
}

type vacancyService struct {
	repo    repository.VacancyRepository
	storage storage.S3Storage
}

func NewVacancyService(repo repository.VacancyRepository, storage *storage.S3Storage) VacancyService {
	return &vacancyService{repo: repo, storage: *storage}
}

func (s *vacancyService) validateWeights(vacancy *models.Vacancy) error {
	total := vacancy.WeightSoft + vacancy.WeightHard + vacancy.WeightCase
	if total != 100 {
		return fmt.Errorf("сумма весов должна равняться 100, получено: %d", total)
	}
	return nil
}

func (s *vacancyService) validateFileType(filename string) error {
	allowedExts := []string{".pdf", ".doc", ".docx", ".txt"}
	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowed := range allowedExts {
		if ext == allowed {
			return nil
		}
	}
	return fmt.Errorf("неподдерживаемый тип файла: %s", ext)
}

func (s *vacancyService) CreateVacancy(ctx context.Context, vacancy *models.Vacancy, file io.Reader, filename string) error {
	if err := s.validateWeights(vacancy); err != nil {
		return err
	}

	if file == nil {
		return fmt.Errorf("файл обязателен для создания вакансии")
	}

	if err := s.validateFileType(filename); err != nil {
		return err
	}

	limitedReader := io.LimitReader(file, MaxVacancyFileSize)

	storageKey, err := s.storage.UploadVacancyFile(ctx, limitedReader, filename)
	if err != nil {
		return fmt.Errorf("ошибка загрузки файла: %w", err)
	}

	vacancy.StorageKey = storageKey
	vacancy.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, vacancy); err != nil {
		s.storage.DeleteFile(ctx, storageKey)
		return fmt.Errorf("ошибка создания вакансии: %w", err)
	}

	return nil
}

func (s *vacancyService) GetVacancy(ctx context.Context, id string) (*models.Vacancy, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *vacancyService) GetVacancyWithFileURL(ctx context.Context, id string) (*models.Vacancy, error) {
	vacancy, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if vacancy.StorageKey != "" {
		url, err := s.storage.GeneratePresignedURL(ctx, vacancy.StorageKey, time.Hour)
		if err == nil {
			vacancy.StorageKey = url
		}
	}

	return vacancy, nil
}

func (s *vacancyService) GetAllVacancies(ctx context.Context) ([]*models.Vacancy, error) {
	vacancies, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, vacancy := range vacancies {
		vacancy.StorageKey = ""
	}

	return vacancies, nil
}

func (s *vacancyService) UpdateVacancy(ctx context.Context, vacancy *models.Vacancy) error {
	existing, err := s.repo.GetByID(ctx, vacancy.ID)
	if err != nil {
		return fmt.Errorf("вакансия не найдена: %w", err)
	}

	if err := s.validateWeights(vacancy); err != nil {
		return err
	}

	vacancy.CreatedAt = existing.CreatedAt
	vacancy.StorageKey = existing.StorageKey
	now := time.Now()
	vacancy.UpdatedAt = &now

	return s.repo.Update(ctx, vacancy)
}

func (s *vacancyService) UpdateVacancyWithFile(ctx context.Context, vacancy *models.Vacancy, file io.Reader, filename string) error {
	existing, err := s.repo.GetByID(ctx, vacancy.ID)
	if err != nil {
		return fmt.Errorf("вакансия не найдена: %w", err)
	}

	if err := s.validateWeights(vacancy); err != nil {
		return err
	}

	if err := s.validateFileType(filename); err != nil {
		return err
	}

	limitedReader := io.LimitReader(file, MaxVacancyFileSize)

	storageKey, err := s.storage.UploadVacancyFile(ctx, limitedReader, filename)
	if err != nil {
		return fmt.Errorf("ошибка загрузки файла: %w", err)
	}

	oldStorageKey := existing.StorageKey
	vacancy.StorageKey = storageKey
	vacancy.CreatedAt = existing.CreatedAt
	now := time.Now()
	vacancy.UpdatedAt = &now

	if err := s.repo.Update(ctx, vacancy); err != nil {
		s.storage.DeleteFile(ctx, storageKey)
		return err
	}

	if oldStorageKey != "" {
		go func() {
			ctx := context.Background()
			_ = s.storage.DeleteFile(ctx, oldStorageKey)
		}()
	}

	return nil
}

func (s *vacancyService) DeleteVacancy(ctx context.Context, id string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("вакансия не найдена: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if existing.StorageKey != "" {
		go func() {
			ctx := context.Background()
			_ = s.storage.DeleteFile(ctx, existing.StorageKey)
		}()
	}

	return nil
}
