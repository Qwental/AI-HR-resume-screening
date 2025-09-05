package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/guylaor/goword"
	"interview/internal/storage"
)

type Job struct {
	Статус           string `json:"status"`
	Название         string `json:"title"`
	Регион           string `json:"region"`
	Город            string `json:"city"`
	Адрес            string `json:"address"`
	ТипТрудового     string `json:"employment_type"`
	ТипЗанятости     string `json:"work_type"`
	График           string `json:"schedule"`
	Доход            string `json:"income"`
	ОкладМакс        string `json:"salary_max"`
	ОкладМин         string `json:"salary_min"`
	ГодоваяПремия    string `json:"annual_bonus"`
	ТипПремирования  string `json:"bonus_type"`
	Обязанности      string `json:"responsibilities"`
	Требования       string `json:"requirements"`
	Образование      string `json:"education"`
	Опыт             string `json:"experience"`
	ЗнаниеПрограмм   string `json:"software_skills"`
	НавыкиКомпьютера string `json:"computer_skills"`
	ИностранныеЯзыки string `json:"languages"`
	УровеньЯзыка     string `json:"language_level"`
	Командировки     string `json:"business_trips"`
	ДопИнформация    string `json:"additional_info"`
}

func ExtractVacancyFromS3Key(ctx context.Context, storage *storage.S3Storage, storageKey string) (*Job, error) {
	reader, err := storage.DownloadFile(ctx, storageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}
	defer reader.Close()

	tempFile, err := os.CreateTemp("", "extraction_*.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	text, err := goword.ParseText(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to parse docx file: %w", err)
	}

	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	job := &Job{}

	keyMap := map[string]*string{
		"Статус":                       &job.Статус,
		"Название":                     &job.Название,
		"Регион":                       &job.Регион,
		"Город":                        &job.Город,
		"Адрес":                        &job.Адрес,
		"Тип трудового":                &job.ТипТрудового,
		"Тип занятости":                &job.ТипЗанятости,
		"Текст график работы":          &job.График,
		"Доход (руб/мес)":              &job.Доход,
		"Оклад макс. (руб/мес)":        &job.ОкладМакс,
		"Оклад мин. (руб/мес)":         &job.ОкладМин,
		"Годовая премия (%)":           &job.ГодоваяПремия,
		"Тип премирования. Описание":   &job.ТипПремирования,
		"Обязанности (для публикации)": &job.Обязанности,
		"Требования (для публикации)":  &job.Требования,
		"Уровень образования":          &job.Образование,
		"Требуемый опыт работы":        &job.Опыт,
		"Знание специальных программ":  &job.ЗнаниеПрограмм,
		"Навыки работы на компьютере":  &job.НавыкиКомпьютера,
		"Знание иностранных языков":    &job.ИностранныеЯзыки,
		"Уровень владения языка":       &job.УровеньЯзыка,
		"Наличие командировок":         &job.Командировки,
		"Дополнительная информация":    &job.ДопИнформация,
	}

	var currentKey *string
	var buffer []string

	for _, line := range lines {
		if line == "" {
			continue
		}

		if ptr, ok := keyMap[line]; ok {
			if currentKey != nil {
				*currentKey = strings.Join(buffer, " ")
			}
			currentKey = ptr
			buffer = nil
		} else if currentKey != nil {
			buffer = append(buffer, line)
		}
	}

	if currentKey != nil {
		*currentKey = strings.Join(buffer, " ")
	}

	return job, nil
}

func extractResumeFromDocx(file io.Reader) (string, error) {
	tempFile, err := os.CreateTemp("", "extract_*.docx")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	text, err := goword.ParseText(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to parse docx: %w", err)
	}

	return strings.TrimSpace(text), nil
}
