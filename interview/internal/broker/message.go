package broker

import (
	"gorm.io/datatypes"
	"time"
)

// ResumeMessage описывает сообщение для обработки резюме
type ResumeMessage struct {
	ID         string         `json:"id"`
	VacancyID  string         `json:"vacancy_id"`
	CreatedAt  time.Time      `json:"created_at"`
	StorageKey string         `json:"storage_key"`
	Mail       string         `json:"mail"`
	Text       datatypes.JSON `json:"text_jsonb"` // Добавляем текст резюме
}
