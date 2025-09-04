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

const MaxResumeFileSize = 20 * 1024 * 1024 // 20MB

const (
	ResumeStatusPending    = "pending"
	ResumeStatusProcessing = "processing"
	ResumeStatusApproved   = "approved"
	ResumeStatusRejected   = "rejected"
	ResumeStatusError      = "error"
)

type ResumeService interface {
	CreateResume(ctx context.Context, resume *models.Resume, file io.Reader, filename string) error
	GetResume(ctx context.Context, id string) (*models.Resume, error)
	GetResumeWithFileURL(ctx context.Context, id string) (*models.Resume, error)
	GetResumesByVacancy(ctx context.Context, vacancyID string) ([]*models.Resume, error)
	DeleteResume(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateStatusAndResult(ctx context.Context, id, status string, result map[string]interface{}) error
}

type resumeService struct {
	repo    repository.ResumeRepository
	storage *storage.S3Storage
}

func NewResumeService(repo repository.ResumeRepository, storage *storage.S3Storage) ResumeService {
	return &resumeService{repo: repo, storage: storage}
}

func (s *resumeService) validateFileType(filename string) error {
	allowedExts := []string{".pdf", ".doc", ".docx", ".txt", ".rtf"}
	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowed := range allowedExts {
		if ext == allowed {
			return nil
		}
	}
	return fmt.Errorf("unsupported file type: %s", ext)
}

func (s *resumeService) CreateResume(ctx context.Context, resume *models.Resume, file io.Reader, filename string) error {
	if file == nil {
		return fmt.Errorf("file is required to create a resume")
	}

	if err := s.validateFileType(filename); err != nil {
		return err
	}

	limitedReader := io.LimitReader(file, MaxResumeFileSize)

	//тут типо достается текст резюме и ваки, ес че json
	//щас делаю

	//тут типо отправляется в брокер
	// TODO: Send to message broker for processing
	// go s.sendToProcessing(ctx, resume)

	//далее там уже если проходит, то сохраняем

	storageKey, err := s.storage.UploadResume(ctx, limitedReader, filename)
	if err != nil {
		return fmt.Errorf("file upload error: %w", err)
	}

	resume.StorageKey = storageKey
	resume.Status = ResumeStatusPending
	resume.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, resume); err != nil {
		s.storage.DeleteFile(ctx, storageKey)
		return fmt.Errorf("failed to create resume: %w", err)
	}

	return nil
}

func (s *resumeService) GetResume(ctx context.Context, id string) (*models.Resume, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *resumeService) GetResumeWithFileURL(ctx context.Context, id string) (*models.Resume, error) {
	resume, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if resume.StorageKey != "" {
		url, err := s.storage.GeneratePresignedURL(ctx, resume.StorageKey, time.Hour)
		if err == nil {
			resume.FileURL = url
		}
	}

	return resume, nil
}

func (s *resumeService) GetResumesByVacancy(ctx context.Context, vacancyID string) ([]*models.Resume, error) {
	resumes, err := s.repo.GetByVacancy(ctx, vacancyID)
	if err != nil {
		return nil, err
	}

	return resumes, nil
}

func (s *resumeService) DeleteResume(ctx context.Context, id string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("resume not found: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete resume: %w", err)
	}

	if existing.StorageKey != "" {
		go func() {
			ctx := context.Background()
			if err := s.storage.DeleteFile(ctx, existing.StorageKey); err != nil {
				fmt.Printf("Warning: failed to delete file %s: %v\n", existing.StorageKey, err)
			}
		}()
	}

	return nil
}

func (s *resumeService) UpdateStatus(ctx context.Context, id, status string) error {
	validStatuses := []string{ResumeStatusPending, ResumeStatusProcessing, ResumeStatusApproved, ResumeStatusRejected, ResumeStatusError}

	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid status: %s", status)
	}

	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *resumeService) UpdateStatusAndResult(ctx context.Context, id, status string, result map[string]interface{}) error {
	if err := s.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	return s.repo.UpdateStatusAndResult(ctx, id, status, result)
}
