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

type ResumeService interface {
	CreateResume(ctx context.Context, resume *models.Resume, file io.Reader, filename string) error
	GetResume(ctx context.Context, id string) (*models.Resume, error)
	GetResumeWithFileURL(ctx context.Context, id string) (*models.Resume, error)
	GetResumesByVacancy(ctx context.Context, vacancyID string) ([]*models.Resume, error)
	DeleteResume(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateStatusAndResult(ctx context.Context, id, status string, result map[string]interface{}) error
	UpdateResult(ctx context.Context, id string, result map[string]interface{}) error
}

type resumeService struct {
	repo        repository.ResumeRepository
	storage     *storage.S3Storage
	vacancyRepo repository.VacancyRepository
}

func NewResumeService(repo repository.ResumeRepository, storage *storage.S3Storage, vacancyRepo repository.VacancyRepository) ResumeService {
	return &resumeService{repo: repo, storage: storage, vacancyRepo: vacancyRepo}
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

	//_ это vacancy
	_, err := s.vacancyRepo.GetByID(ctx, resume.VacancyID)
	if err != nil {
		return fmt.Errorf("vacancy not found: %w", err)
	}

	/* go func() {
		time.Sleep(1 * time.Second)

		fileData, err := io.ReadAll(io.LimitReader(file, MaxResumeFileSize))
		if err != nil {
			log.Printf("failed to read file: %w", err)
		}

		resumeText, err := extractResumeFromDocx(bytes.NewReader(fileData))
		if err != nil {
			log.Printf("failed to extract resume text: %w", err)
		}

		var vacancyData *Job
		if vacancy.TextJSONB != nil && len(vacancy.TextJSONB) > 0 {
			var vacancyDataMap map[string]interface{}
			if err := json.Unmarshal(vacancy.TextJSONB, &vacancyDataMap); err == nil {
				if structuredData, ok := vacancyDataMap["structured_data"]; ok {
					if jsonBytes, err := json.Marshal(structuredData); err == nil {
						vacancyData = &Job{}
						json.Unmarshal(jsonBytes, vacancyData)
						log.Printf("Using existing vacancy data for vacancy %s", vacancy.ID)
					}
				}
			}
		} else if vacancy.StorageKey != "" {
			log.Printf("Extracting vacancy data from file for vacancy %s", vacancy.ID)

			var err error
			vacancyData, err = ExtractVacancyFromS3Key(context.Background(), s.storage, vacancy.StorageKey)
			if err != nil {
				log.Printf("Failed to extract vacancy data for vacancy %s: %v", vacancy.ID, err)
			} else {
				vacancyDataMap := map[string]interface{}{
					"structured_data": vacancyData,
					"extracted_at":    time.Now(),
				}

				if jsonData, err := json.Marshal(vacancyDataMap); err == nil {
					vacancy.TextJSONB = datatypes.JSON(jsonData)

					if err := s.vacancyRepo.Update(context.Background(), vacancy); err != nil {
						log.Printf("Failed to update vacancy %s: %v", vacancy.ID, err)
					} else {
						log.Printf("Vacancy %s data extracted and saved", vacancy.ID)
					}
				}
			}
		} else {
			log.Printf("Vacancy %s has no file to extract data from", vacancy.ID)
		}

		if resumeText != "" || vacancyData != nil {
			messageData := map[string]interface{}{
				"resume_id":    resume.ID,
				"vacancy_id":   vacancy.ID,
				"resume_text":  resumeText,
				"vacancy_data": vacancyData,
				"timestamp":    time.Now(),
				"status":       "ready_for_analysis",
			}

			// TODO: Отправляем в message broker для AI анализа
			log.Printf("Sending to message broker: resume %s for vacancy %s", resume.ID, vacancy.ID)
			// s.messageBroker.Send("resume.analysis", messageData)

			// Для отладки - показываем что извлекли
			log.Printf("Extracted data summary:")
			log.Printf("- Resume text length: %d chars", len(resumeText))
			if vacancyData != nil {
				log.Printf("- Vacancy title: %s", vacancyData.Название)
				log.Printf("- Vacancy requirements: %.100s...", vacancyData.Требования)
			}
		} else {
			log.Printf("No data extracted for resume %s and vacancy %s", resume.ID, vacancy.ID)
		}

	}()

	*/

	limitedReader := io.LimitReader(file, MaxResumeFileSize)

	storageKey, err := s.storage.UploadResume(ctx, limitedReader, filename)
	if err != nil {
		return fmt.Errorf("file upload error: %w", err)
	}

	resume.StorageKey = storageKey
	resume.Status = "Прошел парсер"
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
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *resumeService) UpdateStatusAndResult(ctx context.Context, id, status string, result map[string]interface{}) error {
	return s.repo.UpdateStatusAndResult(ctx, id, status, result)
}

func (s *resumeService) UpdateResult(ctx context.Context, id string, result map[string]interface{}) error {

	return s.repo.UpdateResult(ctx, id, result)
}
