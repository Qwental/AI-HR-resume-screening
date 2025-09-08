package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/datatypes"
	"interview/internal/broker"
	"interview/internal/models"
	"interview/internal/repository"
	"interview/internal/storage"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"
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
	publisher   broker.Publisher // будет  отпправлять в брокер

}

func NewResumeService(
	repo repository.ResumeRepository,
	storage *storage.S3Storage,
	vacancyRepo repository.VacancyRepository,
	publisher broker.Publisher,
) ResumeService {
	return &resumeService{
		repo:        repo,
		storage:     storage,
		vacancyRepo: vacancyRepo,
		publisher:   publisher,
	}
}

func (s *resumeService) validateFileType(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".docx", ".pdf", ".txt"}

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("unsupported file type: %s. Allowed types: %s",
		ext, strings.Join(allowedExts, ", "))
}

func (s *resumeService) CreateResume(ctx context.Context, resume *models.Resume, file io.Reader, filename string) error {
	if file == nil {
		return fmt.Errorf("file is required to create a resume")
	}

	if err := s.validateFileType(filename); err != nil {
		return err
	}

	//_ это vacancy
	vacancy, err := s.vacancyRepo.GetByID(ctx, resume.VacancyID)
	if err != nil {
		return fmt.Errorf("vacancy not found: %w", err)
	}

	// Читаем файл с ограничением размера
	fileData, err := io.ReadAll(io.LimitReader(file, MaxResumeFileSize))
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Дополнительная валидация содержимого файла
	if len(fileData) == 0 {
		return fmt.Errorf("file is empty")
	}

	fileType := detectFileType(fileData)
	if fileType == FileTypeUnknown {
		return fmt.Errorf("unsupported file format. Please upload DOCX, PDF or TXT file")
	}

	log.Printf("📄 Processing %s file: %s (%d bytes)", getFileTypeName(fileType), filename, len(fileData))

	// Загружаем файл в S3
	storageKey, err := s.storage.UploadResume(ctx, bytes.NewReader(fileData), filename)
	if err != nil {
		return fmt.Errorf("file upload error: %w", err)
	}

	resume.StorageKey = storageKey
	resume.Status = "Прошел парсер"
	resume.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, resume); err != nil {
		// Если создание записи в БД не удалось, удаляем файл из S3
		if deleteErr := s.storage.DeleteFile(ctx, storageKey); deleteErr != nil {
			log.Printf("❌ Failed to cleanup uploaded file after DB error: %v", deleteErr)
		}
		return fmt.Errorf("failed to create resume: %w", err)
	}

	log.Printf("✅ Resume created successfully: %s", resume.ID)

	// Асинхронно обрабатываем и отправляем в брокер
	go s.processResumeAsync(resume, fileData, vacancy)

	return nil
}

// ← Новый метод для асинхронной обработки
func (s *resumeService) processResumeAsync(resume *models.Resume, fileData []byte, vacancy *models.Vacancy) {
	ctx := context.Background()

	// Извлекаем текст из резюме с использованием универсальной функции
	resumeText, err := ExtractTextFromFile(fileData, resume.StorageKey)
	if err != nil {
		log.Printf("❌ Не удалось извлечь текст резюме для %s: %v", resume.ID, err)

		// Обновляем статус резюме на error
		if updateErr := s.repo.UpdateStatus(ctx, resume.ID, "error"); updateErr != nil {
			log.Printf("Failed to update resume status to error: %v", updateErr)
		}
		return
	}

	log.Printf("✅ Успешно извлечен текст резюме %s: %d символов", resume.ID, len(resumeText))

	// --- НОВОЕ: сохраняем ПЛЕЙН-ТЕКСТ в БД (колонка text)
	if err := s.repo.UpdateText(ctx, resume.ID, resumeText); err != nil {
		log.Printf("❌ Не удалось сохранить текст резюме в БД для %s: %v", resume.ID, err)
		// продолжаем процесс, чтобы не блокировать анализ
	} else {
		log.Printf("💾 Текст резюме %s сохранен в БД (колонка text)", resume.ID)
	}
	// --- КОНЕЦ НОВОГО БЛОКА ---

	// Подготавливаем текст вакансии
	var vacancyTextJSON datatypes.JSON
	var vacancyData *Job

	if vacancy.TextJSONB != nil && len(vacancy.TextJSONB) > 0 {
		// Используем уже существующие данные вакансии
		vacancyTextJSON = vacancy.TextJSONB

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
		// Извлекаем данные из файла вакансии
		log.Printf("Extracting vacancy data from file for vacancy %s", vacancy.ID)

		vacancyData, err = ExtractVacancyFromS3Key(ctx, s.storage, vacancy.StorageKey)
		if err != nil {
			log.Printf("Failed to extract vacancy data for vacancy %s: %v", vacancy.ID, err)
		} else {
			// Сохраняем извлеченные данные вакансии
			vacancyDataMap := map[string]interface{}{
				"structured_data": vacancyData,
				"extracted_at":    time.Now(),
			}

			if jsonData, err := json.Marshal(vacancyDataMap); err == nil {
				vacancy.TextJSONB = datatypes.JSON(jsonData)
				vacancyTextJSON = vacancy.TextJSONB

				if err := s.vacancyRepo.Update(ctx, vacancy); err != nil {
					log.Printf("Failed to update vacancy %s: %v", vacancy.ID, err)
				} else {
					log.Printf("Vacancy %s data extracted and saved", vacancy.ID)
				}
			}
		}
	}

	// Если не удалось получить данные вакансии, создаем базовый JSON
	if vacancyTextJSON == nil {
		basicVacancyData := map[string]interface{}{
			"title":       vacancy.Title,
			"description": vacancy.Description,
			"created_at":  vacancy.CreatedAt,
		}
		if jsonData, err := json.Marshal(basicVacancyData); err == nil {
			vacancyTextJSON = datatypes.JSON(jsonData)
		}
	}

	// Подготавливаем текст резюме в JSON формате
	resumeTextJSON := datatypes.JSON("{}")
	if resumeText != "" {
		resumeDataMap := map[string]interface{}{
			"text":         resumeText,
			"extracted_at": time.Now(),
			"file_name":    resume.StorageKey,
			"file_type":    string(getFileTypeName(detectFileType(fileData))),
			"size_bytes":   len(fileData),
		}
		if jsonData, err := json.Marshal(resumeDataMap); err == nil {
			resumeTextJSON = datatypes.JSON(jsonData)
		}
	}

	// 🚀 Создаем и отправляем сообщение в брокер
	message := broker.ResumeMessage{
		ID:          resume.ID,
		VacancyID:   resume.VacancyID,
		TextResume:  resumeTextJSON,     // JSON с текстом резюме
		TextVacancy: vacancyTextJSON,    // JSON с текстом вакансии
		WeightSoft:  vacancy.WeightSoft, // Вес soft skills
		WeightHard:  vacancy.WeightHard, // Вес hard skills
		WeightCase:  vacancy.WeightCase, // Вес кейсов/опыта
	}

	// Отправляем сообщение
	if err := s.publisher.PublishResumeMessage(ctx, message); err != nil {
		log.Printf("Failed to publish resume message for %s: %v", resume.ID, err)

		// Обновляем статус резюме на error
		if updateErr := s.repo.UpdateStatus(ctx, resume.ID, "error"); updateErr != nil {
			log.Printf("Failed to update resume status to error: %v", updateErr)
		}
	} else {
		log.Printf("✅ Successfully published resume message: resume %s for vacancy %s", resume.ID, resume.VacancyID)

		//// Обновляем статус резюме на processing
		//if updateErr := s.repo.UpdateStatus(ctx, resume.ID, ResumeStatusProcessing); updateErr != nil {
		//	log.Printf("Failed to update resume status to processing: %v", updateErr)
		//}
	}

	// Логируем для отладки
	if resumeText != "" {
		log.Printf("📄 Resume text extracted: %d characters", len(resumeText))
	}
	if vacancyData != nil {
		log.Printf("📋 Vacancy data: %s", vacancyData.Название)
	}
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
